// Package monitor provide a simple monitoring interface for zerver, all monitor is
// handled GET request
// use Handle to add a custom monitor, it should be called before Enable
// for there is only one change to init
package monitor

import (
	"net/http"
	"net/url"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/cosiner/gohper/unsafe2"
	"github.com/cosiner/zerver"
)

var path = "/status"

func Enable(p string, rt zerver.Router, rootFilters zerver.RootFilters) (err error) {
	if p != "" {
		path = p
	}
	if !initRoutes() {
		return
	}

	for subpath, handler := range routes {
		if err = rt.HandleFunc(path+subpath, "GET", handler); err != nil {
			return
		}
		options = append(options, "GET "+path+subpath+": "+infos[subpath]+"\n")
	}

	if rootFilters == nil {
		err = rt.Handle(path, globalFilter)
	} else {
		rootFilters.Add(globalFilter)
	}

	return
}

func NewServer(p string) (*zerver.Server, error) {
	if p == "" {
		p = "/"
	}

	s := zerver.NewServer()
	// s.AddHandleFunc("/stop", "GET", func(req zerver.Request, resp zerver.Response) {
	// 	req.Server().Destroy()
	// })
	// infos["/stop"] = "stop pprof server"
	return s, Enable("/", s.Router, s.RootFilters)
}

func globalFilter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	resp.SetContentType("text/plain", nil)

	if resp.Status() == http.StatusNotFound {
		resp.SetHeader("Location", path+"/options?from="+url.QueryEscape(req.URL().Path))
		resp.ReportMovedPermanently()
	} else if resp.Status() == http.StatusMethodNotAllowed {
		resp.WriteString("The pprof interface only support GET request\n")
	} else {
		chain(req, resp)
	}
}

var inited bool

func Handle(path, info string, handler zerver.HandleFunc) {
	infos[path], routes[path] = info, handler
}

var options = make([]string, 0, len(infos)+1)
var routes = make(map[string]zerver.HandleFunc)
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
			if secs := req.Param("seconds"); secs != "" {
				var err error
				if t, err = strconv.Atoi(secs); err != nil {
					resp.ReportBadRequest()
					resp.WriteString(secs + " is not a integer number\n")
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
			if from := req.Param("from"); from != "" {
				resp.Write(unsafe2.Bytes("There is no this pprof option: " + from + "\n"))
			}
			for i := range options {
				resp.Write(unsafe2.Bytes(options[i]))
			}
		})

	return inited
}
