// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/facebookgo/stack"
)

type formatter struct{}

func (formatter) Format(e *logrus.Entry) ([]byte, error) {
	var t logrus.Formatter

	frame := stack.Caller(7)
	e.Data["func"] = fmt.Sprintf("%s:%d", frame.Name, frame.Line)

	if os.Getenv("LOG_FORMAT") == "json" {
		t = &logrus.JSONFormatter{}
	} else {
		t = &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "15:04:05.000",
		}
	}

	return t.Format(e)
}

func InitLogger() {
	logrus.SetFormatter(&formatter{})
}
