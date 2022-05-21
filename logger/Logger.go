package logger

import (
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Initialize(conf Config) {
	initialFields := map[string]interface{}{
		"app": &map[string]string{
			"name": conf.AppName,
			"arch": runtime.GOARCH,
			"os":   runtime.GOOS,
		},
	}

	if conf.Version != "" {
		m, _ := initialFields["app"].(map[string]string)
		m["version"] = conf.Version
	}

	if conf.CommitSha != "" {
		m, _ := initialFields["app"].(map[string]string)
		m["commit"] = conf.CommitSha
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(getLogLevel(conf.LogLevel)),
		Development:      conf.IsDevelopment,
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		Sampling:         nil,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "severity",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		InitialFields: initialFields,
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	Logger = logger
}

func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	}

	return zapcore.DebugLevel
}
