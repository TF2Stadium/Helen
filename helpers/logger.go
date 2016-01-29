// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/stack"
)

type formatter struct{}

func (formatter) Format(e *logrus.Entry) ([]byte, error) {
	var t logrus.Formatter

	frame := stack.Caller(7)
	frame.File = stack.StripGOPATH(frame.File)
	e.Data["function"] = frame.Name
	e.Data["file"] = fmt.Sprintf("%s:%d", frame.File, frame.Line)

	if os.Getenv("LOG_FORMAT") == "json" {
		t = &logrus.JSONFormatter{}
	} else {
		t = &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.Stamp,
		}
	}

	return t.Format(e)
}

func InitLogger() {
	logrus.SetFormatter(&formatter{})
}
