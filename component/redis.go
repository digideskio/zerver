package component

import (
	"time"

	"github.com/cosiner/zerver"
	"github.com/garyburd/redigo/redis"
)

const (
	OPT_REDIS  = "RedisOption"
	COMP_REDIS = "RedisComponent"
)

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

	Redis struct {
		redis.Pool
		errorLogger func(...interface{})
	}
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
	r.MaxIdle = o.MaxIdle
	r.MaxActive = o.MaxActive
	r.IdleTimeout = time.Duration(o.IdleTimeout) * time.Second
	r.Dial = o.Dial
	r.errorLogger = env.Server().Errorln // use server error logger
	return r.Update("PING")
}

func (r *Redis) Conn() redis.Conn {
	return r.Get()
}

func (r *Redis) Exec(cmd string, args ...interface{}) (interface{}, error) {
	c := r.Get()
	reply, err := c.Do(cmd, args...)
	r.onErrorLog(c.Close())
	return reply, err
}

func (r *Redis) Query(cmd string, args ...interface{}) (interface{}, error) {
	c := r.Get()
	reply, err := c.Do(cmd, args...)
	r.onErrorLog(c.Close())
	return reply, err
}

func (r *Redis) Update(cmd string, args ...interface{}) error {
	c := r.Get()
	_, err := c.Do(cmd, args...)
	r.onErrorLog(c.Close())
	return err
}

func (r *Redis) Destroy() {
	r.onErrorLog(r.Close())
}

func (r *Redis) onErrorLog(err error) {
	if err != nil {
		r.errorLogger(err)
	}
}
