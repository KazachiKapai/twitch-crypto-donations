package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type LogrusAdapter struct {
	logger *logrus.Logger
}

func New(logger *logrus.Logger) *LogrusAdapter {
	return &LogrusAdapter{logger: logger}
}

func (a *LogrusAdapter) Info(msg string, ctx ...interface{}) {
	fields := logrus.Fields{}
	for i := 0; i < len(ctx); i += 2 {
		if i+1 < len(ctx) {
			key := fmt.Sprint(ctx[i])
			fields[key] = ctx[i+1]
		}
	}
	a.logger.WithFields(fields).Info(msg)
}
