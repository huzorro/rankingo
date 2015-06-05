package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/huzorro/spfactor/sexredis"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	RANKING_COMMON_NORM_QUEUE = "ranking:common:norm:queue"
	RANKING_AREA_NORM_QUEUE   = "ranking:area:norm:queue"
	RANKING_NORM_QUEUE        = "ranking:norm:queue"
)

type Order struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

type Index struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

type PutIn struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

type Recoder struct {
	c   *Cfg
	log *log.Logger
	db  *sql.DB
}

//计算价格
type Cost struct {
	c   *Cfg
	log *log.Logger
	db  *sql.DB
}

//计费
type Payoff struct {
	c   *Cfg
	log *log.Logger
	db  *sql.DB
}

//优化日志
type OrderLog struct {
	c   *Cfg
	log *log.Logger
	db  *sql.DB
}

type OrderResult struct {
	State int64  `json:"state"`
	Msg   string `json:"msg"`
	Data  struct {
		Host    string `json:"host"`
		Keyword string `json:"keyword"`
		Sort    string `json:"sort"`
	} `json:"data"`
}

type IndexResult struct {
	State int64  `json:"state"`
	Msg   string `json:"msg"`
	Data  []struct {
		Keyword     string `json:"keyword"`
		Allindex    int64  `json:"allindex"`
		Mobileindex int64  `json:"mobileindex"`
		So360index  int64  `json:"so360index"`
	} `json:"data"`
}

