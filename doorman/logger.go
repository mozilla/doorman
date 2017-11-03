package doorman

import (
	"github.com/sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
	"os"
)

var authzLog logrus.Logger

func init() {
	authzLog = logrus.Logger{
		Out:       os.Stdout,
		Formatter: &mozlogrus.MozLogFormatter{LoggerName: "iam", Type: "request.authorization"},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
}
