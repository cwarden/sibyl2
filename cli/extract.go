package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"os"
	"sibyl2/pkg/core"
	"sibyl2/pkg/extractor"
	"sibyl2/pkg/model"
	"time"
)

var userSrc string
var userLangType string
var userExtractType string
var userOutputFile string

var allowExtractType = []string{extractor.TypeExtractSymbol, extractor.TypeExtractFunction}

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
				userOutputFile = fmt.Sprintf("sibyl-%s-%d.json", langType, time.Now().Unix())
			}

			// call
			runner := &core.Runner{}
			fileUnits, err := runner.File2Units(userSrc, langType)
			if err != nil {
				panic(err)
			}

			langExtractor := extractor.GetExtractor(langType)
			var results []*model.FileResult
			for _, eachFileUnit := range fileUnits {
				switch userExtractType {
				case extractor.TypeExtractSymbol:
					symbols, err := langExtractor.ExtractSymbols(eachFileUnit.Units)
					if err != nil {
						panic(err)
					}
					var retUnits []model.DataType
					for _, each := range symbols {
						retUnits = append(retUnits, model.DataType(each))
					}
					fileResult := &model.FileResult{
						Path:     eachFileUnit.Path,
						Language: eachFileUnit.Language,
						Type:     userExtractType,
						Units:    retUnits,
					}
					results = append(results, fileResult)
				case extractor.TypeExtractFunction:
					functions, err := langExtractor.ExtractFunctions(eachFileUnit.Units)
					if err != nil {
						panic(err)
					}
					var retUnits []model.DataType
					for _, each := range functions {
						retUnits = append(retUnits, model.DataType(each))
					}
					fileResult := &model.FileResult{
						Path:     eachFileUnit.Path,
						Language: eachFileUnit.Language,
						Type:     userExtractType,
						Units:    retUnits,
					}
					results = append(results, fileResult)
				}
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
