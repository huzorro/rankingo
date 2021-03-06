package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	//"github.com/djimenez/iconv-go"
	"github.com/huzorro/spfactor/sexredis"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
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
	db  *sql.DB
}

type Index struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
	db  *sql.DB
}

//放入norm队列
type PutIn struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

//放入task队列
type PutInTask struct {
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

//正态分布
type NormCreate struct {
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
type Payment struct {
	c   *Cfg
	log *log.Logger
	db  *sql.DB
}

type VerifyKey struct {
	c   *Cfg
	log *log.Logger
	db  *sql.DB
}

//校验目标页面是否存在指定关键字
func (self VerifyKey) SProcess(msg *sexredis.Msg) {
	self.log.Printf("verify keyword is valid?")
	var keymsg KeyMsg
	//msg type ok?
	m := msg.Content.(string)
	if err := json.Unmarshal([]byte(m), &keymsg); err != nil {
		self.log.Printf("Unmarshal json fails %s", err)
		msg.Err = errors.New("Unmarshal json fails")
		return
	}
	resp, err := HttpGet("http://" + keymsg.Destlink)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		self.log.Printf("verify keyword fails %s", err)
		msg.Err = errors.New("verify keyword is fails")
		return
	}
	doc, err := goquery.NewDocumentFromResponse(resp)

	if err != nil {
		self.log.Printf("get dest link page  document fails %s", err)
		msg.Err = errors.New("get dest link page  document fails")
		return
	}

	doc.Find("meta").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if c, f := s.Attr("name"); f {
			if strings.Contains(strings.ToLower(c), "keywords") {
				if keywords, f := s.Attr("content"); f {
					utf8, _ := GbkToUtf8([]byte(keywords))
					//utf8, _ := iconv.ConvertString(keywords, "gbk", "utf-8")
					self.log.Printf("%s, %s", keywords, string(utf8))
					g := strings.Contains(keywords, keymsg.Keyword)
					//u := strings.Contains(utf8, keymsg.Keyword)
					u := strings.Contains(string(utf8), keymsg.Keyword)
					if g || u {
						self.log.Printf("%s in %s found", keymsg.Keyword, keymsg.Destlink)
						msg.Content = keymsg
						msg.Err = nil
						return false
					} else {
						self.log.Printf("%s in %s not found", keymsg.Keyword, keymsg.Destlink)
						msg.Err = errors.New("verify keyword fails")
						return true
					}
				} else {
					self.log.Printf("%s in %s not found", keymsg.Keyword, keymsg.Destlink)
					msg.Err = errors.New("verify keyword fails")
					return true
				}
			} else {
				self.log.Printf("%s in %s not found", keymsg.Keyword, keymsg.Destlink)
				msg.Err = errors.New("verify keyword fails")
				return true
			}
		} else {
			self.log.Printf("%s in %s not found", keymsg.Keyword, keymsg.Destlink)
			msg.Err = errors.New("verify keyword fails")
			return true
		}
	})
	//	if text := doc.Find("head").Text(); text != "" {
	//		utf8, _ := iconv.ConvertString(text, "gbk", "utf-8")
	//		self.log.Printf("%s, %s", text, utf8)
	//		f := strings.Contains(text, keymsg.Keyword)
	//		u := strings.Contains(utf8, keymsg.Keyword)
	//		if f || u {
	//			self.log.Printf("%s in %s found", keymsg.Keyword, keymsg.Destlink)
	//			msg.Content = keymsg
	//		} else {
	//			self.log.Printf("%s in %s not found", keymsg.Keyword, keymsg.Destlink)
	//			msg.Err = errors.New("verify keyword fails")
	//			return
	//		}
	//	}
}

