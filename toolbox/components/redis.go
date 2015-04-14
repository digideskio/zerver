package components

import (
	"time"

	"github.com/cosiner/zerver"
	"github.com/garyburd/redigo/redis"
)

const (
	OPT_REDIS  = "RedisOption"
	COMP_REDIS = "RedisComponent"
)

var ErrNoItem = redis.ErrNil

type (
	RedisOption struct {
		// Maximum number of idle connections in the pool.
		MaxIdle int

		// Maximum number of connections allocated by the pool at a given time.
		// When zero, there is no limit on the number of connections in the pool.
		MaxActive int

		// Close connections after remaining idle for this seconds. If the value
		// is zero, then idle connections are not closed. Applications should set
		// the timeout to a value less than the server's timeout.
		IdleTimeout int

		Addr string

		Dial func() (redis.Conn, error)
	}

	Redis redis.Pool
)

func (o *RedisOption) init() {
	if o.MaxIdle == 0 {
		o.MaxIdle = 8
	}
	if o.Dial == nil {
		var addr string
		if addr = o.Addr; addr == "" {
			addr = ":6379"
		}
		o.Dial = func() (redis.Conn, error) { return redis.Dial("tcp", addr) }
	}
}

func NewRedis() *Redis {
	return &Redis{}
}

func (r *Redis) Init(env zerver.Enviroment) error {
	var o *RedisOption
	if op := env.Server().Attr(OPT_REDIS); op == nil {
		o = &RedisOption{}
	} else {
		o = op.(*RedisOption)
		env.Server().RemoveAttr(OPT_REDIS)
	}
	o.init()
	pool := (*redis.Pool)(r)
	pool.MaxIdle = o.MaxIdle
	pool.MaxActive = o.MaxActive
	pool.IdleTimeout = time.Duration(o.IdleTimeout) * time.Second
	pool.Dial = o.Dial
	return r.Update("PING")
}

func (r *Redis) Conn() redis.Conn {
	return (*redis.Pool)(r).Get()
}

func (r *Redis) Exec(cmd string, args ...interface{}) (interface{}, error) {
	c := (*redis.Pool)(r).Get()
	reply, err := c.Do(cmd, args...)
	c.Close()
	return reply, err
}

func (r *Redis) Query(cmd string, args ...interface{}) (interface{}, error) {
	c := (*redis.Pool)(r).Get()
	reply, err := c.Do(cmd, args...)
	c.Close()
	return reply, err
}

func (r *Redis) Update(cmd string, args ...interface{}) error {
	c := (*redis.Pool)(r).Get()
	_, err := c.Do(cmd, args...)
	c.Close()
	return err
}

func (r *Redis) Pool() *redis.Pool {
	return (*redis.Pool)(r)
}

func (r *Redis) Destroy() {
	(*redis.Pool)(r).Close()
}
