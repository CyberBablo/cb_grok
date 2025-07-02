package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZapConfig для логгера
type ZapConfig struct {
	Level        string   `yaml:"level"`
	Development  bool     `yaml:"development"`
	Encoding     string   `yaml:"encoding"`
	OutputPaths  []string `yaml:"output_paths"`
	FileLog      bool     `yaml:"file_log"`
	FilePath     string   `yaml:"filename"`
	FileMaxSize  int      `yaml:"file_max_size"`
	FileCompress bool     `yaml:"file_compress"`
}

// NewZapLogger создает новый экземпляр zap логгера
func NewZapLogger(cfg ZapConfig) (*zap.Logger, error) {
	// Парсим уровень логирования
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Создаем encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
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
	}

	// В development режиме используем более читаемый формат
	if cfg.Development {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		encoderConfig.ConsoleSeparator = " | "
	}

	// Создаем encoder
	var encoder zapcore.Encoder
	switch cfg.Encoding {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Создаем writers для разных output paths
	var writers []zapcore.WriteSyncer
	for _, path := range cfg.OutputPaths {
		switch path {
		case "stdout":
			writers = append(writers, zapcore.AddSync(os.Stdout))
		case "stderr":
			writers = append(writers, zapcore.AddSync(os.Stderr))
		default:
			return nil, fmt.Errorf("unsupported output path: %s", path)
		}
	}

	// Объединяем writers
	combinedWriter := zapcore.NewMultiWriteSyncer(writers...)

	// Создаем core
	core := zapcore.NewCore(encoder, combinedWriter, level)
	core_merge := core

	if cfg.FileLog {
		logFile := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.FileMaxSize,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   cfg.FileCompress,
		}

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileWriteSyncer := zapcore.AddSync(logFile)
		fileCore := zapcore.NewCore(
			fileEncoder,
			fileWriteSyncer,
			zapcore.InfoLevel,
		)
		core_merge = zapcore.NewTee(core, fileCore)

	} else {

		core_merge = zapcore.NewTee(core)
	}

	// Создаем логгер
	logger := zap.New(core_merge)

	// Добавляем опции
	if cfg.Development {
		logger = logger.WithOptions(
			zap.Development(),
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
	} else {
		logger = logger.WithOptions(
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.PanicLevel),
		)
	}

	zap.ReplaceGlobals(logger)

	return logger, nil
}

// MustNewZapLogger создает логгер или паникует при ошибке
func MustNewZapLogger(cfg ZapConfig) *zap.Logger {
	logger, err := NewZapLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
}

// WithContext добавляет контекстуальные поля к логгеру
func WithContext(logger *zap.Logger, fields ...zap.Field) *zap.Logger {
	return logger.With(fields...)
}
