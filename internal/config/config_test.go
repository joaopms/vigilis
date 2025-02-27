package config

import (
	"testing"
)

func TestConfigValidation(t *testing.T) {
	cases := []struct {
		Name          string
		ExpectedError string
		Data          string
	}{
		// Bad data
		{
			Name:          "empty",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras' Error:Field validation for 'Cameras' failed on the 'required' tag",
			Data:          "",
		},
		{
			Name:          "not-yaml",
			ExpectedError: "[1:1] string was used where mapping is expected\n>  1 | dummydummy\n       ^\n",
			Data:          "dummydummy",
		},

		// Storage
		{
			Name:          "only-storage",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras' Error:Field validation for 'Cameras' failed on the 'required' tag",
			Data: `---
storage:
`,
		},
		{
			Name:          "only-storage-and-empty-path",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras' Error:Field validation for 'Cameras' failed on the 'required' tag",
			Data: `---
storage:
  path:
`,
		},
		{
			Name:          "only-storage-and-path-without-slash",
			ExpectedError: "Key: 'VigilisConfig.Storage.Path' Error:Field validation for 'Path' failed on the 'dirpath' tag\nKey: 'VigilisConfig.Cameras' Error:Field validation for 'Cameras' failed on the 'required' tag",
			Data: `---
storage:
  path: /tmp/vigilis
`,
		},
		{
			Name:          "only-storage-and-path",
			ExpectedError: "Key: 'VigilisConfig.Cameras' Error:Field validation for 'Cameras' failed on the 'required' tag",
			Data: `---
storage:
  path: /tmp/vigilis/
`,
		},

		// Cameras
		{
			Name:          "only-cameras",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras' Error:Field validation for 'Cameras' failed on the 'required' tag",
			Data: `---
cameras:
`,
		},
		{
			Name:          "only-cameras-and-empty-values",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras[0].Id' Error:Field validation for 'Id' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras[0].Name' Error:Field validation for 'Name' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras[0].StreamUrl' Error:Field validation for 'StreamUrl' failed on the 'required' tag",
			Data: `---
cameras:
  - id:
    name:
    stream_url:
`,
		},
		{
			Name:          "only-cameras-and-short-values",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras[0].StreamUrl' Error:Field validation for 'StreamUrl' failed on the 'gte' tag",
			Data: `---
cameras:
  - id: a
    name: A
    stream_url: a://a
`,
		},
		{
			Name:          "only-cameras-and-invalid-stream-url",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras[0].StreamUrl' Error:Field validation for 'StreamUrl' failed on the 'url' tag",
			Data: `---
cameras:
  - id: a
    name: A
    stream_url: 12345678
`,
		},
		{
			Name:          "only-cameras-and-valid-stream-url",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag",
			Data: `---
cameras:
  - id: a
    name: A
    stream_url: rtsp://a
`,
		},
		{
			Name:          "only-cameras-and-long-values",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras[0].Id' Error:Field validation for 'Id' failed on the 'lte' tag\nKey: 'VigilisConfig.Cameras[0].Name' Error:Field validation for 'Name' failed on the 'lte' tag",
			Data: `---
cameras:
  - id: 12345678901234567890A
    name: 123456789012345678901234567890A
    stream_url: rtsp://a
`,
		},
		{
			Name:          "only-two-cameras-and-id-not-alphanum",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras[0].Id' Error:Field validation for 'Id' failed on the 'alphanum' tag\nKey: 'VigilisConfig.Cameras[1].Id' Error:Field validation for 'Id' failed on the 'alphanum' tag",
			Data: `---
cameras:
  - id: "a b"
    name: "a b"
    stream_url: rtsp://a
  - id: a! 
    name: "a b!"
    stream_url: rtsp://a
`,
		},
		{
			Name:          "only-two-cameras-and-repeated-ids",
			ExpectedError: "Key: 'VigilisConfig.Storage' Error:Field validation for 'Storage' failed on the 'required' tag\nKey: 'VigilisConfig.Cameras' Error:Field validation for 'Cameras' failed on the 'unique' tag",
			Data: `---
cameras:
  - id: a
    name: A
    stream_url: rtsp://a
  - id: a
    name: B
    stream_url: rtsp://a
`,
		},
		{
			Name:          "valid-storage-two-valid-cameras",
			ExpectedError: "",
			Data: `---
storage:
  path: /tmp/vigilis/

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
			Name:          "valid-storage-two-valid-cameras-empty-recorder",
			ExpectedError: "",
			Data: `---
storage:
  path: /tmp/vigilis/

recorder:

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
			Name:          "valid-storage-two-valid-cameras-empty-ffmpegpath",
			ExpectedError: "",
			Data: `---
storage:
  path: /tmp/vigilis/

recorder:
  ffmpeg_path: ""

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
			Name:          "valid-storage-two-valid-cameras-invalid-ffmpegpath",
			ExpectedError: "Key: 'VigilisConfig.Recorder.FfmpegPath' Error:Field validation for 'FfmpegPath' failed on the 'filepath' tag",
			Data: `---
storage:
  path: /tmp/vigilis/

recorder:
  ffmpeg_path: "."

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
			Name:          "valid-storage-two-valid-cameras-valid-ffmpegpath",
			ExpectedError: "",
			Data: `---
storage:
  path: /tmp/vigilis/

recorder:
  ffmpeg_path: /usr/bin/ffmpeg

cameras:
  - id: a
    name: A
    stream_url: rtsp://a
  - id: b
    name: B
    stream_url: rtsp://b
`,
		},
	}

	for _, caseData := range cases {
		t.Run(caseData.Name, func(t *testing.T) {
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

			// BOOM + valid = fail
			// BOOM + invalid = pass
			// NO BOOM + valid = pass
			// NO BOOM + invalid = fail
			if (err == nil && caseData.ExpectedError != "") || (err != nil && (err.Error() != caseData.ExpectedError)) {
				// NOTE Set breakpoint here to get the expectedError when creating new tests
				t.Fail()
			}
		})
	}
}
