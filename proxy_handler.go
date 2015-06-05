package main

import (
	"encoding/json"
	"errors"
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
	//
	self.log.Printf("check proxy and put on queue")
}
