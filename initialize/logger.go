package initialize

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"oneclick-metrics-go/global"
	"strings"
)

func InitLogger() {
	var level zapcore.Level
	err := level.UnmarshalText([]byte(strings.ToLower(global.ServerConfig.LogInfo.Level)))
	if err != nil {
		level = zapcore.InfoLevel
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, _ := cfg.Build()

	zap.ReplaceGlobals(logger)
}
