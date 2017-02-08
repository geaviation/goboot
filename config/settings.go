package config

import (
	"os"
	"github.com/cloudfoundry-community/go-cfenv"
	"strconv"
	"sync"
)

type Settings struct {
	Env   *cfenv.App

	cache map[string]string //cached env and uris

	sync.Mutex
}

func (s Settings) getEnv(name string) string {
	s.Lock()
	defer s.Unlock()

	key := "env_" + name

	var val string
	var present bool
	if val, present = s.cache[key]; present {
		return val
	}
	val = os.Getenv(name)
	s.cache[key] = val

	return val
}

func (s Settings) getUriByName(name string) string {
	s.Lock()
	defer s.Unlock()

	key := "name_" + name

	var val string
	var ok bool
	if val, ok = s.cache[key]; ok {
		return val
	}

	svc, err := s.Env.Services.WithName(name)
	if err != nil {
		val = ""
	} else {

		val = svc.Credentials["uri"].(string)
	}
	s.cache[key] = val

	return val
}

func (s Settings) getUriByLabel(label string) string {
	s.Lock()
	defer s.Unlock()

	key := "label_" + label

	var val string
	var ok bool
	if val, ok = s.cache[key]; ok {
		return val
	}

	svc, err := s.Env.Services.WithLabel(label)
	if err != nil {
		val = ""
	} else {

		val = svc[0].Credentials["uri"].(string)
	}
	s.cache[key] = val

	return val
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

func (s Settings) GetEnv(key string) string {
	return s.getEnv(key)
}

func (s Settings) GetIntEnv(key string, def int) int {
	v := s.getEnv(key)
	if v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return def
}

func NewSettings() Settings {
	var s Settings
	env, _ := cfenv.Current()

	s = Settings{
		Env: env,
		cache: make(map[string]string),
	}
	return s
}
