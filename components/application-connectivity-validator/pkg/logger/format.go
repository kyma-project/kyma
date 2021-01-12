package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Format string

const (
	JSON = "json"
	TEXT = "text"
)

func (f Format) toZapEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	switch f {
	case JSON:
		return zapcore.NewJSONEncoder(encoderConfig)
	case TEXT:
		return zapcore.NewConsoleEncoder(encoderConfig)
	default:
		panic("unknown encoder")
	}
}