func (self *Order) SProcess(msg *sexredis.Msg) {
	self.log.Printf("get order .....")
	//msg type ok?
	var (
		keymsg KeyMsg
		normsg NormMsg
		or     OrderResult
		count  int
	)
	//	m := msg.Content.(string)
	//	if err := json.Unmarshal([]byte(m), &keymsg); err != nil {
	//		self.log.Printf("Unmarshal json fails %s", err)
	//		msg.Err = errors.New("Unmarshal json fails")
	//		return
	//	}
	keymsg = msg.Content.(KeyMsg)
	urlValues := url.Values{}
	urlValues.Add("key", self.c.OrderApiKey)
	urlValues.Add("host", keymsg.Destlink)
	urlValues.Add("wd", keymsg.Keyword)
	query := urlValues.Encode()
	self.log.Printf("%s", self.c.OrderApi+"?"+query)
	for {
		resp, err := HttpGet(self.c.OrderApi + "?" + query)
		defer func() {
			if resp != nil {
				resp.Body.Close()
			}
		}()
		if err != nil {
			//			msg.Err = errors.New("get order fails")
			self.log.Printf("get order fails %s", err)
			time.Sleep(3000 * time.Millisecond)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		self.log.Printf("%s", string(body))
		if err := json.Unmarshal(body, &or); err != nil {
			msg.Err = errors.New("json Unmarshal fails")
			self.log.Printf("json Unmarshal fails %s", err)
			self.log.Printf("use ranking_detail of result")
			stmtOut, err := self.db.Prepare(`SELECT current_order FROM ranking_detail WHERE keyword = ? AND destlink = ?`)
			defer stmtOut.Close()
			if err != nil {
				self.log.Printf("db prepare fails %s", err)
				msg.Err = errors.New("db prepare fails")
				return
			}
			row := stmtOut.QueryRow(keymsg.Keyword, keymsg.Destlink)
			order := 50
			if err := row.Scan(&order); err != nil {
				self.log.Printf("row scan fails or result is zero")
			}
			normsg.COrder = int64(order)
			normsg.KeyMsg = keymsg
			msg.Content = normsg
			return
		}
		if or.State == int64(1) {
			break
		}
		time.Sleep(3000 * time.Millisecond)
		//如果连续5次没有从api取到结果
		//那么从ranking_detail表取结果
		//如果ranking_detail表仍然没有取到结果
		//使用一个缺省的排名和指数order = 50 index = 200
		//等待次日任务运行时从api获取最新数据
		if count += 1; count >= 5 {
			self.log.Printf("use ranking_detail of result")
			stmtOut, err := self.db.Prepare(`SELECT current_order FROM ranking_detail WHERE keyword = ? AND destlink = ?`)
			defer stmtOut.Close()
			if err != nil {
				self.log.Printf("db prepare fails %s", err)
				msg.Err = errors.New("db prepare fails")
				return
			}
			row := stmtOut.QueryRow(keymsg.Keyword, keymsg.Destlink)
			order := 50
			if err := row.Scan(&order); err != nil {
				self.log.Printf("row scan fails or result is zero")
			}
			normsg.COrder = int64(order)
			normsg.KeyMsg = keymsg
			msg.Content = normsg
			return
		}

	}
	cos := strings.Split(or.Data.Sort, ",")
	co, err := strconv.ParseInt(cos[0], 10, 64)
	if err != nil {
		if co, err = strconv.ParseInt(string([]rune(cos[0])[:3]), 10, 64); err != nil {
			if co, err = strconv.ParseInt(string([]rune(cos[0])[:2]), 10, 64); err != nil {
			}
		}
	}
	if normsg.HOrder == 0 {
		normsg.HOrder = co
	}
	normsg.COrder = co
	normsg.KeyMsg = keymsg
	msg.Content = normsg
}

func (self *Index) SProcess(msg *sexredis.Msg) {
	self.log.Printf("get index .....")
	var (
		ir    IndexResult
		count int
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
		defer func() {
			if resp != nil {
				resp.Body.Close()
			}
		}()
		if err != nil {
			//			msg.Err = errors.New("get index fails")
			self.log.Printf("get index fails %s", err)
			time.Sleep(3000 * time.Millisecond)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		self.log.Printf("%s", string(body))
		if err := json.Unmarshal(body, &ir); err != nil {
			msg.Err = errors.New("json Unmarshal fails")
			self.log.Printf("json Unmarshal fails %s", err)
			self.log.Printf("use ranking_detail of result")
			stmtOut, err := self.db.Prepare(`SELECT current_index FROM ranking_detail WHERE keyword = ? AND destlink = ?`)
			defer stmtOut.Close()
			if err != nil {
				self.log.Printf("db prepare fails %s", err)
				msg.Err = errors.New("db prepare fails")
				return
			}
			row := stmtOut.QueryRow(m.KeyMsg.Keyword, m.KeyMsg.Destlink)
			index := 200
			if err := row.Scan(&index); err != nil {
				self.log.Printf("row scan fails or result is zero")
			}
			m.CIndex = int64(index)
			msg.Content = m
			return
		}
		if ir.State == 1 {
			break
		}
		time.Sleep(3000 * time.Millisecond)
		//如果连续5次没有从api取到结果
		//那么从ranking_detail表取结果
		//如果ranking_detail表仍然没有取到结果
		//使用一个缺省的排名和指数order = 50 index = 200
		//等待次日任务运行时从api获取最新数据
		if count += 1; count >= 5 {
			self.log.Printf("use ranking_detail of result")
			stmtOut, err := self.db.Prepare(`SELECT current_index FROM ranking_detail WHERE keyword = ? AND destlink = ?`)
			defer stmtOut.Close()
			if err != nil {
				self.log.Printf("db prepare fails %s", err)
				msg.Err = errors.New("db prepare fails")
				return
			}
			row := stmtOut.QueryRow(m.KeyMsg.Keyword, m.KeyMsg.Destlink)
			index := 200
			if err := row.Scan(&index); err != nil {
				self.log.Printf("row scan fails or result is zero")
			}
			m.CIndex = int64(index)
			msg.Content = m
			return
		}
	}
	self.log.Printf("%+v", ir)
	if m.HIndex == 0 {
		m.HIndex = ir.Data[0].Allindex
	}
	m.CIndex = ir.Data[0].Allindex
	msg.Content = m
}

//首页按天计费
func (self *Payment) SProcess(msg *sexredis.Msg) {
	self.log.Printf("payment for keyword")
	m := msg.Content.(NormMsg)
	if m.KeyMsg.Status == RANKING_STATUS_CANCEL {
		return
	}
	//欠费超过设定容忍度,取消优化
	stmtOut, err := self.db.Prepare("SELECT balance FROM ranking_pay WHERE uid IN(?)")
	defer stmtOut.Close()
	if err != nil {
		self.log.Printf("db.Prepare fails %s", err)
		msg.Err = errors.New("db.Prepare fails")
		return
	}
	row := stmtOut.QueryRow(m.KeyMsg.Uid)
	var balance int64
	if err := row.Scan(&balance); err != nil {
		self.log.Printf("row.Scan fails %s", err)
		msg.Err = errors.New("row.Scan fails")
		return
	}
	if balance < self.c.Owed {
		m.KeyMsg.Status = RANKING_STATUS_CANCEL
		return
	}

	//根据指数和单价计算费用
	if m.CIndex < self.c.NIBase {
		m.CIndex = self.c.NIBase
	}
	m.Cost = m.CIndex * self.c.Price
	msg.Content = m
	if m.COrder > 10 {
		return
	}

	tx, err := self.db.Begin()
	if err != nil {
		self.log.Printf("tx begin fails %s", err)
		msg.Err = errors.New("tx begin fails")
		return
	}

	stmtIn, err := tx.Prepare(`UPDATE ranking_pay SET balance = balance - ? WHERE uid IN(?)`)
	defer stmtIn.Close()
	if err != nil {
		self.log.Printf("tx.Prepare fails %s", err)
		msg.Err = errors.New("tx.Prepare fails")
		return
	}

	_, err = stmtIn.Exec(m.Cost, m.KeyMsg.Uid)
	if err != nil {
		tx.Rollback()
		self.log.Printf("exec sql fails %s", err)
		msg.Err = errors.New("exec sql fails")
		return
	}
	//记录消费记录
	stmtInPayLog, err := tx.Prepare(`INSERT INTO ranking_consume_log(uid, kid, balance) VALUES(?, ?, ?)`)
	defer stmtInPayLog.Close()
	if err != nil {
		self.log.Printf("tx.Prepare fails %s", err)
		msg.Err = errors.New("tx.Prepare fails")
		return
	}
	_, err = stmtInPayLog.Exec(m.KeyMsg.Uid, m.KeyMsg.Id, m.Cost)
	if err != nil {
		tx.Rollback()
		self.log.Printf("exec sql fails %s", err)
		msg.Err = errors.New("exec sql fails")
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		self.log.Printf("tx.Commit fails %s", err)
		msg.Err = errors.New("tx.Commit fails %s")
		return
	}

}

//按时间段生成正态分布数据
func (self *NormCreate) SProcess(msg *sexredis.Msg) {
	//msg type ok?
	m := msg.Content.(NormMsg)
	m.Hour = make(map[string]int64)
	//生成24小时的正态分布数据 样本数量1000
	nt := normByTime(-1.2, 1.2, 0.1, 1000)
	//随机种子
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var ratio int64
	if m.COrder > 10 {
		ratio = r.Int63n(self.c.NIRatio[1]-self.c.NIRatio[0]) + self.c.NIRatio[0]
	} else {
		ratio = r.Int63n(self.c.OneIRatio[1]-self.c.OneIRatio[0]) + self.c.OneIRatio[0]
	}
	for k, v := range nt {
		m.Hour[fmt.Sprint(k)] = int64(math.Floor(float64(m.CIndex) * (float64(ratio) / float64(100)) * v))
	}

	if m.CIndex <= 500 {
		ratio = r.Int63n(10-4) + 4
		for i := 0; i < int(ratio); i++ {
			ratioH := r.Int63n(22-8) + 8
			m.Hour[fmt.Sprint(ratioH)] += 1
		}
	} else if m.CIndex <= 1000 {
		ratio = r.Int63n(30-20) + 20
		for i := 0; i < int(ratio); i++ {
			ratioH := r.Int63n(22-8) + 8
			m.Hour[fmt.Sprint(ratioH)] += 1
		}
	}
	msg.Content = m
}

//正态分布数据置入队列
//没有地区信息的正态分布数据放入 norm_common_queue
//带有地区信息的正态分布数据放入 norm_area_queue
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
	if m.KeyMsg.KeyCity != "" || m.KeyMsg.KeyProvince != "" {
		if _, err := rc.RPush(RANKING_AREA_NORM_QUEUE, js); err != nil {
			self.log.Printf("put msg in queue fails %s", err)
			msg.Err = errors.New("put msg in queue fails")
			return
		}
		self.log.Printf("put msg in %s", RANKING_AREA_NORM_QUEUE)
	} else {
		if _, err := rc.RPush(RANKING_COMMON_NORM_QUEUE, js); err != nil {
			self.log.Printf("put msg in queue fails %s", err)
			msg.Err = errors.New("put msg in queue fails")
			return
		}
		self.log.Printf("put msg in %s", RANKING_COMMON_NORM_QUEUE)
	}

}
func (self *PutInTask) SProcess(msg *sexredis.Msg) {
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

	if m.KeyMsg.Status == RANKING_STATUS_CANCEL {
		self.log.Printf("the keyword is cancel and not put in")
		return
	}
	taskMsg := TaskMsg{NormMsg: m}
	taskMsg.InitTime = time.Now().UnixNano() / (1000 * 1000)

	if js, err = json.Marshal(taskMsg); err != nil {
		self.log.Printf("Marshal json fails %s", err)
		msg.Err = errors.New("Marshal json fails")
		return
	}
	if _, err := rc.RPush(RANKING_TASK_QUEUE, js); err != nil {
		self.log.Printf("put msg in queue fails %s", err)
		msg.Err = errors.New("put msg in queue fails")
		return
	}
	self.log.Printf("put msg in %s", RANKING_TASK_QUEUE)
}
func (self *Recoder) SProcess(msg *sexredis.Msg) {
	//msg type ok?
	m := msg.Content.(NormMsg)
	stmtIn, err := self.db.Prepare(`INSERT INTO ranking_detail(id, uid, owner, keyword, destlink, history_order, 
	current_order, history_index, current_index, city_key, province_key, cost, status, logtime) 
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
	current_order = VALUES(current_order), current_index = VALUES(current_index), cost = VALUES(cost)`)

	defer stmtIn.Close()
	if err != nil {
		self.log.Printf("db.Prepare fails %s", err)
		msg.Err = errors.New("db.Prepare fails")
		return
	}
	if _, err := stmtIn.Exec(m.KeyMsg.Id, m.KeyMsg.Uid, m.KeyMsg.Owner, m.KeyMsg.Keyword, m.KeyMsg.Destlink, m.HOrder, m.COrder,
		m.HIndex, m.CIndex, m.KeyMsg.KeyCity, m.KeyMsg.KeyProvince, m.Cost, m.KeyMsg.Status, time.Now().Format("2006-01-02 15:04:05")); err != nil {
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
	if _, err := stmtIn.Exec(m.KeyMsg.Id, m.KeyMsg.Uid, m.KeyMsg.Owner, m.KeyMsg.Keyword, m.KeyMsg.Destlink, m.HOrder, m.COrder,
		m.HIndex, m.CIndex, m.KeyMsg.KeyCity, m.KeyMsg.KeyProvince, m.Cost, m.KeyMsg.Status, m.KeyMsg.Logtime); err != nil {
		self.log.Printf("stmtIn.Exec fails %s", err)
		msg.Err = errors.New("stmtIn.Exec fails")
		return
	}
	self.log.Printf("db order log in %s", "ranking_log")
}
