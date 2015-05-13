package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/huzorro/rankingo/thread"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
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

func (self *Control) SProcess(msg *thread.Msg) {
	//msg type ok?

	if _, ok := msg.Content.(int64); !ok {
		self.log.Printf("Msg type error %+", msg)
		msg.Err = errors.New("Msg type error")
		return
	}

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
	msg.Content = string(bytesResult)
}

func (self *Submit) SProcess(msg *thread.Msg) {
	//msg type ok?
	var (
		m     string
		ok    bool
		v     TaskMsg
		value string
	)
	if m, ok = msg.Content.(string); !ok {
		self.log.Printf("Msg type error %+", msg)
		msg.Err = errors.New("Msg type error")
		return
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", self.c.SubmitUri, strings.NewReader(m))
	if err != nil {
		self.log.Printf("submit request fails %s", err)
		msg.Err = errors.New("submit request fails")
		return
	}
	if err := json.Unmarshal([]byte(m), &v); err != nil {
		self.log.Printf("json decode fails %s", err)
		msg.Err = errors.New("json decode fails")
		return
	}
	sa := url.Values{}
	vType := reflect.TypeOf(&v).Elem()
	vValue := reflect.ValueOf(&v).Elem()

	for i := 0; i < vType.NumField(); i++ {
		tag := vType.Field(i).Tag.Get("json")
		fN := vType.Field(i).Name
		switch vType.Field(i).Type.Kind() {
		case reflect.String:
			value = vValue.FieldByName(fN).String()
		case reflect.Int:
			value = strconv.FormatInt(vValue.FieldByName(fN).Int(), 10)
		case reflect.Float64:
			value = strconv.FormatFloat(vValue.FieldByName(fN).Float(), 'f', -1, 64)
		default:
			//
		}
		sa.Add(tag, value)
	}
	unescape, _ := url.QueryUnescape(sa.Encode())
	self.c.Authorization = fmt.Sprintf("%x", md5.Sum([]byte(unescape)))

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
	self.log.Printf(string(body))
}
