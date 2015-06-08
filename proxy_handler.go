package main

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/huzorro/spfactor/sexredis"
	"log"
)

const (
	RANKING_PROXY_QUEUE = "ranking:proxy:queue"
)

type ProxyGet struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

type ProxyCheck struct {
	c   *Cfg
	log *log.Logger
	p   sexredis.RedisPool
}

func (self *ProxyGet) SProcess(msg *sexredis.Msg) {
	//
	self.log.Printf("get proxy")
	var (
		proxy ProxyMsg
	)

	proxy.CheckTime = 30
	proxy.City = "北京"
	proxy.Country = "中国"
	proxy.Ip = "192.168.1.111"
	proxy.Isp = "联通"
	proxy.Level = 3
	proxy.Port = "8080"
	proxy.Province = "北京"
	proxy.Type = "https"
	rc, err := self.p.Get()
	defer self.p.Close(rc)

	if err != nil {
		self.log.Printf("get redis connection fails %s", err)
		msg.Err = errors.New("get redis connection fails")
		return
	}

	m, err := json.Marshal(proxy)
	if err != nil {
		self.log.Printf("Marshal json fails %s", err)
		msg.Err = errors.New("Marshal json fails")
		return
	}
	rc.RPush(RANKING_PROXY_QUEUE, m)

}

func (self *ProxyCheck) SProcess(msg *sexredis.Msg) {
	self.log.Printf("check proxy and put on queue")
	//msg type ok?
	m := msg.Content.(ProxyMsg)
	resp, err := HttpGetFromProxy(self.c.CheckApi, "https://"+m.Ip+":"+m.Port)
	if err != nil {
		self.log.Println("proxy check fails %s", err)
		msg.Err = errors.New("proxy check fails")
		return
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		self.log.Println("go query create document fails %s", err)
		msg.Err = errors.New("go query create document fails")
		return
	}
	defer resp.Body.Close()
	if _, exists := doc.Find("#f_email").Attr("name"); !exists {
		self.log.Println("can not get the specified element validation fails")
		msg.Err = errors.New("can not get the specified element validation fails")
		return
	}

	js, err := json.Marshal(m)
	if err != nil {
		self.log.Printf("Marshal json fails %s", err)
		msg.Err = errors.New("Marshal json fails")
		return
	}
	rc, err := self.p.Get()
	defer self.p.Close(rc)
	rc.RPush(RANKING_PROXY_QUEUE, js)
}
