package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Format string

const (
	JSON Format = "json"
	TEXT Format = "text"
)

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
