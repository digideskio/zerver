package component

import (
	"encoding/json"
	"testing"

	"github.com/cosiner/kv"

	"github.com/cosiner/gohper/testing2"
	"github.com/cosiner/zerver"
)

func TestOptions(t *testing.T) {
	tt := testing2.Wrap(t)
	attrs := `
    {
        "server":{
            "listenAddr":"127.0.0.1:4000",
            "pathVarCount":10
        },
        "redis": {
            "maxIdle":10,
            "maxActive":10,
            "idleTimeout":10,
            "addr":":6379"
        },
        "template": {
            "delimLeft":"<%",
            "delimRight":"%>",
            "path":["views"]
        },
        "xsrf": {
            "timeout":600,
            "secret":"1234567"
        }
    }
    `
	type options struct {
		Server   zerver.ServerOption
		Redis    kv.RedisOption
		Template TemplateOption
		Xsrf     Xsrf
	}

	o := options{}
	tt.Nil(json.Unmarshal([]byte(attrs), &o))
	tt.Eq("127.0.0.1:4000", o.Server.ListenAddr)
	tt.Eq(10, o.Server.PathVarCount)

	tt.Eq(10, o.Redis.MaxIdle)
	tt.Eq(10, o.Redis.MaxActive)
	tt.Eq(10, o.Redis.IdleTimeout)
	tt.Eq(":6379", o.Redis.Addr)

	tt.Eq("<%", o.Template.DelimLeft)
	tt.Eq("%>", o.Template.DelimRight)
	tt.DeepEq([]string{"views"}, o.Template.Path)

	tt.Eq(int64(600), o.Xsrf.Timeout)
	tt.Eq("1234567", o.Xsrf.Secret)
}
