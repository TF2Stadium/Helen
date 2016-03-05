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

var (
	json logrus.Formatter = &logrus.JSONFormatter{
		TimestampFormat: time.Stamp,
	}
	text logrus.Formatter = &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05.000",
	}
)

func (formatter) Format(e *logrus.Entry) ([]byte, error) {
	var f logrus.Formatter

	frame := stack.Caller(7)

	if os.Getenv("LOG_FORMAT") == "json" {
		e.Data["func"] = frame.Name
		e.Data["line"] = frame.Line
		e.Data["file"] = frame.File

		f = json
	} else {
		e.Data["func"] = fmt.Sprintf("%s:%d", frame.Name, frame.Line)

		f = text
	}

	return f.Format(e)
}

func init() {
	logrus.SetFormatter(&formatter{})
}
