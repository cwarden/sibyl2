package sibyl2

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/williamfzc/sibyl2/pkg/core"
	"reflect"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/williamfzc/sibyl2/pkg/extractor"
)

const (
	TemplateMergeFuncPrefix = "MERGE " +
		"(:Repo {id: $repo_id})" +
		"-[:INCLUDE]->" +
		"(rev:Rev {hash: $rev_hash})"
	TemplateMergeFuncFile = "MERGE (rev)" +
		"-[:INCLUDE]->" +
		"(f:File {path: $file_path, lang: $file_lang})"
	TemplateMergeFuncSelf = "MERGE (f)" +
		"-[:INCLUDE]->" +
		"(:Func {" +
		"name: $func_name, " +
		"receiver: $func_receiver, " +
		"parameters: $func_parameters, " +
		"returns: $func_returns, " +
		"span: $func_span, " +
		"extras: $func_extras," +
		"signature: $func_signature })"

	TemplateMatchFuncFull = "MATCH " +
		"(repo:Repo {id: $repo_id})" +
		"-[:INCLUDE]->" +
		"(rev:Rev {hash: $rev_hash})" +
		"-[:INCLUDE]->" +
		"(%s:File {path: %s, lang: $file_lang})" +
		"-[:INCLUDE]->" +
		"(func%d:Func {signature: $func%d_signature})"

	TemplateMergeLinkInclude       = "MERGE (%s)-[:INCLUDE]->(%s)"
	TemplateMergeLinkFileReference = "MERGE (%s)-[:FILE_REFERENCE]->(%s)"
	TemplateMergeLinkFuncReference = "MERGE (%s)-[:FUNC_REFERENCE]->(%s)"
)

type Neo4jDriver struct {
	neo4j.DriverWithContext
}

func (d *Neo4jDriver) UploadFileResult(wc *WorkspaceConfig, f *extractor.FunctionFileResult, ctx context.Context) error {
	// session is cheap to create
	session := d.DriverWithContext.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	_, err := session.ExecuteWrite(ctx, createFunctionFileTransaction(wc, f, ctx))
	if err != nil {
		return err
	}
	return nil
}

func (d *Neo4jDriver) UploadFuncContext(wc *WorkspaceConfig, f *FunctionContext, ctx context.Context) error {
	// session is cheap to create
	session := d.DriverWithContext.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	_, err := session.ExecuteWrite(ctx, createFuncGraphTransaction(wc, f, ctx))
	if err != nil {
		return err
	}
	return nil
}

func (d *Neo4jDriver) QueryFiles(wc *WorkspaceConfig, ctx context.Context) ([]string, error) {
	session := d.DriverWithContext.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	ret, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `MATCH (:Repo {id: $repoId})-[:INCLUDE]->(:Rev {hash: $revHash})-[:INCLUDE]->(f:File) RETURN f.path`
		results, err := tx.Run(ctx, query, map[string]any{
			"repoId":  wc.RepoId,
			"revHash": wc.RevHash,
		})
		if err != nil {
			return nil, err
		}
		nodes, err := results.Collect(ctx)
		if err != nil {
			return nil, err
		}
		var ret []string
		for _, each := range nodes {
			ret = append(ret, each.Values[0].(string))
		}
		return ret, nil
	})
	if err != nil {
		return nil, err
	}
	return ret.([]string), nil
}

func (d *Neo4jDriver) QueryFunctions(wc *WorkspaceConfig, path string, ctx context.Context) ([]*FunctionWithPath, error) {
	session := d.DriverWithContext.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	ret, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `MATCH (:Repo {id: $repoId})-[:INCLUDE]->(:Rev {hash: $revHash})-[:INCLUDE]->(file:File {path: $path})-[:INCLUDE]->(f) RETURN f, file`
		results, err := tx.Run(ctx, query, map[string]any{
			"repoId":  wc.RepoId,
			"revHash": wc.RevHash,
			"path":    path,
		})
		if err != nil {
			return nil, err
		}
		nodes, err := results.Collect(ctx)
		if err != nil {
			return nil, err
		}
		var ret []*FunctionWithPath
		for _, each := range nodes {
			rawMap := each.Values[0].(dbtype.Node).Props
			file := each.Values[1].(dbtype.Node).Props

			// special handlers for neo4j :)
			var params []any
			var returns []any
			var span map[string]any
			_ = json.Unmarshal([]byte(rawMap["parameters"].(string)), &params)
			_ = json.Unmarshal([]byte(rawMap["returns"].(string)), &returns)
			_ = json.Unmarshal([]byte(rawMap["span"].(string)), &span)
			rawMap["parameters"] = params
			rawMap["returns"] = returns
			rawMap["span"] = span

			f, err := extractor.FromMap(rawMap)
			if err != nil {
				return nil, err
			}
			fwp := &FunctionWithPath{
				Function: f,
				Path:     file["lang"].(string),
				Language: core.LangType(file["path"].(string)),
			}
			ret = append(ret, fwp)
		}
		return ret, nil
	})
	if err != nil {
		return nil, err
	}
	return ret.([]*FunctionWithPath), nil
}

