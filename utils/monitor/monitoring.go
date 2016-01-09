package monitor

import (
	"net/http"
	"net/url"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/cosiner/gohper/io2"
	"github.com/cosiner/gohper/unsafe2"
	"github.com/cosiner/zerver"
	"github.com/cosiner/zerver/handler"
)

type getHandler struct {
	doGet zerver.HandleFunc
	handler.NopMethodHandler
}

func (g *getHandler) Get(req zerver.Request, resp zerver.Response) {
	g.doGet(req, resp)
}

var path = "/status"

func Enable(monitorPath string, rt zerver.Router) (err error) {
	if monitorPath != "" {
		path = monitorPath
	}
	if !initRoutes() {
		return
	}

	for subpath, handler := range routes {
		if err = rt.Handler(path+subpath, handler); err != nil {
			return
		}
		options = append(options, "GET "+path+subpath+": "+infos[subpath]+"\n")
	}
	err = rt.Filter(path, zerver.FilterFunc(globalFilter))
	return
}

func globalFilter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	status := resp.StatusCode(0)
	if status == http.StatusNotFound {
		resp.Headers().Set("Location", path+"/options?from="+url.QueryEscape(req.URL().Path))
		resp.StatusCode(http.StatusMovedPermanently)
	} else if status == http.StatusMethodNotAllowed {
		io2.WriteString(resp, "The pprof interface only support GET request\n")
	} else {
		chain(req, resp)
	}
}

var inited bool

func Handle(path, info string, fn zerver.HandleFunc) {
	infos[path], routes[path] = info, handler.WrapMethodHandler(&getHandler{doGet: fn})
}

var options = make([]string, 0, len(infos)+1)
var routes = make(map[string]zerver.Handler)
var infos = make(map[string]string)

func pprofLookupHandler(name string) zerver.HandleFunc {
	return func(req zerver.Request, resp zerver.Response) {
		pprof.Lookup(name).WriteTo(resp, 2)
	}
}

func initRoutes() bool {
	if inited {
		return false
	}
	inited = true

	Handle("/goroutine", "Get goroutine info",
		pprofLookupHandler("goroutine"))
	Handle("/heap", "Get heap info",
		pprofLookupHandler("heap"))
	Handle("/thread", "Get thread create info",
		pprofLookupHandler("threadcreate"))
	Handle("/block", "Get block info",
		pprofLookupHandler("block"))

	Handle("/cpu", "Get CPU info, default seconds is 30, use ?seconds= to reset",
		func(req zerver.Request, resp zerver.Response) {
			var t int
			if secs := req.Vars().QueryVar("seconds"); secs != "" {
				var err error
				if t, err = strconv.Atoi(secs); err != nil {
					resp.StatusCode(http.StatusBadRequest)
					io2.WriteString(resp, secs+" is not a integer number\n")
					return
				}
			}
			if t <= 0 {
				t = 30
			}
			pprof.StartCPUProfile(resp)
			time.Sleep(time.Duration(t) * time.Second)
			pprof.StopCPUProfile()
		})

	Handle("/memory", "Get memory info",
		func(req zerver.Request, resp zerver.Response) {
			runtime.GC()
			pprof.WriteHeapProfile(resp)
		})

	Handle("/routes", "Get all routes",
		func(req zerver.Request, resp zerver.Response) {
			req.Server().PrintRouteTree(resp)
		})

	Handle("/options", "Get all pprof options",
		func(req zerver.Request, resp zerver.Response) {
			if from := req.Vars().QueryVar("from"); from != "" {
				resp.Write(unsafe2.Bytes("There is no this pprof option: " + from + "\n"))
			}
			for i := range options {
				resp.Write(unsafe2.Bytes(options[i]))
			}
		})

	return inited
}