func (self *Order) SProcess(msg *sexredis.Msg) {
	self.log.Printf("get order .....")
	//msg type ok?
	var (
		keymsg KeyMsg
		normsg NormMsg
		or     OrderResult
	)
	m := msg.Content.(string)
	if err := json.Unmarshal([]byte(m), &keymsg); err != nil {
		self.log.Printf("Unmarshal json fails %s", err)
		msg.Err = errors.New("Unmarshal json fails")
		return
	}
	urlValues := url.Values{}
	urlValues.Add("key", self.c.OrderApiKey)
	urlValues.Add("host", keymsg.Destlink)
	urlValues.Add("wd", keymsg.Keyword)
	query := urlValues.Encode()
	self.log.Printf("%s", self.c.OrderApi+"?"+query)
	for {
		resp, err := HttpGet(self.c.OrderApi + "?" + query)
		defer resp.Body.Close()
		if err != nil {
			msg.Err = errors.New("get order fails")
			self.log.Printf("get order fails %s", err)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		self.log.Printf("%s", string(body))
		if err := json.Unmarshal(body, &or); err != nil {
			msg.Err = errors.New("json Unmarshal fails")
			self.log.Printf("json Unmarshal fails %s", err)
			return
		}
		if or.State == int64(1) {
			break
		}
		time.Sleep(3000 * time.Millisecond)
	}
	cos := strings.Split(or.Data.Sort, ",")
	co, err := strconv.ParseInt(cos[0], 10, 64)
	if err != nil {
		if co, err = strconv.ParseInt(string([]rune(cos[0])[:3]), 10, 64); err != nil {
			if co, err = strconv.ParseInt(string([]rune(cos[0])[:2]), 10, 64); err != nil {
			}
		}
	}
	normsg.COrder = co
	normsg.KeyMsg = keymsg
	msg.Content = normsg
}

func (self *Index) SProcess(msg *sexredis.Msg) {
	self.log.Printf("get index .....")
	var (
		ir IndexResult
	)
	//msg type ok?
	m := msg.Content.(NormMsg)
	urlValues := url.Values{}
	urlValues.Add("key", self.c.IndexApiKey)
	urlValues.Add("kws", m.KeyMsg.Keyword)
	query := urlValues.Encode()
	self.log.Printf("%s", self.c.IndexApi+"?"+query)
	for {
		resp, err := HttpGet(self.c.IndexApi + "?" + query)
		defer resp.Body.Close()
		if err != nil {
			msg.Err = errors.New("get index fails")
			self.log.Printf("get index fails %s", err)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		self.log.Printf("%s", string(body))
		if err := json.Unmarshal(body, &ir); err != nil {
			msg.Err = errors.New("json Unmarshal fails")
			self.log.Printf("json Unmarshal fails %s", err)
			return
		}
		if ir.State == 1 {
			break
		}
		time.Sleep(3000 * time.Millisecond)
	}
	self.log.Printf("%+v", ir)
	m.CIndex = ir.Data[0].Allindex
	msg.Content = m
}

func (self *PutIn) SProcess(msg *sexredis.Msg) {
	var (
		js []byte
	)
	//msg type ok?
	m := msg.Content.(NormMsg)
	rc, err := self.p.Get()
	defer self.p.Close(rc)
	if err != nil {
		self.log.Printf("get redis connection fails %s", err)
		msg.Err = errors.New("get redis connection fails")
		return
	}
	if js, err = json.Marshal(m); err != nil {
		self.log.Printf("Marshal json fails %s", err)
		msg.Err = errors.New("Marshal json fails")
		return
	}
	if m.KeyMsg.Status == RANKING_STATUS_CANCEL {
		self.log.Printf("the keyword is cancel and not put in")
		return
	}
	if _, err := rc.RPush(RANKING_NORM_QUEUE, js); err != nil {
		self.log.Printf("put msg in queue fails %s", err)
		msg.Err = errors.New("put msg in queue fails")
		return
	}
	self.log.Printf("put msg in %s", RANKING_NORM_QUEUE)
}

func (self *Recoder) SProcess(msg *sexredis.Msg) {
	//msg type ok?
	m := msg.Content.(NormMsg)
	stmtIn, err := self.db.Prepare(`REPLACE INTO ranking_detail(id, uid, owner, keyword, destlink, history_order, 
	current_order, history_index, current_index, city_key, province_key, cost, status, logtime) 
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	defer stmtIn.Close()
	if err != nil {
		self.log.Printf("db.Prepare fails %s", err)
		msg.Err = errors.New("db.Prepare fails")
		return
	}
	if _, err := stmtIn.Exec(m.KeyMsg.Id, m.KeyMsg.Uid, m.KeyMsg.Owner, m.KeyMsg.Keyword, m.KeyMsg.Destlink, m.COrder, m.COrder,
		m.CIndex, m.CIndex, m.KeyMsg.KeyCity, m.KeyMsg.KeyProvince, m.Cost, m.KeyMsg.Status, m.KeyMsg.Logtime); err != nil {
		self.log.Printf("stmtIn.Exec fails %s", err)
		msg.Err = errors.New("stmtIn.Exec fails")
		return
	}
	self.log.Printf("db recoder in %s", "ranking_detail")
}

func (self *OrderLog) SProcess(msg *sexredis.Msg) {
	//msg type ok?
	m := msg.Content.(NormMsg)
	if m.KeyMsg.Status == RANKING_STATUS_CANCEL {
		self.log.Printf("the keyword is cancel and not log in")
		return
	}
	stmtIn, err := self.db.Prepare(`INSERT INTO ranking_log(id, uid, owner, keyword, destlink, history_order, 
	current_order, history_index, current_index, city_key, province_key, cost, status, logtime) 
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	defer stmtIn.Close()
	if err != nil {
		self.log.Printf("db.Prepare fails %s", err)
		msg.Err = errors.New("db.Prepare fails")
		return
	}
	if _, err := stmtIn.Exec(m.KeyMsg.Id, m.KeyMsg.Uid, m.KeyMsg.Owner, m.KeyMsg.Keyword, m.KeyMsg.Destlink, m.COrder, m.COrder,
		m.CIndex, m.CIndex, m.KeyMsg.KeyCity, m.KeyMsg.KeyProvince, m.Cost, m.KeyMsg.Logtime); err != nil {
		self.log.Printf("stmtIn.Exec fails %s", err)
		msg.Err = errors.New("stmtIn.Exec fails")
		return
	}
	self.log.Printf("db order log in %s", "ranking_log")
}
