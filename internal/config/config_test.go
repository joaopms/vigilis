package config

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"slices"
	"strings"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	cases := []struct {
		Name             string
		ExpectedError    string
		MustNotHaveError string
		Data             string
	}{
		// Bad data
		{
			Name:          "empty",
			ExpectedError: "",
			Data:          "",
		},
		{
			Name:          "not-yaml",
			ExpectedError: "[1:1] string was used where mapping is expected\n>  1 | dummydummy\n       ^\n",
			Data:          "dummydummy",
		},

		// Storage
		{
			Name:          "invalid-storage-no-values",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag",
			Data: `---
storage:
`,
		},
		{
			Name:          "invalid-storage-empty-path",
			ExpectedError: "Key: 'VigilisConfig.Storage.Path' Error:Field validation for 'Path' failed on the 'required' tag",
			Data: `---
storage:
  path:
`,
		},
		{
			Name:          "invalid-storage-valid-path-without-slash",
			ExpectedError: "Key: 'VigilisConfig.Storage.RetentionDays' Error:Field validation for 'RetentionDays' failed on the 'required' tag",
			Data: `---
storage:
  path: /tmp/vigilis
`,
		},
		{
			Name:          "invalid-storage-valid-path-with-slash",
			ExpectedError: "Key: 'VigilisConfig.Storage.RetentionDays' Error:Field validation for 'RetentionDays' failed on the 'required' tag",
			Data: `---
storage:
  path: /tmp/vigilis/
`,
		},
		{
			Name:          "invalid-storage-empty-retention-days",
			ExpectedError: "Key: 'VigilisConfig.Storage.RetentionDays' Error:Field validation for 'RetentionDays' failed on the 'required' tag",
			Data: `---
storage:
  retention_days:
`,
		},
		{
			Name:          "invalid-storage-zero-retention-days",
			ExpectedError: "Key: 'VigilisConfig.Storage.RetentionDays' Error:Field validation for 'RetentionDays' failed on the 'required' tag",
			Data: `---
storage:
  retention_days: 0
`,
		},
		{
			Name:          "invalid-storage-invalid-floating-retention-days",
			ExpectedError: "Key: 'VigilisConfig.Storage.RetentionDays' Error:Field validation for 'RetentionDays' failed on the 'required' tag",
			Data: `---
storage:
  path: /tmp/vigilis
  retention_days: 0.5
`,
		},
		{
			Name:             "invalid-storage-valid-integer-retention-days",
			MustNotHaveError: "VigilisConfig.Storage.RetentionDays",
			Data: `---
storage:
  path: /tmp/vigilis
  retention_days: 1
`,
		},

		// Cameras
		{
			Name:          "invalid-cameras-no-values",
			ExpectedError: "Key: 'VigilisConfig.Cameras' Error:Field validation for 'Cameras' failed on the 'required' tag",
			Data: `---
cameras:
`,
		},
		{
			Name:          "invalid-cameras-empty-id",
			ExpectedError: "Key: 'VigilisConfig.Cameras[0].Id' Error:Field validation for 'Id' failed on the 'required' tag",
			Data: `---
cameras:
  - id:
`,
		},
		{
			Name:          "invalid-cameras-empty-name",
			ExpectedError: "Key: 'VigilisConfig.Cameras[0].Name' Error:Field validation for 'Name' failed on the 'required' tag",
			Data: `---
cameras:
  - name:
`,
		},
		{
			Name:          "invalid-cameras-empty-stream-url",
			ExpectedError: "Key: 'VigilisConfig.Cameras[0].StreamUrl' Error:Field validation for 'StreamUrl' failed on the 'required' tag",
			Data: `---
cameras:
  - stream_url:
`,
		},
		{
			Name:          "invalid-cameras-short-stream-url",
			ExpectedError: "Key: 'VigilisConfig.Cameras[0].StreamUrl' Error:Field validation for 'StreamUrl' failed on the 'gte' tag",
			Data: `---
cameras:
  - stream_url: a://a
`,
		},
		{
			Name:          "invalid-cameras-invalid-stream-url",
			ExpectedError: "Key: 'VigilisConfig.Cameras[0].StreamUrl' Error:Field validation for 'StreamUrl' failed on the 'url' tag",
			Data: `---
cameras:
  - stream_url: 12345678
`,
		},
		{
			Name:             "invalid-cameras-valid-stream-url",
			MustNotHaveError: "VigilisConfig.Cameras.StreamUrl",
			Data: `---
cameras:
  - stream_url: rtsp://a
`,
		},
		{
			Name:          "invalid-cameras-long-id",
			ExpectedError: "Key: 'VigilisConfig.Cameras[0].Id' Error:Field validation for 'Id' failed on the 'lte' tag",
			Data: `---
cameras:
  - id: 12345678901234567890A
`,
		},
		{
			Name:          "invalid-cameras-long-name",
			ExpectedError: "Key: 'VigilisConfig.Cameras[0].Name' Error:Field validation for 'Name' failed on the 'lte' tag",
			Data: `---
cameras:
  - name: 123456789012345678901234567890A
`,
		},
		{
			Name:          "invalid-cameras-id-with-space",
			ExpectedError: "Key: 'VigilisConfig.Cameras[0].Id' Error:Field validation for 'Id' failed on the 'slug' tag",
			Data: `---
cameras:
  - id: "a b"
    name: "a b"
`,
		},
		{
			Name:          "invalid-cameras-id-not-alphanum",
			ExpectedError: "Key: 'VigilisConfig.Cameras[0].Id' Error:Field validation for 'Id' failed on the 'slug' tag",
			Data: `---
cameras:
  - id: a! 
    name: "a b!"
`,
		},
		{
			Name:          "invalid-cameras-repeated-ids",
			ExpectedError: "Key: 'VigilisConfig.Cameras' Error:Field validation for 'Cameras' failed on the 'unique' tag",
			Data: `---
cameras:
  - id: a
    name: A
  - id: a
    name: B
`,
		},
		{
			Name:             "valid-two-cameras",
			MustNotHaveError: "VigilisConfig.Cameras",
			Data: `---
cameras:
  - id: a
    name: A
    stream_url: rtsp://a
  - id: b
    name: B
    stream_url: rtsp://b
`,
		},
		{
			Name:             "valid-cameras-with-underscore-or-dash-in-id",
			MustNotHaveError: "VigilisConfig.Cameras",
			Data: `---
cameras:
  - id: a_b
    name: "A B"
    stream_url: rtsp://a-b
  - id: c-d 
    name: "C D"
    stream_url: rtsp://c-d
`,
		},
		{
			Name:          "valid-recorder-no-values",
			ExpectedError: "",
			Data: `---
recorder:
`,
		},
		{
			Name:          "invalid-recorder-empty-ffmpegpath",
			ExpectedError: "Key: 'VigilisConfig.Recorder.FfmpegPath' Error:Field validation for 'FfmpegPath' failed on the 'filepath' tag",
			Data: `---
recorder:
  ffmpeg_path: ""
`,
		},
		{
			Name:          "invalid-recorder-relative-ffmpegpath",
			ExpectedError: "Key: 'VigilisConfig.Recorder.FfmpegPath' Error:Field validation for 'FfmpegPath' failed on the 'filepath' tag",
			Data: `---
recorder:
  ffmpeg_path: .
`,
		},
		{
			Name:             "valid-recorder-full-ffmpegpath",
			MustNotHaveError: "VigilisStorage.Recorder.FfmpegPath",
			Data: `---
recorder:
  ffmpeg_path: /usr/bin/ffmpeg
`,
		},
	}

	for _, caseData := range cases {
		t.Run(caseData.Name, func(t *testing.T) {
			// Variables are defined here so "goto fail" has access to them
			var (
				fails     validator.ValidationErrors
				errorType string
			)

			//defer func() {
			//	b := recover()
			//	if b != nil {
			//		fmt.Println(b)
			//	}
			//}()

			// Clear the variable between runs
			defer func() {
				Vigilis = VigilisConfig{}
			}()

			err := Parse([]byte(caseData.Data))
			//fmt.Println(Vigilis, caseData.ExpectedError, err)

			// NO BOOM + valid = pass
			// NO BOOM + invalid = fail
			if err == nil && caseData.ExpectedError != "" {
				errorType = "expected error not thrown"
				goto fail
			}

			// BOOM + valid = fail
			// BOOM + invalid = pass
			if err != nil {
				// Fail on non-validation errors
				isValidationError := errors.As(err, &fails)
				if !isValidationError {
					if err.Error() == caseData.ExpectedError {
						return
					}

					t.Error("Invalid ExpectedError")
					t.Errorf("Wanted: %v", caseData.ExpectedError)
					t.Errorf("Got %v", err)

					errorType = "invalid ExpectedError"
					goto fail
				}

				if caseData.ExpectedError == "" {
					if caseData.MustNotHaveError == "" {
						// If no error was expected, exit with success
						return
					}

					// Fail if an error was thrown for a value
					if slices.ContainsFunc(fails, func(fieldError validator.FieldError) bool {
						// Checking for prefix so, for example, VigilisConfig.Cameras matches VigilisConfig.Cameras[0].Id
						if strings.HasPrefix(fieldError.Namespace(), caseData.MustNotHaveError) {
							t.Log("error:", fieldError.Error())
							return true
						}
						return false
					}) {
						errorType = "error for value was thrown"
						goto fail
					}

					return
				}

				// Fail if the expected error was NOT thrown
				if !slices.ContainsFunc(fails, func(fieldError validator.FieldError) bool {
					return fieldError.Error() == caseData.ExpectedError
				}) {
					for _, fail := range fails {
						t.Log("error:", fail)
					}
					t.Log("expected:", caseData.ExpectedError)
					errorType = "expected error not thrown"
					goto fail
				}
			}

			return

		fail:
			// Variables for easier debugging/new tests
			var stringFails []string
			for _, fail := range fails {
				stringFails = append(stringFails, fail.Error())
			}

			t.Log("fail reason:", errorType)

			// NOTE Set breakpoint here to get the ExpectedError when creating new tests.
			t.Fail()
		})
	}
}
