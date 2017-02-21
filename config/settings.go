// Copyright 2017 The Goboot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package config provides Settings interface for accessing Predix environment variables.
// The values can be boolean, number, string, or JSON object.
//
// An example shell script for defining and setting a json as value is given here.
// you can read the JSON or any part of it by providing the environment name as key and
// the list of object names and array indexes as the path to the value.
//
// e.g. for MY_ENV defined below, Settings.GetEnv("MY_ENV", "array", "1") will return "b"
//
//   #!/usr/bin/env bash
//   #
//   function define(){
//     IFS='\n' read -r -d '' ${1} || true;
//   }
//   #
//   export value="this is a value"
//   #
//   define MY_ENV << JSON
//   {
//   "home": "$HOME",
//   "array": ["a", "b", "c"],
//   "pwd": "`pwd`"
//   }
//   JSON
//   #
//   export MY_ENV
package config

import (
	"os"
	"github.com/cloudfoundry-community/go-cfenv"
	"strconv"
	"sync"
	"encoding/json"
	"fmt"
)

type Settings struct {
	Env   *cfenv.App

	cache map[string]interface{} //cached env and uris

	sync.Mutex
}

func (r Settings) String() string {
	return fmt.Sprintf("%s", r.cache)
}

func (r Settings) getEnv(name string) interface{} {
	r.Lock()
	defer r.Unlock()

	key := "env_" + name

	var t interface{}
	var ok bool
	if t, ok = r.cache[key]; ok {
		return t
	}

	v := os.Getenv(name)
	if err := json.Unmarshal([]byte(v), &t); err != nil {
		r.cache[key] = v
		return v
	}
	r.cache[key] = t
	return t
}

func (r Settings) getService(name string) interface{} {
	r.Lock()
	defer r.Unlock()

	key := "service_" + name

	var t interface{}
	var ok bool
	if t, ok = r.cache[key]; ok {
		return t
	}

	s, err := r.Env.Services.WithName(name)
	if err != nil {
		ss, err := r.Env.Services.WithLabel(name)
		if err != nil {
			t = nil
		} else {

			t = ss[0]
		}
	} else {

		t = *s
	}
	r.cache[key] = t

	return t
}

// GetService looks up by name and then by label and returns the service
// from VCAP_SERVICES environment variable
func (r Settings) GetService(names ...string) interface{} {
	for _, name := range names {
		s := r.getService(name)
		if s != nil {
			return s
		}
	}

	return nil
}

func (r Settings) PostgresUri() string {
	a := []string{"postgres"}
	return r.ServiceUri(a...)
}

func (r Settings) RabbitmqUri() string {
	a := []string{"rabbitmq-36", "p-rabbitmq-35"}
	return r.ServiceUri(a...)
}

// ServiceUri looks up by name and then by label and returns service uri
// from the VCAP_SERVICES environment variable
func (r Settings) ServiceUri(names ...string) string {
	return r.GetService(names...).(cfenv.Service).Credentials["uri"].(string)
}

// GetEnv returns env value for the given name.
// If the value is JSON and path is provided, return the part specified.
func (r Settings) GetEnv(name string, path ...string) interface{} {
	v := r.getEnv(name)
	if v == nil || len(path) == 0 {
		return v
	}
	m, ok := v.(map[string]interface{})
	if !ok || m == nil {
		return nil
	}
	return traverse(path, m)
}

// GetEnv returns env string value for the given name.
// If the value is JSON and path is provided, return the part specified.
func (r Settings) GetStringEnv(name string, path ...string) string {
	t := r.GetEnv(name, path...)

	switch t.(type) {
	case string:
		return t.(string)
	case nil:
		return ""
	case float64:
	default:
	}

	return fmt.Sprintf("%v", t)
}

// GetEnv returns env boolean value for the given name.
// If the value is JSON and path is provided, return the part specified.
func (r Settings) GetBoolEnv(name string, path ...string) bool {
	t := r.GetEnv(name, path...)

	b, err := strconv.ParseBool(fmt.Sprintf("%v", t))
	if err == nil {
		return b
	}
	return false
}

// GetEnv returns env int value for the given name.
// If the value is JSON and path is provided, return the part specified.
func (r Settings) GetIntEnv(name string, path ...string) int {
	t := r.GetEnv(name, path...)

	i, err := strconv.Atoi(fmt.Sprintf("%v", t))
	if err == nil {
		return i
	}

	return 0
}

func (r Settings) Parse(v interface{}) error {
	return parse(v)
}

func traverse(path []string, t interface{}) interface{} {
	var next = func() interface{} {
		switch t.(type) {
		case []interface{}:
			idx, err := strconv.Atoi(path[0])
			if err == nil {
				return t.([]interface{})[idx]
			}
		case map[string]interface{}:
			return t.(map[string]interface{})[path[0]]
		}
		return nil
	}

	switch len(path) {
	case 0:
		return t
	case 1:
		return next()
	default:
		v := next()
		switch v.(type) {
		case []interface{}:
			return traverse(path[1:], v)
		case map[string]interface{}:
			return traverse(path[1:], v)
		}
	}

	return nil
}

func NewSettings() *Settings {
	env, _ := cfenv.Current()

	return &Settings{
		Env: env,
		cache: make(map[string]interface{}),
	}
}

var (
	settings = NewSettings()
)

func AppSettings() *Settings {
	return settings
}