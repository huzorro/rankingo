package main

import (
	"encoding/json"
	"errors"
	"github.com/huzorro/spfactor/sexredis"
	"log"
	"reflect"
	"time"
)

type NormMsg struct {
	KeyMsg KeyMsg           `json:"keyMsg"`
	COrder int64            `json:"corder"`
	CIndex int64            `json:"cindex"`
	HOrder int64            `json:"horder"`
	HIndex int64            `json:"hindex"`
	Cost   int64            `json:"cost"`
	Hour   map[string]int64 `json:"hour, omitempty"`
	Cancel bool             `json:"cancel, omitempty"`
}

type ProxyMsg struct {
	//对于使用adsl的服务器, ip设置为服务器ip
	Ip string `json:"ip"`
	//对于使用adsl的服务器, port设置为adsl拨号成功后的出口ip的详细信息
	Port string `json:"port"`
}

type TaskMsg struct {
	InitTime int64    `json:"initTime"`
	NormMsg  NormMsg  `json:"norm"`
	ProxyMsg ProxyMsg `json:"proxy"`
}
type Task struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

func (self *Task) SProcess(msg *sexredis.Msg) {
	self.log.Printf("norm handler %+v", msg)
	//msg type ok?
	var (
		taskMsg  TaskMsg
		normMsg  NormMsg
		proxyMsg ProxyMsg
		js       []byte
	)
	msgs := make([]string, 0)
	val := reflect.ValueOf(msg.Content)
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			e := val.Index(i)
			m := e.Interface().(sexredis.Msg)
			msgs = append(msgs, m.Content.(string))
			//			switch m.Content.(type) {
			//			case NormMsg:
			//				taskMsg.NormMsg = m.Content.(NormMsg)
			//			case ProxyMsg:
			//				taskMsg.ProxyMsg = m.Content.(ProxyMsg)
			//			default:
			//				msg.Err = errors.New("invalid kind")
			//				return
			//			}
		}
	}
	if err := json.Unmarshal([]byte(msgs[0]), &normMsg); err != nil {
		self.log.Printf("Unmarshal json fails %s", err)
		msg.Err = errors.New("get redis connection fails")
		return
	}
	if err := json.Unmarshal([]byte(msgs[1]), &proxyMsg); err != nil {
		self.log.Printf("Unmarshal json fails %s", err)
		msg.Err = errors.New("get redis connection fails")
		return
	}
	taskMsg.NormMsg = normMsg
	taskMsg.ProxyMsg = proxyMsg
	rc, err := self.p.Get()
	defer self.p.Close(rc)
	if err != nil {
		self.log.Printf("get redis connection fails %s", err)
		msg.Err = errors.New("get redis connection fails")
		return
	}
	//millis
	taskMsg.InitTime = time.Now().UnixNano() / (1000 * 1000)
	if js, err = json.Marshal(taskMsg); err != nil {
		self.log.Printf("Marshal json fails %s", err)
		msg.Err = errors.New("Marshal json fails")
		return
	}
	if _, err := rc.RPush(RANKING_TASK_QUEUE, string(js)); err != nil {
		self.log.Printf("put msg in queue fails %s", err)
		msg.Err = errors.New("put msg in queue fails")
		return
	}
	msg.Content = taskMsg
}
