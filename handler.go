package main

import (
	"errors"
	"github.com/huzorro/spfactor/sexredis"
	"io/ioutil"
	"log"
	"os/exec"
)

type Control struct {
	c   *Cfg
	log *log.Logger
}

//berserkJS --command --script=/home/huzorro/vagrant/jsprojects/ranking/baidu.js

func (self *Control) SProcess(msg *sexredis.Msg) {
	cmd := exec.Command(self.c.RankPath, self.c.RankParam)
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		self.log.Printf("get std out pipe fails %s", err)
		msg.Err = errors.New("get std out pipe fails")
		return
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		self.log.Printf("get std err pipe fails %s", err)
		msg.Err = errors.New("get std err pipe fails")
		return
	}

	if err := cmd.Start(); err != nil {
		self.log.Printf("cmd exec fails %s", err)
		msg.Err = errors.New("cmd exec fials")
		return
	}
	bytesErr, err := ioutil.ReadAll(errPipe)
	if err != nil {
		self.log.Printf("get bytes error from err pipe fails %s", err)
		msg.Err = errors.New("get bytes error from err pipe fails")
		return
	}

	if len(bytesErr) != 0 {
		self.log.Printf("cmd exec fails %s", string(bytesErr))
		msg.Err = errors.New("cmd exec fails")
		return
	}
	bytesResult, err := ioutil.ReadAll(outPipe)

	if err != nil {
		self.log.Printf("get cmd exec result fails %s", err)
		msg.Err = errors.New("get cmd exec result fails")
		return
	}
	log.Printf(string(bytesResult))
}
