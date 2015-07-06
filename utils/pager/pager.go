package pager

import (
	"github.com/cosiner/gohper/utils/pager"
	"github.com/cosiner/zerver"
)

var pagers *pager.PagerGroup
var PageParam = "page"

func init() {
	pagers = &pager.PagerGroup{}
}

func Add(beginPage, beginIndex, pageSize int) *pager.Pager {
	return pagers.Add(beginPage, beginIndex, pageSize)
}

func Range(req zerver.Request, pager *pager.Pager) (start, count int) {
	return pager.BeginByString(req.Param(PageParam)), pager.PageSize
}
