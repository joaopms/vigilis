package util

import (
	"bytes"
	"vigilis/internal/logger"
)

func LogBuffer(buffer bytes.Buffer, bufferName string, logFunc logger.LogFunction, prefix string) {
	for buffer.Len() > 0 {
		output, err := buffer.ReadBytes('\n')
		if err != nil {
			logFunc("%v > Error reading %v: %v", prefix, bufferName, err)
		}

		// Remove the last character, \n
		output = output[:len(output)-1]

		logFunc("%v > %v", prefix, string(output))
	}
}
