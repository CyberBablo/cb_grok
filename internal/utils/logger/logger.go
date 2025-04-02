package logger

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var l *zap.Logger

func NewLogger() *zap.Logger {
	logger := zap.Must(zap.NewDevelopment())
	l = logger
	return logger
}

func GetInstance() *zap.Logger {
	if l == nil {
		return NewLogger()
	}
	return l
}

var Module = fx.Module("logger",
	fx.Provide(NewLogger),
)
