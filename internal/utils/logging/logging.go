package logging

import "github.com/sirupsen/logrus"

var (
	logger *logrus.Entry
)

func init() {
	if logger == nil {
		logger = logrus.NewEntry(logrus.New())
	}
}

func WithError(e error) *logrus.Entry {
	return logger.WithError(e)
}

func Entry() *logrus.Entry {
	return logger
}

func Error(args ...interface{}) {
	logger.Error(args...)
}
