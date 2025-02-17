package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
	"vigilis/internal/logger"
)

type (
	VigilisConfig struct {
		Storage *Storage `yaml:"storage" validate:"required"`

		Cameras *[]Camera `yaml:"cameras" validate:"required,gte=1"`
	}

	Storage struct {
		Path string `yaml:"path" validate:"required,dirpath,gte=1"`
	}

	Camera struct {
		Id        string `yaml:"id" validate:"required,alphanum,gte=1,lte=20"`
		Name      string `yaml:"name" validate:"required,alphaunicode,gte=1,lte=30"`
		StreamUrl string `yaml:"stream_url" validate:"required,url,gte=8"`
	}
)

var Vigilis VigilisConfig

func Parse(data []byte) {
	// Setup the data validator
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Try to decode and validate the config
	err := yaml.UnmarshalWithOptions(data, &Vigilis, yaml.Validator(validate), yaml.Strict())
	if err != nil {
		logger.Fatal("Error parsing the config.\n%v", err)
	}
}
