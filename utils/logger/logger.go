package logger

import (
	"os"
	"time"

	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type LogInfo logrus.Fields

func InitLogger() {

	env := os.Getenv("ENV")
	isLocalHost := env == "local"
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})

	if !isLocalHost {
		// configure file system hook
		configureLocalFileSystemHook()
	}
}

func configureLocalFileSystemHook() {
	var level logrus.Level
	loglevel := os.Getenv("LOG_LEVEL")
	path := "/var/log/provapi.log"
	if loglevel == "debug" {
		level = logrus.DebugLevel
		logrus.SetLevel(level)
	}
	if loglevel == "error" {
		level = logrus.ErrorLevel
		logrus.SetLevel(level)
	}
	if loglevel == "warn" {
		level = logrus.WarnLevel
		logrus.SetLevel(level)
	}
	if loglevel == "info" {
		level = logrus.InfoLevel
		logrus.SetLevel(level)
	}
	rLogs, err := rotatelogs.New(
		path+".%Y_%m_%d_%H_%M",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(30*86400)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(86400)*time.Second),
	)
	if err != nil {
		logrus.Errorln("Local file system hook initialize fail")
		return
	}
	logrus.AddHook(lfshook.NewHook(lfshook.WriterMap{
		level: rLogs,
	}, &logrus.JSONFormatter{}))
}
