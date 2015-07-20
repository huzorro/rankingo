package main

import (
	"encoding/json"
	"errors"
	"github.com/huzorro/spfactor/sexredis"
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"time"
)

type AreaProxyKuaidaili struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

type TaskQueuePutIn struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

func (self *AreaProxyKuaidaili) SProcess(msg *sexredis.Msg) {
	self.log.Printf("get proxy of the area from kuaidaili")
	//msg type ok?
	m := msg.Content.(NormMsg)
	var (
		taskMsg   TaskMsg
		proxy     ProxyMsg
		kuaidaili ProxyKuaidaili
		count     int64
	)

	api, err := url.Parse("http://www.kuaidaili.com/api/getproxy")
	if err != nil {
		self.log.Printf("url parse fails %s", err)
		msg.Err = errors.New("url parse fails")
		return
	}
	q := api.Query()
	//订单号
	q.Add("orderid", "")
	//数量
	q.Add("num", "1")
	//稳定性
	q.Add("quality", "1")

	//地区
	if m.KeyMsg.KeyCity != "" {
		q.Add("area", m.KeyMsg.KeyCity)
	} else {
		q.Add("area", m.KeyMsg.KeyProvince)
	}
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

	for {
		if count > 10 {
			q.Set("area", "中国")
		}
		api.RawQuery = q.Encode()
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
		self.log.Printf(string(body))
		if kuaidaili.Code == int64(0) {
			proxy.Ip = strings.Split(kuaidaili.Data.ProxyList[0], ":")[0]
			proxy.Port = strings.Split(kuaidaili.Data.ProxyList[0], ":")[1]
			if result, err := ProxyCheck(self.c.CheckApiSogou, proxy.Ip, proxy.Port, "input[name=_asf]", "value"); !result {
				if result, err = ProxyCheck(self.c.CheckApi360, proxy.Ip, proxy.Port, "#search-button", "value"); result {
					break
				}
				self.log.Printf("check proxy %s", err)

			} else {
				break
			}
		}
		time.Sleep(3000 * time.Millisecond)
		count++
	}

	taskMsg.NormMsg = m
	taskMsg.ProxyMsg = proxy
	msg.Content = taskMsg
}

//验证完成放入任务队列
func (self *TaskQueuePutIn) SProcess(msg *sexredis.Msg) {
	self.log.Printf("check proxy and put on queue of task")
	if msg.Err != nil {
		return
	}
	//msg type ok
	m := msg.Content.(TaskMsg)
	//millis
	m.InitTime = time.Now().UnixNano() / (1000 * 1000)
	js, err := json.Marshal(m)
	if err != nil {
		self.log.Printf("Marshal json fails %s", err)
		msg.Err = errors.New("Marshal json fails")
		return
	}
	rc, err := self.p.Get()
	defer self.p.Close(rc)
	rc.RPush(RANKING_TASK_QUEUE, js)
	msg.Content = m
}
