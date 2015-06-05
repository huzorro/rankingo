package main

type Result struct {
	Total        int
	CurrentPage  int
	TotalPage    int
	CurrentTotal int
	Data         Details
}

type Details []interface{}

func (p Details) Len() int {
	return len(p)
}
func (p Details) Page(currentpage int, d *Result) {
	c := d.Data.Len()
	s := p.Len()
	if s == 0 {
		return
	}
	d.Total = s
	cp := s / c
	if cp <= 0 {
		cp = 1
	}
	if 1 <= (s % c) {
		cp++
	}
	d.TotalPage = cp
	if cp < currentpage {
		d.Data = Details{}
	} else {
		if currentpage <= 0 {
			data := p[:c]
			d.Data = data
			d.CurrentPage = currentpage
			d.CurrentTotal = len(data)

		} else {
			m := currentpage * c
			j := m - c
			if m <= s {
				data := p[j:m]
				d.Data = data
				d.CurrentPage = currentpage
				d.CurrentTotal = len(data)
			} else {
				data := p[j:]
				d.Data = data
				d.CurrentPage = currentpage
				d.CurrentTotal = len(data)
			}
		}
	}

}
