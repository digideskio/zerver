package component

import (
	"time"

	"github.com/cosiner/ygo/log"
	"github.com/cosiner/zerver"
	"github.com/garyburd/redigo/redis"
)

const (
	REDIS = "Redis"
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
		logger log.Logger
		redis.Pool
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

func (r *Redis) Init(env zerver.Environment) error {
	compEnv := env.(zerver.ComponentEnvironment)
	var o *RedisOption
	if op := compEnv.GetSetAttr(REDIS, nil); op == nil {
		o = &RedisOption{}
	} else {
		o = op.(*RedisOption)
	}
	o.init()

	r.MaxIdle = o.MaxIdle
	r.MaxActive = o.MaxActive
	r.IdleTimeout = time.Duration(o.IdleTimeout) * time.Second
	r.Dial = o.Dial
	r.logger = env.Logger()

	_, err := r.Get().Do("PING")
	return err
}

func (r *Redis) Conn() redis.Conn {
	return r.Get()
}

func (r *Redis) Exec(cmd string, args ...interface{}) (interface{}, error) {
	c := r.Get()
	reply, err := c.Do(cmd, args...)
	r.PanicLog(c.Close())

	return reply, err
}

func (r *Redis) Query(cmd string, args ...interface{}) (interface{}, error) {
	c := r.Get()
	reply, err := c.Do(cmd, args...)
	r.PanicLog(c.Close())

	return reply, err
}

func (r *Redis) Update(cmd string, args ...interface{}) error {
	c := r.Get()
	_, err := c.Do(cmd, args...)
	r.PanicLog(c.Close())

	return err
}

func (r *Redis) Destroy() {
	r.PanicLog(r.Close())
}

func (r *Redis) PanicLog(err error) {
	if err != nil {
		r.logger.Panicln(err)
	}
}
