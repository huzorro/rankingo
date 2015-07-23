package main

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/huzorro/spfactor/sexredis"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"
)

const (
	RANKING_PROXY_QUEUE = "ranking:proxy:queue"
)

type ProxyGetKuaidaili struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

type ProxyCheckOschina struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

type ProxyCheckSogou struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

type ProxyCheck360 struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

type ProxyQueuePutIn struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

//正常返回格式样例：
//{"msg": "", "code": 0, "data": {"count": 10, "proxy_list": ["112.95.241.76:80"]}}

//异常返回格式样例：
//{"msg": "参数错误", "code": -1, "data": ""}

type ProxyKuaidaili struct {
	Msg  string `json:"msg"`
	Code int64  `json:"code"`
	Data struct {
		Count     int64    `json:"count"`
		ProxyList []string `json:"proxy_list"`
	} `json:"data"`
}

func (self *ProxyGetKuaidaili) SProcess(msg *sexredis.Msg) {
	self.log.Printf("get proxy for kuaidaili")
	var (
		proxy     ProxyMsg
		kuaidaili ProxyKuaidaili
	)

	api, err := url.Parse("http://www.kuaidaili.com/api/getproxy")
	if err != nil {
		self.log.Printf("url parse fails %s", err)
		msg.Err = errors.New("url parse fails")
		return
	}
	q := api.Query()
	//订单号
	q.Add("orderid", "953736356849843")
	//数量
	q.Add("num", "1")
	//稳定性
	q.Add("quality", "1")
	//地区
	q.Add("area", "中国")
	//协议 2:https
	q.Add("protocol", "2")
	//请求方式 get
	q.Add("method", "1")
	//游览器ua支持 pc
	q.Add("browser", "1")
	//高匿名
	q.Add("an_ha", "1")
	//结果返回格式
	q.Add("format", "json")
	api.RawQuery = q.Encode()
	for {
		resp, err := HttpGet(api.String())
		defer resp.Body.Close()
		if err != nil {
			self.log.Printf("proxy kuaidaili get fails %s", err)
			msg.Err = errors.New("proxy kuaidaili get fails %s")
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &kuaidaili)

		if err != nil {
			self.log.Printf("Unmarshal fails %s", err)
			msg.Err = errors.New("Unmarshal json fails")
			return
		}
		if kuaidaili.Code == int64(0) {
			break
		}
		time.Sleep(3000 * time.Millisecond)
	}

	proxy.Ip = strings.Split(kuaidaili.Data.ProxyList[0], ":")[0]
	proxy.Port = strings.Split(kuaidaili.Data.ProxyList[0], ":")[1]
	msg.Content = proxy
}

func (self *ProxyCheckOschina) SProcess(msg *sexredis.Msg) {
	self.log.Printf("check proxy and put on queue")
	if msg.Err != nil {
		return
	}

	//msg type ok?
	m := msg.Content.(ProxyMsg)
	resp, err := HttpGetFromProxy(self.c.CheckApiOschina, "https://"+m.Ip+":"+m.Port)
	if err != nil {
		self.log.Printf("proxy check fails %s", err)
		msg.Err = errors.New("proxy check fails")
		return
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		self.log.Printf("go query create document fails %s", err)
		msg.Err = errors.New("go query create document fails")
		return
	}
	defer resp.Body.Close()
	if _, exists := doc.Find("#f_email").Attr("name"); !exists {
		self.log.Printf("can not get the specified element validation fails")
		msg.Err = errors.New("can not get the specified element validation fails")
		return
	}

	msg.Content = m
}

func (self *ProxyCheckSogou) SProcess(msg *sexredis.Msg) {
	self.log.Printf("check proxy and put on queue")
	if msg.Err != nil {
		return
	}
	//msg type ok?
	m := msg.Content.(ProxyMsg)
	resp, err := HttpGetFromProxy(self.c.CheckApiSogou, "https://"+m.Ip+":"+m.Port)
	if err != nil {
		self.log.Printf("proxy check fails %s", err)
		msg.Err = errors.New("proxy check fails")
		return
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		self.log.Printf("go query create document fails %s", err)
		msg.Err = errors.New("go query create document fails")
		return
	}
	defer resp.Body.Close()
	if _, exists := doc.Find("input[name=_asf]").Attr("value"); !exists {
		self.log.Printf("can not get the specified element validation fails")
		msg.Err = errors.New("can not get the specified element validation fails")
		return
	}

	msg.Content = m
}

func (self *ProxyCheck360) SProcess(msg *sexredis.Msg) {
	self.log.Printf("check proxy and put on queue")
	if msg.Err != nil {
		return
	}
	//msg type ok?
	m := msg.Content.(ProxyMsg)
	resp, err := HttpGetFromProxy(self.c.CheckApi360, "https://"+m.Ip+":"+m.Port)
	if err != nil {
		self.log.Printf("proxy check fails %s", err)
		msg.Err = errors.New("proxy check fails")
		return
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		self.log.Printf("go query create document fails %s", err)
		msg.Err = errors.New("go query create document fails")
		return
	}
	defer resp.Body.Close()
	if _, exists := doc.Find("#search-button").Attr("value"); !exists {
		self.log.Printf("can not get the specified element validation fails")
		msg.Err = errors.New("can not get the specified element validation fails")
		return
	}

	msg.Content = m
}

func (self *ProxyQueuePutIn) SProcess(msg *sexredis.Msg) {
	self.log.Printf("check proxy and put on queue")
	if msg.Err != nil {
		return
	}
	//msg type ok
	m := msg.Content.(ProxyMsg)
	js, err := json.Marshal(m)
	if err != nil {
		self.log.Printf("Marshal json fails %s", err)
		msg.Err = errors.New("Marshal json fails")
		return
	}
	rc, err := self.p.Get()
	defer self.p.Close(rc)
	rc.RPush(RANKING_PROXY_QUEUE, js)
	msg.Content = m
}
