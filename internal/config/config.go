package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
	"regexp"
	"time"
)

// Validation rules at https://pkg.go.dev/github.com/go-playground/validator/v10
type (
	VigilisConfig struct {
		Storage *Storage `yaml:"storage" validate:"required"`

		Cameras []*Camera `yaml:"cameras" validate:"required,gt=0,unique=Id,dive"`

		Recorder *Recorder `yaml:"recorder" validate:"omitempty"`
	}

	Storage struct {
		Path          string `yaml:"path" validate:"required,dirpath,gte=1"`
		RetentionDays int    `yaml:"retention_days" validate:"required,number,gte=1"`
	}

	Camera struct {
		Id        string `yaml:"id" validate:"required,slug,gte=1,lte=20"`
		Name      string `yaml:"name" validate:"required,gte=1,lte=30"`
		StreamUrl string `yaml:"stream_url" validate:"required,url,gte=8"`
	}

	Recorder struct {
		FfmpegPath string `yaml:"ffmpeg_path" validate:"filepath"`
	}
)

var Vigilis = VigilisConfig{
	Recorder: &Recorder{
		FfmpegPath: "ffmpeg",
	},
}

func Parse(data []byte) error {
	// Setup the data validator
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Register the custom validator for slugs
	err := validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		regex := regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
		return regex.MatchString(fl.Field().String())
	})
	if err != nil {
		return err
	}

	// Try to decode the config
	err = yaml.UnmarshalWithOptions(data, &Vigilis, yaml.Strict())
	if err != nil {
		return err
	}

	// Try to validate the config
	err = validate.Struct(&Vigilis)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) RetentionDaysDuration() time.Duration {
	return time.Hour * 24 * time.Duration(s.RetentionDays)
}
