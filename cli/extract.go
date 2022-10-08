package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"os"
	"sibyl2/pkg"
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
			langType := model.LangTypeValueOf(userLangType)
			if langType == model.LangUnknown {
				panic(errors.New("unknown lang type: " + userLangType))
			}

			if !slices.Contains(allowExtractType, userExtractType) {
				panic(errors.New("non-allow extract type: " + userExtractType))
			}

			if userOutputFile == "" {
				userOutputFile = fmt.Sprintf("sibyl-%s-%d.json", langType, time.Now().Unix())
			}

			results, err := pkg.SibylApi.Extract(userSrc, langType, userExtractType)
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
