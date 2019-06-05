// Copyright 2017 The Goboot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Usage: var log = logging.ContextLogger
// Setup env JSON value:
// goboot_logging={
//   "level": "DEBUG"
// }
// Logging levels: DEBUG, INFO, WARN, ERROR, PANIC, FATAL
// default is DEBUG
// FATAL will terminate your app
package logging

import (
	"os"
	"github.com/sirupsen/logrus"
	"github.com/gostones/goboot/config"
)

var contextLogger *logrus.Entry

const goboot_logging string = "goboot_logging"

var settings = config.AppSettings()

func init() {
	//Predix logstash only accepts text from stdout for now
	//logrus.SetFormatter(&logrusrus.JSONFormatter{})
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(os.Stdout)

	//Logrus has six logging levels: Debug, Info, Warning, Error, Fatal and Panic.
	//
	level := logLevel()
	logrus.SetLevel(level)

	//
	app_name := settings.GetStringEnv("VCAP_APPLICATION", "application_name")
	contextLogger = logrus.WithFields(logrus.Fields{
		"application_name": app_name,
	})

	//
	contextLogger.Infof("Logrus initialized. log level: %s", level)
}

//default to debug if env not set
func logLevel() (level logrus.Level) {
	l := settings.GetStringEnv(goboot_logging, "level")
	level, err := logrus.ParseLevel(l)
	if (err != nil) {
		level = logrus.DebugLevel
	}
	return
}

func Logger() *logrus.Entry {
	return contextLogger
}