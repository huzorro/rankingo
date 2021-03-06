package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/huzorro/spfactor/sexredis"
	"log"
)

type RankResult struct {
	c   *Cfg
	log *log.Logger
	db  *sql.DB
}

type UpdateOrder struct {
	c   *Cfg
	log *log.Logger
	db  *sql.DB
}

//根据刷新成功的结果 更新排名
func (self *UpdateOrder) SProcess(msg *sexredis.Msg) {
	self.log.Printf("update order of rank result")
	//msg type ok?
	if _, ok := msg.Content.(TaskMsg); !ok {
		msg.Err = errors.New("msg type errors")
		return
	}
	m := msg.Content.(TaskMsg)
	if m.NormMsg.KeyMsg.Status == 100 {
		self.log.Printf("not update so mobile rank")
		return
	}
	stmtIn, err := self.db.Prepare(`UPDATE ranking_detail SET current_order = ? WHERE keyword = ? AND destlink = ?`)
	if err != nil {
		self.log.Printf("db.Prepare exec fails %s", err)
		msg.Err = errors.New("db.Prepare exec fails")
		return
	}
	defer stmtIn.Close()
	if rs, err := stmtIn.Exec(m.NormMsg.COrder, m.NormMsg.KeyMsg.Keyword, m.NormMsg.KeyMsg.Destlink); err != nil {
		self.log.Printf("update order fails %s", err)
		msg.Err = errors.New("update order fails")
		return
	} else {
		id, _ := rs.RowsAffected()
		self.log.Printf("update order success of RowsAffected=%d", id)
	}

}

//客户端提交的结果数据入库
func (self *RankResult) SProcess(msg *sexredis.Msg) {
	self.log.Printf("rank result put in db")
	var (
		taskMsg TaskMsg
		hours   []byte
	)
	//msg type ok?
	m := msg.Content.(string)
	if err := json.Unmarshal([]byte(m), &taskMsg); err != nil {
		self.log.Printf("Unmarshal json fails %s", err)
		msg.Err = errors.New("Unmarshal json fails")
		return
	}
	stmtIn, err := self.db.Prepare(`INSERT INTO ranking_result_log 
	(itime, corder, horder, cindex, hindex, cost, hours, cancel, 	
	vpsip, adsltext, keyid, uid, owner, keyword, destlink, city, province, status, ilogtime) 
	VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	defer stmtIn.Close()

	if err != nil {
		self.log.Printf("db.Prepare exec fails %s", err)
		msg.Err = errors.New("db.Prepare exec fails")
		return
	}
	//map to json
	if hours, err = json.Marshal(taskMsg.NormMsg.Hour); err != nil {
		self.log.Printf("marshal json fails %s", err)
		msg.Err = errors.New("marshal json fails")
		return
	}
	if rs, err := stmtIn.Exec(taskMsg.InitTime, taskMsg.NormMsg.COrder, taskMsg.NormMsg.HOrder,
		taskMsg.NormMsg.CIndex, taskMsg.NormMsg.HIndex,
		taskMsg.NormMsg.Cost, string(hours), taskMsg.NormMsg.Cancel, taskMsg.ProxyMsg.Ip,
		taskMsg.ProxyMsg.Port, taskMsg.NormMsg.KeyMsg.Id, taskMsg.NormMsg.KeyMsg.Uid,
		taskMsg.NormMsg.KeyMsg.Owner, taskMsg.NormMsg.KeyMsg.Keyword,
		taskMsg.NormMsg.KeyMsg.Destlink, taskMsg.NormMsg.KeyMsg.KeyCity, taskMsg.NormMsg.KeyMsg.KeyProvince,
		taskMsg.NormMsg.KeyMsg.Status, taskMsg.NormMsg.KeyMsg.Logtime); err != nil {
		self.log.Printf("rank result put in db fails %s", err)
		msg.Err = errors.New("rank result put in db fails")
		return
	} else {
		msg.Content = taskMsg
		lastId, _ := rs.LastInsertId()
		self.log.Printf("rank result put in db success of LastInsertId=%d", lastId)
	}
}
