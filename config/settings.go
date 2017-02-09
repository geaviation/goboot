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

func (s Settings) GetEnv(key string, name ...string) interface{} {
	v := s.getEnv(key)
	if len(name) == 0 {
		return v
	}

	return traverse(name, v.(map[string]interface{}))
}

func (s Settings) GetStringEnv(key string, name ...string) string {
	t := s.GetEnv(key, name...)

	switch t.(type) {
	case map[string]interface{}:
	case []interface{}:
	case string:
		return t.(string)
	case nil:
		return ""
	case int:
	case float64:
	case bool:
	default:
	}

	return fmt.Sprintf("%v", t)
}

func (s Settings) GetBoolEnv(key string, name ...string) bool {
	t := s.GetEnv(key, name...)

	b, err := strconv.ParseBool(fmt.Sprintf("%v", t))
	if err == nil {
		return b
	}
	return false
}

func (s Settings) GetIntEnv(key string, name ...string) int {
	t := s.GetEnv(key, name...)

	i, err := strconv.Atoi(fmt.Sprintf("%v", t))
	if err == nil {
		return i
	}

	return 0
}

func traverse(name []string, m map[string]interface{}) interface{} {
	switch len(name) {
	case 0:
		return m
	case 1:
		return m[name[0]]
	default:
		v := m[name[0]]
		switch v.(type) {
		case map[string]interface{}:
			return traverse(name[1:], v.(map[string]interface{}))
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
