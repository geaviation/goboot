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
	"fmt"
	"github.com/gostones/goboot/config"
	"github.com/gostones/goboot/logging"
	"github.com/newrelic/go-agent"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

var settings = config.AppSettings()
var log = logging.Logger()

type NewRelicEnv struct {
	Enable  bool   `env:"goboot_newrelic.enable"`
	Name    string `env:"goboot_newrelic.name"`
	License string `env:"goboot_newrelic.license"`
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
		log.Debugf("NewRelic app name read from VCAP_APPLICATION: %v", name)
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
	Config      newrelic.Config
)

func HandleFuncAdapter(handler func(http.ResponseWriter, *http.Request), name ...string) func(http.ResponseWriter, *http.Request) {
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

func funcAdapter(fn func(), name ...string) func() (err error) {
	var defaultName string
	if len(name) == 0 {
		defaultName = ""
	} else {
		defaultName = strings.Join(name, "/")
	}
	return func() (err error) {
		var txn newrelic.Transaction
		if Application != nil {
			var pattern = defaultName
			if pattern == "" {
				pattern = nameOf(fn)
			}
			txn = Application.StartTransaction(pattern, nil, nil)
			defer txn.End()
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("Error: %s", r)
					txn.NoticeError(err)
				}
			}()
		}

		fn()

		return
	}
}

func nameOf(f interface{}) string {
	v := reflect.ValueOf(f)
	if v.Kind() == reflect.Func {
		if r := runtime.FuncForPC(v.Pointer()); r != nil {
			return r.Name()
		}
	}
	return v.String()
}

func Trace(fn func(), name ...string) (err error) {
	err = funcAdapter(fn, name...)()
	return
}
