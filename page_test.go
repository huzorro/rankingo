package main

import (
	"testing"
)

func Test_page(t *testing.T) {
	//	d := Details{Detail{Flow_id: 1}, Detail{Flow_id: 2}, Detail{Flow_id: 3}, Detail{Flow_id: 4}, Detail{Flow_id: 5}}
	src := Details{"a", "b", "c"}
	d := make(Details, 1000000)

	ps := 3
	u := d.Len() / ps
	if 1 <= (d.Len() % ps) {
		u++
	}
	//	for i := 1; i <= 4; i++ {
	r := Result{Data: make(Details, ps)}
	d.Page(0, &r)
	t.Errorf("%+v", r)
	//	}
	copy(r.Data, src)

	t.Errorf("%+v", r)
}
