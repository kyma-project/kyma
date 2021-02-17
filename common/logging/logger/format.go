package logger

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Format string

const (
	JSON Format = "json"
	TEXT Format = "text"
)

var allFormats = []Format{JSON, TEXT}

func MapFormat(input string) (Format, error) {
	var format = Format(input)
	switch format {
	case JSON, TEXT:
		return format, nil
	default:
		return format, errors.New(fmt.Sprintf("Given log format: %s, doesn't match with any of %v", format, allFormats))
	}
}

func (f Format) toZapEncoder() (zapcore.Encoder, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.MessageKey = "message"
	switch f {
	case JSON:
		return zapcore.NewJSONEncoder(encoderConfig), nil
	case TEXT:
		return zapcore.NewConsoleEncoder(encoderConfig), nil
	default:
		return nil, errors.New("unknown encoder")
	}
}
