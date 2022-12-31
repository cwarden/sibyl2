package upload

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

const (
	configPath = "."
	configFile = "sibyl-upload-config.json"
	configType = "json"
)

type SrcConfigPart struct {
	Src          string `mapstructure:"src"`
	Lang         string `mapstructure:"lang"`
	WithCtx      bool   `mapstructure:"withCtx"`
	IncludeRegex string `mapstructure:"includeRegex"`
	ExcludeRegex string `mapstructure:"excludeRegex"`
}

type ServerConfigPart struct {
	Url   string `mapstructure:"url"`
	Batch int    `mapstructure:"batch"`
	Dry   bool   `mapstructure:"dry"`
}

type uploadConfig struct {
	*SrcConfigPart    `mapstructure:"src"`
	*ServerConfigPart `mapstructure:"server"`
}

func (config *uploadConfig) ToMap() (map[string]any, error) {
	var m map[string]interface{}
	err := mapstructure.Decode(config, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (config *uploadConfig) ToJson() ([]byte, error) {
	toMap, err := config.ToMap()
	if err != nil {
		return nil, err
	}
	return json.Marshal(toMap)
}

func defaultConfig() *uploadConfig {
	return &uploadConfig{
		&SrcConfigPart{
			Src:          ".",
			Lang:         "",
			WithCtx:      false,
			IncludeRegex: "",
			ExcludeRegex: "",
		},
		&ServerConfigPart{
			Url:   "http://127.0.0.1:9876",
			Batch: 50,
			Dry:   false,
		},
	}
}
