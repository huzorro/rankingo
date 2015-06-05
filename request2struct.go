package main

import (
	"github.com/go-martini/martini"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

type RtoStruct interface {
	Rto(s interface{})
}

type RtoStructer struct {
	r   *http.Request
	log *log.Logger
}

func ToStruct() martini.Handler {
	return func(r *http.Request, log *log.Logger, c martini.Context) {
		c.MapTo(&RtoStructer{r, log}, (*RtoStruct)(nil))
	}
}

func (self *RtoStructer) Rto(s interface{}) {
	var p string
	self.r.ParseForm()
	rType := reflect.TypeOf(s).Elem()
	rValue := reflect.ValueOf(s).Elem()
	for i := 0; i < rType.NumField(); i++ {
		fN := rType.Field(i).Name
		if self.r.Method == "POST" {
			p, _ = url.QueryUnescape(self.r.FormValue(fN))
		} else {
			p, _ = url.QueryUnescape(self.r.PostFormValue(fN))
		}
		switch rType.Field(i).Type.Kind() {
		case reflect.String:
			rValue.FieldByName(fN).SetString(p)
		case reflect.Int64:
			in, _ := strconv.ParseInt(p, 10, 64)
			rValue.FieldByName(fN).SetInt(in)
		case reflect.Float64:
			fl, _ := strconv.ParseFloat(p, 64)
			rValue.FieldByName(fN).SetFloat(fl)
		default:
			log.Printf("unknow type %s", p)
		}
	}
}
