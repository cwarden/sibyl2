package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/opensibyl/sibyl2"
	"github.com/opensibyl/sibyl2/pkg/server/binding"
	"github.com/opensibyl/sibyl2/pkg/server/object"
	"github.com/opensibyl/sibyl2/pkg/server/queue"
)

var sharedContext context.Context
var sharedDriver binding.Driver
var sharedQueue queue.Queue

// @Summary func query
// @Param   repo  query string true  "repo"
// @Param   rev   query string true  "rev"
// @Param   file  query string true  "file"
// @Param   lines query string false "specific lines"
// @Produce json
// @Success 200 {array} object.FunctionWithSignature
// @Router  /api/v1/func [get]
// @Tags    BasicQuery
func HandleFunctionsQuery(c *gin.Context) {
	repo := c.Query("repo")
	rev := c.Query("rev")
	file := c.Query("file")
	lines := c.Query("lines")

	ret, err := handleFunctionQuery(repo, rev, file, lines)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, ret)
}

func handleFunctionQuery(repo string, rev string, file string, lines string) ([]*object.FunctionWithSignature, error) {
	wc := &object.WorkspaceConfig{
		RepoId:  repo,
		RevHash: rev,
	}
	if err := wc.Verify(); err != nil {
		return nil, err
	}
	var functions []*object.FunctionWithSignature
	var err error
	if lines == "" {
		functions, err = sharedDriver.ReadFunctions(wc, file, sharedContext)
	} else {
		linesStrList := strings.Split(lines, ",")
		var lineNums = make([]int, 0, len(linesStrList))
		for _, each := range linesStrList {
			num, err := strconv.Atoi(each)
			if err != nil {
				return nil, fmt.Errorf("strconv convert error: %w", err)
			}
			lineNums = append(lineNums, num)
		}
		functions, err = sharedDriver.ReadFunctionsWithLines(wc, file, lineNums, sharedContext)
	}
	if err != nil {
		return nil, fmt.Errorf("read function error: %w", err)
	}
	return functions, nil
}

// @Summary func ctx query
// @Param   repo  query string true  "repo"
// @Param   rev   query string true  "rev"
// @Param   file  query string true  "file"
// @Param   lines query string false "specific lines"
// @Produce json
// @Success 200 {array} object.FunctionContextSlimWithSignature
// @Router  /api/v1/funcctx [get]
// @Tags    BasicQuery
func HandleFunctionContextsQuery(c *gin.Context) {
	repo := c.Query("repo")
	rev := c.Query("rev")
	file := c.Query("file")
	lines := c.Query("lines")

	wc := &object.WorkspaceConfig{
		RepoId:  repo,
		RevHash: rev,
	}

	ret, err := handleFunctionQuery(repo, rev, file, lines)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	ctxs := make([]*object.FunctionContextSlimWithSignature, 0)
	for _, each := range ret {
		funcCtx, err := sharedDriver.ReadFunctionContextWithSignature(wc, each.Signature, sharedContext)
		if err != nil {
			c.JSON(http.StatusBadRequest, fmt.Errorf("err when read with sign: %w", err))
			return
		}
		wrapper := &object.FunctionContextSlimWithSignature{
			FunctionContextSlim: funcCtx,
			Signature:           funcCtx.GetSignature(),
		}
		ctxs = append(ctxs, wrapper)
	}
	c.JSON(http.StatusOK, ctxs)
}

// @Summary class query
// @Param   repo query string true "repo"
// @Param   rev  query string true "rev"
// @Param   file query string true "file"
// @Produce json
// @Success 200 {array} sibyl2.ClazzWithPath
// @Router  /api/v1/clazz [get]
// @Tags    BasicQuery
func HandleClazzesQuery(c *gin.Context) {
	repo := c.Query("repo")
	rev := c.Query("rev")
	file := c.Query("file")

	ret, err := handleClazzQuery(repo, rev, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, ret)
}

func handleClazzQuery(repo string, rev string, file string) ([]*sibyl2.ClazzWithPath, error) {
	wc := &object.WorkspaceConfig{
		RepoId:  repo,
		RevHash: rev,
	}
	if err := wc.Verify(); err != nil {
		return nil, err
	}
	var functions []*sibyl2.ClazzWithPath
	var err error
	functions, err = sharedDriver.ReadClasses(wc, file, sharedContext)
	if err != nil {
		return nil, err
	}
	return functions, nil
}