func (d *Neo4jDriver) QueryFunctionWithSignature(wc *WorkspaceConfig, signature string, ctx context.Context) (*FunctionWithPath, error) {
	session := d.DriverWithContext.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	ret, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `MATCH (:Repo {id: $repoId})-[:INCLUDE]->(:Rev {hash: $revHash})-[:INCLUDE]->(file:File)-[:INCLUDE]->(f:Func {signature: $signature}) RETURN f, file`
		results, err := tx.Run(ctx, query, map[string]any{
			"repoId":    wc.RepoId,
			"revHash":   wc.RevHash,
			"signature": signature,
		})
		if err != nil {
			return nil, err
		}
		nodes, err := results.Collect(ctx)
		if err != nil {
			return nil, err
		}
		var ret []*FunctionWithPath
		for _, each := range nodes {
			rawMap := each.Values[0].(dbtype.Node).Props
			file := each.Values[1].(dbtype.Node).Props

			// special handlers for neo4j :)
			var params []any
			var returns []any
			var span map[string]any
			_ = json.Unmarshal([]byte(rawMap["parameters"].(string)), &params)
			_ = json.Unmarshal([]byte(rawMap["returns"].(string)), &returns)
			_ = json.Unmarshal([]byte(rawMap["span"].(string)), &span)
			rawMap["parameters"] = params
			rawMap["returns"] = returns
			rawMap["span"] = span

			f, err := extractor.FromMap(rawMap)
			if err != nil {
				return nil, err
			}
			fwp := &FunctionWithPath{
				Function: f,
				Path:     file["lang"].(string),
				Language: core.LangType(file["path"].(string)),
			}
			ret = append(ret, fwp)
		}
		// normally, it will contain only one element
		if len(ret) == 0 {
			return nil, nil
		}
		return ret[0], nil
	})
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return nil, nil
	}
	return ret.(*FunctionWithPath), nil
}

func createFunctionFileTransaction(wc *WorkspaceConfig, f *extractor.FunctionFileResult, ctx context.Context) neo4j.ManagedTransactionWork {
	return func(tx neo4j.ManagedTransaction) (any, error) {
		for _, each := range f.Units {
			funcMap, _ := each.ToMap()

			// neo4j can not handle nested properties and null
			// awful support and I really hate this shit
			for k, v := range funcMap {
				kind := reflect.ValueOf(v).Kind()
				// nested map and slice -> string
				if kind == reflect.Map || kind == reflect.Slice {
					pureContent, _ := json.Marshal(v)
					funcMap[k] = string(pureContent)
					continue
				}
				// nil -> empty string
				if v == nil {
					funcMap[k] = ""
				}
			}

			merged := []string{
				TemplateMergeFuncPrefix,
				TemplateMergeFuncFile,
				TemplateMergeFuncSelf,
			}

			// todo: merge these run together
			_, err := tx.Run(ctx, strings.Join(merged, " "), map[string]any{
				"repo_id":         wc.RepoId,
				"rev_hash":        wc.RevHash,
				"file_path":       f.Path,
				"file_lang":       f.Language,
				"func_name":       each.Name,
				"func_receiver":   funcMap["receiver"],
				"func_parameters": funcMap["parameters"],
				"func_returns":    funcMap["returns"],
				"func_span":       funcMap["span"],
				"func_extras":     funcMap["extras"],
				"func_signature":  each.GetSignature(),
			})
			_, err = tx.Run(ctx, fmt.Sprintf(TemplateMergeLinkInclude, "file", "func"), nil)
			if err != nil {
				return nil, err
			}
		}

		// temp ignore return values
		return nil, nil
	}
}

func createFuncGraphTransaction(wc *WorkspaceConfig, f *FunctionContext, ctx context.Context) neo4j.ManagedTransactionWork {
	return func(tx neo4j.ManagedTransaction) (any, error) {
		for i, each := range f.Calls {
			id := i + 1
			eachFuncName := fmt.Sprintf("func%d", id)
			eachMerged := []string{
				fmt.Sprintf(TemplateMatchFuncFull, "srcfile", "$srcPath", 0, 0),
				fmt.Sprintf(TemplateMatchFuncFull, "targetfile", "$targetPath", id, id),
			}
			if each.Path != f.Path {
				eachMerged = append(eachMerged, fmt.Sprintf(TemplateMergeLinkFileReference, "srcfile", "targetfile"))
			}
			eachMerged = append(eachMerged,
				// create link
				fmt.Sprintf(TemplateMergeLinkFuncReference, "func0", eachFuncName),
				"RETURN *")

			// todo: merge these run together
			_, err := tx.Run(ctx, strings.Join(eachMerged, " "), map[string]any{
				"repo_id":                           wc.RepoId,
				"rev_hash":                          wc.RevHash,
				"srcPath":                           f.Path,
				"targetPath":                        each.Path,
				"file_lang":                         each.Language,
				"func0_signature":                   f.GetSignature(),
				fmt.Sprintf("func%d_signature", id): each.GetSignature(),
			})
			if err != nil {
				return nil, err
			}
		}

		// temp ignore return values
		return nil, nil
	}
}
