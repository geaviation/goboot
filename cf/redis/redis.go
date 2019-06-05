package redis

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/garyburd/redigo/redis"
	"github.com/gostones/goboot/config"
	"github.com/gostones/goboot/logging"
)

var log = logging.Logger()

type vcap struct {
	MaxIdle     int    `env:"VCAP_SERVICES.user-provided.0.credentials.redis.max_idle"`
	MaxActive   int    `env:"VCAP_SERVICES.user-provided.0.credentials.redis.max_active"`
	IdleTimeout string `env:"VCAP_SERVICES.user-provided.0.credentials.redis.idle_timeout"`
	Wait        bool   `env:"VCAP_SERVICES.user-provided.0.credentials.redis.wait"`
}

// GetPoolForService creates a Redigo Pool to connect to Redis service given
//  the bound service name.
func GetPoolForService(a ...string) *redis.Pool {
	settings := config.AppSettings()

	s := settings.RedisService(a...)
	if s == nil {
		log.Debugf("Redis service not found for: %v", a)
		return nil
	}
	v := s.(cfenv.Service)

	vcap := vcap{}
	err := settings.Parse(&vcap)
	if err != nil {
		log.Errorf("redis init (vcap) error: %v", err)
	}
	log.Debugln("redis init (vcap)", vcap)
	idleTimeout, err := time.ParseDuration(vcap.IdleTimeout)
	if err != nil {
		log.Errorf("redis init (vcap) error: %v", err)
	}

	addr := fmt.Sprintf("%s:%g", v.Credentials["host"], v.Credentials["port"])

	pool := &redis.Pool{
		MaxIdle:     vcap.MaxIdle,
		IdleTimeout: idleTimeout,
		MaxActive:   vcap.MaxActive,
		Wait:        vcap.Wait,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				log.Error("Failed to create redis connection: ", err)
				return nil, err
			}
			if pw := v.Credentials["password"]; pw != nil {
				if _, err := c.Do("AUTH", pw); err != nil {
					log.Error("Error authenticating to redis: ", err)
					c.Close()
					return nil, err
				}
			} else {
				log.Error("no password set for redis THIS IS BAD OUTSIDE DEVELOPMENT INSTANCES")
			}
			return c, nil
		},
		// TODO: do TestOnBorrow?
	}

	return pool
}

// Set gets a key to the value
func Set(pool *redis.Pool, key string, value interface{}) error {
	conn := pool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("SET", key, value)
	if err != nil {
		return err
	}

	return nil
}

// SetWithExpire gets a key to the value with an expiration period
func SetWithExpire(pool *redis.Pool, key string, value interface{}, expire int64) error {

	if err := Set(pool, key, value); err != nil {
		return err
	}

	conn := pool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("EXPIRE", key, expire)
	if err != nil {
		//TODO: rollback the SET when the EXPIRE fails
		return err
	}

	return nil
}

// Get gets the value of key
func Get(pool *redis.Pool, key string) (interface{}, error) {
	conn := pool.Get()
	if conn.Err() != nil {
		return nil, conn.Err()
	}
	defer conn.Close()

	value, err := conn.Do("GET", key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// GetAll gets all keys
func GetAll(pool *redis.Pool) (interface{}, error) {
	conn := pool.Get()
	if conn.Err() != nil {
		return nil, conn.Err()
	}
	defer conn.Close()

	value, err := conn.Do("KEYS", "*")
	if err != nil {
		return nil, err
	}

	return value, nil
}

// Delete deletes the entry for key
func Delete(pool *redis.Pool, key string) error {
	conn := pool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("DEL", key)

	return err
}

// RedisClient is the redis client
type RedisClient struct {
	pool    *redis.Pool
	setting interface{}
}

// NewRedisClient returns a redis client
func NewRedisClient(name ...string) *RedisClient {
	settings := config.AppSettings()

	return &RedisClient{
		setting: settings.RedisService(name...),
		pool:    GetPoolForService(name...),
	}
}

//
func (r *RedisClient) Credentials() string {
	v := r.setting.(cfenv.Service)
	return toJSONString(v.Credentials)
}

// Set applies the value to key
func (r *RedisClient) Set(key string, value interface{}) error {
	return Set(r.pool, key, toJSONString(value))
}

// SetWithExpire applies the value to key with an expiration
func (r *RedisClient) SetWithExpire(key string, value interface{}, expire int64) error {
	return SetWithExpire(r.pool, key, toJSONString(value), expire)
}

// Get returns the value for key
func (r *RedisClient) Get(key string, value interface{}) error {
	d, err := Get(r.pool, key)
	if d == nil || err != nil {
		return err
	}
	fromJSONString([]byte(d.([]uint8)), value)
	return err
}

// GetAllKeys returns all keys
func (r *RedisClient) GetAllKeys() ([]string, error) {
	values, err := GetAll(r.pool)
	if err != nil {
		return nil, err
	}

	keys := []string{}
	for _, v := range values.([]interface{}) {
		keys = append(keys, string(v.([]uint8)))
	}
	return keys, nil
}

// Delete deletes the key and its value
func (r *RedisClient) Delete(key string) error {
	return Delete(r.pool, key)
}

// Do is a generic redis command function
func (r *RedisClient) Do(cmd string, args ...interface{}) (interface{}, error) {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.Do(cmd, args...)
}

// Info returns server information
func (r *RedisClient) Info() (string, error) {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return "", err
	}
	defer conn.Close()

	info, err := conn.Do("INFO")
	return rString(info), err
}

// HSet sets a hash value
// returns true if new value
func (r *RedisClient) HSet(hash, field string, value interface{}) (bool, error) {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return false, err
	}
	defer conn.Close()
	exists, err := conn.Do("HSET", hash, field, ToRedis(value))
	return (exists.(int64) == 1), err
}

