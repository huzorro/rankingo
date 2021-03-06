package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/huzorro/spfactor/sexredis"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	RANKING_TASK_QUEUE = "ranking:task:queue"
)

type Control struct {
	c   *Cfg
	log *log.Logger
}

type Submit struct {
	c   *Cfg
	log *log.Logger
}

//berserkJS --command --script=/home/huzorro/vagrant/jsprojects/ranking/baidu.js

func (self *Control) SProcess(msg *sexredis.Msg) {
	//msg type ok?

	if _, ok := msg.Content.(int64); !ok {
		self.log.Printf("Msg type error %+", msg)
		msg.Err = errors.New("Msg type error")
		return
	}
	//随机种子
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	radio := r.Int63n(7)
	rankParams := [7][]string{self.c.RankParam, self.c.RankMobileParam, self.c.RankParam, self.c.RankParam, self.c.RankParam, self.c.RankParam, self.c.RankParam}
	var cmd *exec.Cmd
	switch true {
	case self.c.RankMix:
		cmd = exec.Command(self.c.RankPath, rankParams[radio]...)
	case self.c.RankPc:
		cmd = exec.Command(self.c.RankPath, rankParams[0]...)
	case self.c.RankMobile:
		cmd = exec.Command(self.c.RankPath, rankParams[1]...)
	default:
		cmd = exec.Command(self.c.RankPath, self.c.RankParam...)
	}
	//	outPipe, err := cmd.StdoutPipe()
	//	if err != nil {
	//		self.log.Printf("get std out pipe fails %s", err)
	//		msg.Err = errors.New("get std out pipe fails")
	//		return
	//	}

	//	errPipe, err := cmd.StderrPipe()
	//	if err != nil {
	//		self.log.Printf("get std err pipe fails %s", err)
	//		msg.Err = errors.New("get std err pipe fails")
	//		return
	//	}

	if err := cmd.Start(); err != nil {
		self.log.Printf("cmd exec fails %s", err)
		msg.Err = errors.New("cmd exec fials")
		return
	}
	var timer *time.Timer
	timer = time.AfterFunc(6*time.Minute, func() {
		timer.Stop()
		if err := cmd.Process.Kill(); err != nil {
			self.log.Printf("exec process timeout to kill fails %s", err)
			msg.Err = errors.New("exec process timeout to kill fails")
			return
		} else {
			self.log.Printf("exec process timeout to kill successed")
			msg.Err = errors.New("exec process timeout to kill successed")
			return
		}
	})

	//	bytesErr, err := ioutil.ReadAll(errPipe)
	//	if err != nil {
	//		self.log.Printf("get bytes error from err pipe fails %s", err)
	//		msg.Err = errors.New("get bytes error from err pipe fails")
	//		return
	//	}

	//	if len(bytesErr) != 0 {
	//		self.log.Printf("cmd exec fails %s", string(bytesErr))
	//		msg.Err = errors.New("cmd exec fails")
	//		return
	//	}
	//	bytesResult, err := ioutil.ReadAll(outPipe)

	//	if err != nil {
	//		self.log.Printf("get cmd exec result fails %s", err)
	//		msg.Err = errors.New("get cmd exec result fails")
	//		return
	//	}
	//	self.log.Printf("read exec result:%s", string(bytesResult))
	//	msg.Content = string(bytesResult)
	if err := cmd.Wait(); err != nil {
		self.log.Printf("cmd wait fails %s", err)
		return
	}
	timer.Stop()
}

func (self *Submit) SProcess(msg *sexredis.Msg) {
	//msg type ok?
	var (
		m  string
		ok bool
	)
	if m, ok = msg.Content.(string); !ok {
		self.log.Printf("Msg type error %+", msg)
		msg.Err = errors.New("Msg type error")
		return
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", self.c.LogUri, strings.NewReader(m))
	if err != nil {
		self.log.Printf("submit request fails %s", err)
		msg.Err = errors.New("submit request fails")
		return
	}
	self.c.Authorization = fmt.Sprintf("%x", md5.Sum([]byte(m)))

	req.Header.Set("Authorization", self.c.Authorization)
	req.Header.Set("Content-Type", self.c.ContentType)
	req.Header.Set("Accept", self.c.Accept)

	resp, err := client.Do(req)

	if err != nil {
		self.log.Printf("submit request fails %s", err)
		msg.Err = errors.New("submit request fails")
		return
	}
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		self.log.Printf("response body read fails %s", err)
		msg.Err = errors.New("response body read fails")
		return
	}
	defer resp.Body.Close()
	self.log.Printf("submit response : %s", string(body))
}
