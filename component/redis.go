package component

import (
	"github.com/cosiner/kv"
	"github.com/cosiner/zerver"
)

const (
	REDIS = "Redis"
)

type (
	RedisOption struct {
		Option kv.RedisOption
		Codec  kv.Codec
	}

	Redis struct {
		kv.Store
	}
)

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

	r.Store = kv.NewRedis(o.Option, o.Codec)

	return nil
}

// func (r *Redis) Conn() redis.Conn {
// 	return r.Get()
// }

// func (r *Redis) Exec(cmd string, args ...interface{}) (interface{}, error) {
// 	c := r.Get()
// 	reply, err := c.Do(cmd, args...)
// 	r.errLog(c.Close())

// 	return reply, err
// }

// func (r *Redis) Query(cmd string, args ...interface{}) (interface{}, error) {
// 	c := r.Get()
// 	reply, err := c.Do(cmd, args...)
// 	r.errLog(c.Close())

// 	return reply, err
// }

// func (r *Redis) Update(cmd string, args ...interface{}) error {
// 	c := r.Get()
// 	_, err := c.Do(cmd, args...)
// 	r.errLog(c.Close())

// 	return err
// }

// func (r *Redis) Destroy() {
// 	r.errLog(r.Close())
// }

// func (r *Redis) errLog(err error) {
// 	if err != nil {
// 		r.logger.Warnln(err)
// 	}
// }

// func (*Redis) IsErrNil(err error) bool {
// 	return err == redis.ErrNil
// }
// func (*Redis) Int(reply interface{}, err error) (int, error) {
// 	return redis.Int(reply, err)
// }
// func (*Redis) Int64(reply interface{}, err error) (int64, error) {
// 	return redis.Int64(reply, err)
// }
// func (*Redis) Uint64(reply interface{}, err error) (uint64, error) {
// 	return redis.Uint64(reply, err)
// }
// func (*Redis) Float64(reply interface{}, err error) (float64, error) {
// 	return redis.Float64(reply, err)
// }
// func (*Redis) String(reply interface{}, err error) (string, error) {
// 	return redis.String(reply, err)
// }
// func (*Redis) Bytes(reply interface{}, err error) ([]byte, error) {
// 	return redis.Bytes(reply, err)
// }
// func (*Redis) Bool(reply interface{}, err error) (bool, error) {
// 	return redis.Bool(reply, err)
// }
// func (*Redis) Values(reply interface{}, err error) ([]interface{}, error) {
// 	return redis.Values(reply, err)
// }
// func (r *Redis) Strings(reply interface{}, err error) ([]string, error) {
// 	return redis.Strings(reply, err)
// }
