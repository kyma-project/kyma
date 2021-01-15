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

var all_formats = []Format{JSON, TEXT}

func MapFormat(input string) (Format, error) {
	var format = Format(input)
	switch format {
	case JSON, TEXT:
		return format, nil
	default:
		return format, errors.New(fmt.Sprintf("Given log format: %s, doesn't match with any of %v", format, all_formats))
	}
}

func (f Format) toZapEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.MessageKey = "message"
	switch f {
	case JSON:
		return zapcore.NewJSONEncoder(encoderConfig)
	case TEXT:
		return zapcore.NewConsoleEncoder(encoderConfig)
	default:
		panic("unknown encoder")
	}
}
