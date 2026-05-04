package logger

import "go.uber.org/zap"

var SugaredLogger *zap.SugaredLogger

func InitLogger() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}

	SugaredLogger = logger.Sugar()

	return nil
}