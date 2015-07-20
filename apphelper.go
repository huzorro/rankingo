package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"html/template"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

var funcMaps = template.FuncMap{
	"html": func(text string) template.HTML {
		return template.HTML(text)
	},
	"loadtimes": func(startTime time.Time) string {
		// 加载时间
		return fmt.Sprintf("%dms", time.Now().Sub(startTime)/1000000)
	},

	"url": func(url string) string {
		// 没有http://或https://开头的增加http://
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			return url
		}

		return "http://" + url
	},
	"add": func(a, b int64) int64 {
		// 加法运算
		return a + b
	},
	"div": func(a, b int64) float64 {
		//除法运算
		return float64(a) / float64(b)
	},
	"formatdate": func(t time.Time) string {
		// 格式化日期
		return t.Format(time.RFC822)
	},
	"nl2br": func(text string) template.HTML {
		return template.HTML(strings.Replace(text, "\n", "<br>", -1))
	},
}

func normByTime(start, end, step float64, sample int) map[int64]float64 {
	var norms []float64
	for i := 0; i < sample; i++ {
		q := rand.NormFloat64()
		if q >= start && q <= end {
			norms = append(norms, q)
		}
	}
	sort.Float64s(norms)
	timesMap := make(map[int64]float64)
	for i := start; i < end; i = i + step {
		qf := i*10 + 12
		qf = math.Floor(qf)
		var ss []float64
		for _, j := range norms {
			if j >= i && j <= i+step {
				ss = append(ss, j)
			}
		}
		timesMap[int64(qf)] = float64(len(ss)) / float64(len(norms))
	}
	return timesMap
}
func ProxyCheck(vurl, ip, port, vflag, vattr string) (bool, error) {
	resp, err := HttpGetFromProxy(vurl, fmt.Sprintf("https://%s:%s", ip, port))
	if err != nil {
		return false, err
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if _, exists := doc.Find(vflag).Attr(vattr); !exists {
		return false, errors.New("Not foud")
	}
	return true, nil
}

//func (self *ProxyCheckOschina) SProcess(msg *sexredis.Msg) {
//	self.log.Printf("check proxy and put on queue")
//	if msg.Err != nil {
//		return
//	}

//	//msg type ok?
//	m := msg.Content.(ProxyMsg)
//	resp, err := HttpGetFromProxy(self.c.CheckApiOschina, "https://"+m.Ip+":"+m.Port)
//	if err != nil {
//		self.log.Println("proxy check fails %s", err)
//		msg.Err = errors.New("proxy check fails")
//		return
//	}
//	doc, err := goquery.NewDocumentFromResponse(resp)
//	if err != nil {
//		self.log.Println("go query create document fails %s", err)
//		msg.Err = errors.New("go query create document fails")
//		return
//	}
//	defer resp.Body.Close()
//	if _, exists := doc.Find("#f_email").Attr("name"); !exists {
//		self.log.Println("can not get the specified element validation fails")
//		msg.Err = errors.New("can not get the specified element validation fails")
//		return
//	}

//	msg.Content = m
//}

//func (self *ProxyCheckSogou) SProcess(msg *sexredis.Msg) {
//	self.log.Printf("check proxy and put on queue")
//	if msg.Err != nil {
//		return
//	}
//	//msg type ok?
//	m := msg.Content.(ProxyMsg)
//	resp, err := HttpGetFromProxy(self.c.CheckApiSogou, "https://"+m.Ip+":"+m.Port)
//	if err != nil {
//		self.log.Println("proxy check fails %s", err)
//		msg.Err = errors.New("proxy check fails")
//		return
//	}
//	doc, err := goquery.NewDocumentFromResponse(resp)
//	if err != nil {
//		self.log.Println("go query create document fails %s", err)
//		msg.Err = errors.New("go query create document fails")
//		return
//	}
//	defer resp.Body.Close()
//	if _, exists := doc.Find("input[name=_asf]").Attr("value"); !exists {
//		self.log.Println("can not get the specified element validation fails")
//		msg.Err = errors.New("can not get the specified element validation fails")
//		return
//	}

//	msg.Content = m
//}

//func (self *ProxyCheck360) SProcess(msg *sexredis.Msg) {
//	self.log.Printf("check proxy and put on queue")
//	if msg.Err != nil {
//		return
//	}
//	//msg type ok?
//	m := msg.Content.(ProxyMsg)
//	resp, err := HttpGetFromProxy(self.c.CheckApi360, "https://"+m.Ip+":"+m.Port)
//	if err != nil {
//		self.log.Println("proxy check fails %s", err)
//		msg.Err = errors.New("proxy check fails")
//		return
//	}
//	doc, err := goquery.NewDocumentFromResponse(resp)
//	if err != nil {
//		self.log.Println("go query create document fails %s", err)
//		msg.Err = errors.New("go query create document fails")
//		return
//	}
//	defer resp.Body.Close()
//	if _, exists := doc.Find("#search-button").Attr("value"); !exists {
//		self.log.Println("can not get the specified element validation fails")
//		msg.Err = errors.New("can not get the specified element validation fails")
//		return
//	}

//	msg.Content = m
//}
