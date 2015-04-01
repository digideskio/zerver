package routers

import (
	"net/url"

	. "github.com/cosiner/zerver"
)

type HostRootFilters struct {
	hosts   []string
	filters []RootFilters
}

// Implement RootFilters

func (hr HostRootFilters) AddRootFilters(host string, rfs RootFilters) {
	l := len(hr.hosts) + 1
	hosts, filters := make([]string, l), make([]RootFilters, l)
	copy(hosts, hr.hosts)
	copy(filters, hr.filters)
	hosts[l], filters[l] = host, rfs
	hr.hosts, hr.filters = hosts, filters
}

func (hr HostRootFilters) Init(s *Server) error {
	for _, f := range hr.filters {
		if e := f.Init(s); e != nil {
			return e
		}
	}
	return nil
}

func (hr HostRootFilters) Add(interface{}) {
	panic("Don't add filter to wrapper directly")
}

// Filters return all root filters
func (hr HostRootFilters) Filters(url *url.URL) []Filter {
	host, hosts := url.Host, hr.hosts
	for i := range hosts {
		if hosts[i] == host {
			return hr.filters[i].Filters(url)
		}
	}
	return nil
}

func (hr HostRootFilters) Destroy() {
	for _, f := range hr.filters {
		f.Destroy()
	}
}
