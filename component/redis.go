package component

import (
	"github.com/cosiner/kv"
	"github.com/cosiner/zerver"
	"github.com/cosiner/gohper/encoding"
)

const (
	REDIS = "Redis"
)

type (
	RedisOption struct {
		Option kv.RedisOption
		Codec  encoding.Codec
	}

	Redis struct {
		kv.Store
	}
)

func NewRedis() *Redis {
	return &Redis{}
}

func (r *Redis) Init(env zerver.Env) error {
	compEnv := env.(*zerver.CompEnv)
	var opt interface{}
	var codec encoding.Codec
	switch t := compEnv.GetSetAttr(REDIS, nil).(type) {
	case nil:
	case *RedisOption:
		opt = t.Option
		codec = t.Codec
	case RedisOption:
		opt = t.Option
		codec = t.Codec
	default:
		opt = t
	}
	r.Store = kv.NewRedis(opt, codec)

	return nil
}

func (r *Redis) Destroy() {
}
