package main

import (
	"github.com/huzorro/spfactor/sexredis"
	"log"
)

type AreaProxy struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
}

func (self *AreaProxy) SProcess(msg *sexredis.Msg) {
	//
	self.log.Printf("get proxy of the area")

}
