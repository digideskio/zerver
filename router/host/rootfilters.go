package host

import (
	"log"
	"net/url"

	"github.com/cosiner/zerver"
)

type RootFilters struct {
	hosts   []string
	filters []zerver.RootFilters
}

func NewRootFilters() *RootFilters {
	return &RootFilters{}
}

func (r *RootFilters) AddRootFilters(host string, rfs zerver.RootFilters) {
	l := len(r.hosts) + 1
	hosts, filters := make([]string, l), make([]zerver.RootFilters, l)
	copy(hosts, r.hosts)
	copy(filters, r.filters)
	hosts[l], filters[l] = host, rfs
	r.hosts, r.filters = hosts, filters
}

func (r *RootFilters) Init(env zerver.Enviroment) error {
	for _, f := range r.filters {
		if e := f.Init(env); e != nil {
			return e
		}
	}
	return nil
}

func (r *RootFilters) Add(interface{}) {
	log.Panicln("Don't add filter to wrapper directly")
}

// Filters return all root filters
func (r *RootFilters) Filters(url *url.URL) []zerver.Filter {
	host, hosts := url.Host, r.hosts
	for i := range hosts {
		if hosts[i] == host {
			return r.filters[i].Filters(url)
		}
	}
	return nil
}

func (r *RootFilters) Destroy() {
	for _, f := range r.filters {
		f.Destroy()
	}
}
