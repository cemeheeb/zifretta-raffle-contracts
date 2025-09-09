package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Initialize() {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.TimeKey = "timestamp"

	var err error
	log, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

func Debug(message string, fields ...zap.Field) {
	log.Debug(message, fields...)
}

func Info(message string, fields ...zap.Field) {
	log.Info(message, fields...)
}

func Warn(message string, fields ...zap.Field) {
	log.Warn(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	log.Fatal(message, fields...)
}
