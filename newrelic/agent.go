// Copyright 2017 The Goboot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Setup env JSON value:
// goboot_newrelic={
//   "enable": true,
//   "name: "Your_App_Name",
//   "license: "__YOUR_NEW_RELIC_LICENSE_KEY__"
// }
//
package newrelic

import (
	"github.com/geaviation/goboot/config"
	"github.com/geaviation/goboot/logging"
	"github.com/newrelic/go-agent"
	"net/http"
	"strings"
)

const (
	goboot_newrelic string = "goboot_newrelic"
	newrelic_appname = "name"
	newrelic_license = "license"
)

var settings = config.AppSettings()
var log = logging.ContextLogger

type NewRelicEnv struct {
	Enable  bool          `env:"goboot_newrelic.enable"`
	Name    string        `env:"goboot_newrelic.name"`
	License string        `env:"goboot_newrelic.license"`
}

func init() {
	env := NewRelicEnv{}

	err := settings.Parse(&env)
	if err != nil {
		log.Errorf("NewRelic init error: %v", err)
		return
	}

	log.Debugf("NewRelic goboot_newrelic.enable: %v", env.Enable)
	if !env.Enable {
		return
	}

	name := env.Name
	log.Debugf("NewRelic goboot_newrelic.name: %v", name)
	if name == "" {
		name = settings.GetStringEnv("VCAP_APPLICATION", "application_name")
		log.Debugf("NewRelic app name read from VCAP_APPLICATION: ", name)
	}

	Config = newrelic.NewConfig(name, env.License)

	Application, err = newrelic.NewApplication(Config)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("NewRelic Application Name: %s  enabled %v: ", name, env.Enable)
}

var (
	Application newrelic.Application
	Config newrelic.Config
)

func HandlerAdapter(handler func(http.ResponseWriter, *http.Request), name...string) func(http.ResponseWriter, *http.Request) {
	var defaultName string
	if len(name) == 0 {
		defaultName = ""
	} else {
		defaultName = strings.Join(name, "/")
	}
	return func(res http.ResponseWriter, req *http.Request) {
		if Application != nil {
			var pattern = defaultName
			if pattern == "" {
				pattern = req.URL.Path
			}
			txn := Application.StartTransaction(pattern, res, req)
			defer txn.End()
		}

		handler(res, req)
	}
}