// Package pprof provide a simple monitoring interface for zerver, all monitor is
// handled GET request
// use AddMonitorHandler to add a monitor, it should be called before EnableMonitoring
// for there is only one change to init
package pprof

import (
	"net/http"
	"net/url"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/cosiner/gohper/lib/sys"
	"github.com/cosiner/gohper/lib/types"
	"github.com/cosiner/zerver"
)

var path = "/status"

func EnableMonitoring(p string, rt zerver.Router, rootFilters zerver.RootFilters) (err error) {
	if p != "" {
		path = p
	}
	if !initRoutes() {
		return
	}
	for subpath, handler := range routes {
		if err = rt.AddFuncHandler(path+subpath, "GET", handler); err != nil {
			return
		}
		options = append(options, "GET "+path+subpath+": "+infos[subpath]+"\n")
	}
	if rootFilters == nil {
		err = rt.AddFuncFilter(path, globalFilter)
	} else {
		rootFilters.AddFuncFilter(globalFilter)
	}
	return
}

func NewMonitorServer(p string) (*zerver.Server, error) {
	if p == "" {
		p = "/"
	}
	s := zerver.NewServer()
	// s.AddFuncHandler("/stop", "GET", func(req zerver.Request, resp zerver.Response) {
	// 	req.Server().Destroy()
	// })
	// infos["/stop"] = "stop pprof server"
	return s, EnableMonitoring("/", s, s.RootFilters)
}

func AddMonitorHandler(info, path string, handler zerver.HandlerFunc) {
	infos[path], routes[path] = info, handler
}

var options = make([]string, 0, len(infos)+1)
var routes = make(map[string]zerver.HandlerFunc)
var infos = make(map[string]string)

func pprofLookupHandler(name string) zerver.HandlerFunc {
	return func(req zerver.Request, resp zerver.Response) {
		pprof.Lookup(name).WriteTo(resp, 2)
	}
}

func globalFilter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	resp.SetContentType("text/plain")
	if resp.Status() == http.StatusNotFound {
		resp.SetHeader("Location", path+"/options?from="+url.QueryEscape(req.URL().Path))
		resp.ReportMovedPermanently()
	} else if resp.Status() == http.StatusMethodNotAllowed {
		sys.WriteStrln(resp, "The pprof interface only support GET request")
	} else {
		chain(req, resp)
	}
}

var inited bool

func initRoutes() bool {
	if inited {
		return false
	}
	inited = true

	infos["/goroutine"] = "Get goroutine info"
	routes["/goroutine"] = pprofLookupHandler("goroutine")
	infos["/heap"] = "Get heap info"
	routes["/heap"] = pprofLookupHandler("heap")
	infos["/thread"] = "Get thread create info"
	routes["/thread"] = pprofLookupHandler("threadcreate")
	infos["/block"] = "Get block info"
	routes["/block"] = pprofLookupHandler("block")

	infos["/cpu"] = "Get CPU info, default seconds is 30, use ?seconds= to reset"
	routes["/cpu"] = func(req zerver.Request, resp zerver.Response) {
		var t int
		if secs := req.Param("seconds"); secs != "" {
			var err error
			if t, err = strconv.Atoi(secs); err != nil {
				resp.ReportBadRequest()
				sys.WriteStrln(resp, secs+" is not a integer number")
				return
			}
		}
		if t <= 0 {
			t = 30
		}
		pprof.StartCPUProfile(resp)
		time.Sleep(time.Duration(t) * time.Second)
		pprof.StopCPUProfile()
	}

	infos["/memory"] = "Get memory info"
	routes["/memory"] = func(req zerver.Request, resp zerver.Response) {
		runtime.GC()
		pprof.WriteHeapProfile(resp)
	}

	infos["/routes"] = "Get all routes"
	routes["/routes"] = func(req zerver.Request, resp zerver.Response) {
		req.Server().PrintRouteTree(resp)
	}

	infos["/statistic"] = "Get server statistic info such as uptime"
	infos["/options"] = "Get all pprof options"
	routes["/options"] = func(req zerver.Request, resp zerver.Response) {
		if from := req.Param("from"); from != "" {
			resp.Write(types.UnsafeBytes("There is no this pprof option: " + from + "\n"))
		}
		for i := range options {
			resp.Write(types.UnsafeBytes(options[i]))
		}
	}

	return inited
}
