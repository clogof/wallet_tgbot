package utils

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Loggers *zap.SugaredLogger

func InitLogger(loggerPath string) error {
	writerSyncer, err := getLogWriter(loggerPath)
	if err != nil {
		return err
	}
	encoder := getEncoder()

	core := zapcore.NewCore(encoder, writerSyncer, zapcore.DebugLevel)

	logger := zap.New(core)
	Loggers = logger.Sugar()

	return nil
}

func getEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewJSONEncoder(config)
}

func getLogWriter(path string) (zapcore.WriteSyncer, error) {
	d := filepath.Dir(path)
	if d != "." {
		err := os.MkdirAll(d, 0777)
		if err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(file), nil
}
