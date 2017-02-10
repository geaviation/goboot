// Copyright 2017 The Goboot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package config provides Settings struct for accessing Predix environment variables.
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

func (s Settings) String() string {
	return fmt.Sprintf("%s", s.cache)
}

func (s Settings) getEnv(name string) interface{} {
	s.Lock()
	defer s.Unlock()

	key := "env_" + name

	var t interface{}
	var ok bool
	if t, ok = s.cache[key]; ok {
		return t
	}

	v := os.Getenv(name)
	if err := json.Unmarshal([]byte(v), &t); err != nil {
		s.cache[key] = v
		return v
	}
	s.cache[key] = t
	return t
}

func (s Settings) getUriByName(name string) string {
	s.Lock()
	defer s.Unlock()

	key := "name_" + name

	var t interface{}
	var ok bool
	if t, ok = s.cache[key]; ok {
		return t.(string)
	}

	svc, err := s.Env.Services.WithName(name)
	if err != nil {
		t = ""
	} else {

		t = svc.Credentials["uri"].(string)
	}
	s.cache[key] = t

	return t.(string)
}

func (s Settings) getUriByLabel(label string) string {
	s.Lock()
	defer s.Unlock()

	key := "label_" + label

	var t interface{}
	var ok bool
	if t, ok = s.cache[key]; ok {
		return t.(string)
	}

	svc, err := s.Env.Services.WithLabel(label)
	if err != nil {
		t = ""
	} else {

		t = svc[0].Credentials["uri"].(string)
	}
	s.cache[key] = t

	return t.(string)
}

func (s Settings) getUri(labels []string, name ...string) string {
	if len(name) != 0 {
		return s.getUriByName(name[0])
	}

	for _, label := range labels {
		uri := s.getUriByLabel(label)
		if uri != "" {
			return uri
		}
	}

	return ""
}

func (s Settings) PostgresUri(name ...string) string {
	labels := []string{"postgres"}
	return s.getUri(labels, name ...)
}

func (s Settings) RabbitmqUri(name ...string) string {
	labels := []string{"rabbitmq-36", "p-rabbitmq-35"}
	return s.getUri(labels, name ...)
}

func (s Settings) ServiceUri(name ...string) string {
	labels := name
	return s.getUri(labels, name ...)
}

// GetEnv returns env value for the given name.
// If the value is JSON and path is provided, return the part specified.
func (s Settings) GetEnv(name string, path ...string) interface{} {
	v := s.getEnv(name)
	if len(path) == 0 {
		return v
	}

	return traverse(path, v.(map[string]interface{}))
}

// GetEnv returns env string value for the given name.
// If the value is JSON and path is provided, return the part specified.
func (s Settings) GetStringEnv(name string, path ...string) string {
	t := s.GetEnv(name, path...)

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
func (s Settings) GetBoolEnv(name string, path ...string) bool {
	t := s.GetEnv(name, path...)

	b, err := strconv.ParseBool(fmt.Sprintf("%v", t))
	if err == nil {
		return b
	}
	return false
}

// GetEnv returns env int value for the given name.
// If the value is JSON and path is provided, return the part specified.
func (s Settings) GetIntEnv(name string, path ...string) int {
	t := s.GetEnv(name, path...)

	i, err := strconv.Atoi(fmt.Sprintf("%v", t))
	if err == nil {
		return i
	}

	return 0
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

func NewSettings() Settings {
	var s Settings
	env, _ := cfenv.Current()

	s = Settings{
		Env: env,
		cache: make(map[string]interface{}),
	}
	return s
}