// HGet gets a hash value
func (r *RedisClient) HGet(hash, field string, value interface{}) error {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return err
	}
	defer conn.Close()

	data, err := conn.Do("HGET", hash, field)
	return FromRedis(data, value, err)
}

// HDel deletes the key(s) of a hash
func (r *RedisClient) HDel(hash string, fields ...string) error {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return err
	}
	defer conn.Close()

	cmd := []interface{}{hash}
	for _, field := range fields {
		cmd = append(cmd, field)
	}
	_, err := conn.Do("HDEL", cmd...)
	return err
}

// HKeys gets the keys of a hash
func (r *RedisClient) HKeys(hash string) ([]string, error) {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return nil, err
	}
	defer conn.Close()

	values, err := conn.Do("HKEYS", hash)
	if err != nil {
		return nil, err
	}
	return keys(values), nil
}

// HVals gets all values of a hash
func (r *RedisClient) HVals(hash string, save func([]uint8) error) error {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return err
	}
	defer conn.Close()

	reply, err := conn.Do("HVALS", hash)
	if err != nil {
		return err
	}
	for _, data := range reply.([]interface{}) {
		if err = save(data.([]uint8)); err != nil {
			return err
		}
	}
	return err
}

// GetKeys gets the keys matching the given regexps (if any)
func (r *RedisClient) GetKeys(matching ...string) ([]string, error) {
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return nil, err
	}
	defer conn.Close()

	if len(matching) == 0 {
		matching = append(matching, "*")
	}

	k := make([]string, 0, 32)
	for _, match := range matching {
		values, err := conn.Do("KEYS", match)
		if err != nil {
			return nil, err
		}
		list := values.([]interface{})
		for _, v := range list {
			k = append(k, rString(v))
		}
	}
	return k, nil
}

// GetKeysWithPrefix gets the keys matching the given prefixes
func (r *RedisClient) GetKeysWithPrefix(matching ...string) ([]string, error) {
	for i, m := range matching {
		matching[i] = m + "*"
	}
	return r.GetKeys(matching...)
}

// MGet gets the values for the given keys
func (r *RedisClient) MGet(keys ...string) ([]interface{}, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("no keys give")
	}
	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		return nil, err
	}
	defer conn.Close()
	reply, err := conn.Do("MGET", iSlice(keys)...)
	return reply.([]interface{}), err
	//return reply.([][]uint8), err
}

// helper func to convert slice of strings to interface{}
func iSlice(str []string) []interface{} {
	list := make([]interface{}, 0, len(str))
	for _, s := range str {
		list = append(list, s)
	}
	return list
}

func keys(values interface{}) []string {
	list := values.([]interface{})
	keys := make([]string, 0, len(list))
	for _, v := range list {
		keys = append(keys, rString(v))
	}
	return keys
}

func toJSONString(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func fromJSONString(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ToRedis converts a value to be written to redis
func ToRedis(value interface{}) interface{} {
	switch value.(type) {
	case []byte, string, int, int64, float64, bool:
		return value
	default:
		b, _ := json.Marshal(value)
		return string(b)
	}
}

// FromRedis converts a binary value returned from redis to its native type
func FromRedis(reply, value interface{}, bad error) (err error) {
	/*
		Per https://godoc.org/github.com/garyburd/redigo/redis

		Go Type                 Conversion
		[]byte                  Sent as is
		string                  Sent as is
		int, int64              strconv.FormatInt(v)
		float64                 strconv.FormatFloat(v, 'g', -1, 64)
		bool                    true -> "1", false -> "0"
		nil                     ""
		all other types         fmt.Print(v)
	*/
	if bad != nil {
		return bad
	}
	switch v := value.(type) {
	case *[]byte:
		*v = reply.([]byte)
	case *string:
		*v = rString(reply)
	case *int:
		var i int64
		i, err = strconv.ParseInt(rString(reply), 10, 32)
		*v = int(i)
	case *int64:
		*v, err = strconv.ParseInt(rString(reply), 10, 64)
	case *float64:
		*v, err = strconv.ParseFloat(rString(reply), 64)
	case *bool:
		*v, err = strconv.ParseBool(rString(reply))
	default:
		fromJSONString([]byte(reply.([]uint8)), value)
	}
	return nil
}

func rString(reply interface{}) string {
	if reply == nil {
		return ""
	}
	return string(reply.([]uint8))
}
