package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

type Configuration struct {
	LogFile   string
	ErrorFile string
	Level     string
	Console   bool
}

func Initialize(configuration Configuration) {

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(configuration.Level)); err != nil {
		level = zapcore.DebugLevel
	}

	// Настройки кодировщика
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core

	if configuration.LogFile != "" {
		logFile, err := os.OpenFile(configuration.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(logFile),
			level,
		)
		cores = append(cores, fileCore)
	}

	if configuration.ErrorFile != "" {
		errorFile, err := os.OpenFile(configuration.ErrorFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		errorCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(errorFile),
			zapcore.ErrorLevel,
		)
		cores = append(cores, errorCore)
	}

	if configuration.Console {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	core := zapcore.NewTee(cores...)
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
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
