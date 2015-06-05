package main

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
)

type Paginator struct {
	Request     *http.Request
	PerPageNums int64
	MaxPages    int64

	nums      int64
	pageRange []int64
	pageNums  int64
	page      int64
}

func (p *Paginator) PageNums() int64 {
	if p.pageNums != 0 {
		return p.pageNums
	}
	pageNums := math.Ceil(float64(p.nums) / float64(p.PerPageNums))
	if p.MaxPages > 0 {
		pageNums = math.Min(pageNums, float64(p.MaxPages))
	}
	p.pageNums = int64(pageNums)
	return p.pageNums
}

func (p *Paginator) Nums() int64 {
	return p.nums
}

func (p *Paginator) SetNums(nums interface{}) {
	p.nums, _ = nums.(int64)
}

func (p *Paginator) Page() int64 {
	if p.page != 0 {
		return p.page
	}
	if p.Request.Form == nil {
		p.Request.ParseForm()
	}
	p.page, _ = strconv.ParseInt(p.Request.Form.Get("p"), 10, 64)
	if p.page > p.PageNums() {
		p.page = p.PageNums()
	}
	if p.page <= 0 {
		p.page = 1
	}
	return p.page
}

func (p *Paginator) Pages() []int64 {
	if p.pageRange == nil && p.nums > 0 {
		var pages []int64
		pageNums := p.PageNums()
		page := p.Page()
		switch {
		case page >= pageNums-4 && pageNums > 9:
			start := pageNums - 9 + 1
			pages = make([]int64, 9)
			for i, _ := range pages {
				pages[i] = start + int64(i)
			}
		case page >= 5 && pageNums > 9:
			start := page - 5 + 1
			pages = make([]int64, int64(math.Min(9, float64(page+4+1))))
			for i, _ := range pages {
				pages[i] = start + int64(i)
			}
		default:
			pages = make([]int64, int64(math.Min(9, float64(pageNums))))
			for i, _ := range pages {
				pages[i] = int64(i) + 1
			}
		}
		p.pageRange = pages
	}
	return p.pageRange
}

func (p *Paginator) PageLink(page int64) string {
	link, _ := url.ParseRequestURI(p.Request.RequestURI)
	values := link.Query()
	if page == 1 {
		values.Del("p")
	} else {
		values.Set("p", strconv.FormatInt(page, 10))
	}
	link.RawQuery = values.Encode()
	return link.String()
}

func (p *Paginator) PageLinkPrev() (link string) {
	if p.HasPrev() {
		link = p.PageLink(p.Page() - int64(1))
	}
	return
}

func (p *Paginator) PageLinkNext() (link string) {
	if p.HasNext() {
		link = p.PageLink(p.Page() + 1)
	}
	return
}

func (p *Paginator) PageLinkFirst() (link string) {
	return p.PageLink(1)
}

func (p *Paginator) PageLinkLast() (link string) {
	return p.PageLink(p.PageNums())
}

func (p *Paginator) HasPrev() bool {
	return p.Page() > 1
}

func (p *Paginator) HasNext() bool {
	return p.Page() < p.PageNums()
}

func (p *Paginator) IsActive(page int64) bool {
	return p.Page() == page
}

func (p *Paginator) Offset() int64 {
	return (p.Page() - 1) * p.PerPageNums
}

func (p *Paginator) HasPages() bool {
	return p.PageNums() > 1
}

func NewPaginator(req *http.Request, per int64, nums interface{}) *Paginator {
	p := Paginator{}
	p.Request = req
	if per <= 0 {
		per = 10
	}
	p.PerPageNums = per
	p.SetNums(nums)
	return &p
}
