package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// noticed this could be clearer
var Log *logrus.Logger

// Init creates a global Logrus logger configured via LOG_LEVEL env variable.
func Init() {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)

	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if levelStr == "" {
		levelStr = "info"
	}
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		Log.Warnf("invalid LOG_LEVEL '%s', falling back to 'info'", levelStr)
		level = logrus.InfoLevel
	}
	Log.SetLevel(level)
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}