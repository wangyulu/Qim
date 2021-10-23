package logger

import (
	"time"

	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type Settings struct {
	Filename    string
	Level       string
	RollingDays uint
	Format      string
}

func Init(settings Settings) error {
	if settings.Level == "" {
		settings.Level = "debug"
	}

	ll, err := logrus.ParseLevel(settings.Level)
	if err == nil {
		std.SetLevel(ll)
	} else {
		std.Error("Invalid log level")
	}

	if settings.Filename == "" {
		return nil
	}

	if settings.RollingDays == 0 {
		settings.RollingDays = 7
	}

	writer, err := rotatelogs.New(
		settings.Filename+".%Y%m%d",
		rotatelogs.WithLinkName(settings.Filename),
		rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithRotationCount(settings.RollingDays),
	)

	if err != nil {
		return err
	}

	var logfmt logrus.Formatter
	if settings.Format == "json" {
		logfmt = &logrus.JSONFormatter{
			DisableTimestamp: false,
		}
	} else {
		logfmt = &logrus.TextFormatter{
			DisableColors: true,
		}
	}

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
	}, logfmt)

	std.AddHook(lfsHook)

	return nil
}
