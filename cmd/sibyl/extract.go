package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/williamfzc/sibyl2"
	"github.com/williamfzc/sibyl2/pkg/core"
	"github.com/williamfzc/sibyl2/pkg/extractor"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var userSrc string
var userLangType string
var userExtractType string
var userOutputFile string

var allowExtractType = []string{
	extractor.TypeExtractSymbol,
	extractor.TypeExtractFunction,
	extractor.TypeExtractCall,
}

func NewExtractCmd() *cobra.Command {
	extractCmd := &cobra.Command{
		Use:    "extract",
		Short:  "test",
		Long:   `test`,
		Hidden: false,
		Run: func(cmd *cobra.Command, args []string) {
			langType := core.LangTypeValueOf(userLangType)
			if langType == core.LangUnknown {
				panic(errors.New("unknown lang type: " + userLangType))
			}

			if !slices.Contains(allowExtractType, userExtractType) {
				panic(errors.New("non-allow extract type: " + userExtractType))
			}

			if userOutputFile == "" {
				userOutputFile = fmt.Sprintf("sibyl-%s-%s-%d.json", userExtractType, langType, time.Now().Unix())
			}

			config := &sibyl2.ExtractConfig{
				LangType:    langType,
				ExtractType: userExtractType,
			}
			results, err := sibyl2.Extract(userSrc, config)
			if err != nil {
				panic(err)
			}
			output, err := json.MarshalIndent(&results, "", "  ")
			if err != nil {
				panic(err)
			}
			err = os.WriteFile(userOutputFile, output, 0644)
			if err != nil {
				panic(err)
			}
		},
	}

	extractCmd.PersistentFlags().StringVar(&userSrc, "src", ".", "src dir path")

	extractCmd.PersistentFlags().StringVar(&userLangType, "lang", "", "lang type of your source code")
	err := extractCmd.MarkPersistentFlagRequired("lang")
	if err != nil {
		panic(err)
	}

	extractCmd.PersistentFlags().StringVar(&userExtractType, "type", "", "what kind of data you want")
	err = extractCmd.MarkPersistentFlagRequired("type")
	if err != nil {
		panic(err)
	}

	extractCmd.PersistentFlags().StringVar(&userOutputFile, "output", "", "output file")
	if err != nil {
		panic(err)
	}

	return extractCmd
}

func init() {
	extractCmd := NewExtractCmd()
	rootCmd.AddCommand(extractCmd)
}
