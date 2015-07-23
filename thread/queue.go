package thread

import (
	"errors"
	"github.com/huzorro/spfactor/sexredis"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//通过任务队列控制子进程的数量
//需要启动子进程、等待子进程执行完毕、获取执行结果回写到任务执行结果队列的处理器
type Queue struct {
	log        *log.Logger
	uri        string
	adslCName  string
	adslUser   string
	adslPasswd string
	sexredis.Queue
	Msgchan chan sexredis.Msg
}

func New() *Queue {
	q := new(Queue)
	q.Msgchan = make(chan sexredis.Msg)
	return q
}
func (self *Queue) SetRequestUri(uri string) {
	self.uri = uri
}
func (self *Queue) SetAdsl(cname, user, passwd string, log *log.Logger) {
	self.adslCName = cname
	self.adslUser = user
	self.adslPasswd = passwd
	self.log = log
}
func (self *Queue) Get() sexredis.Msg {
	var (
		msg sexredis.Msg
	)
	resp, err := http.Get(self.uri)
	defer resp.Body.Close()
	if err != nil {
		// handle error
		msg.Err = errors.New("thread queue request data fails")
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			msg.Err = errors.New("thread queue response fails")
		} else {
			msg.Content, err = strconv.ParseInt(string(body), 10, 64)
		}
	}
	return msg
}

/*
use channel implement like python yield
*/
func (self *Queue) Consume() {
	//	self.Msgchan = make(chan sexredis.Msg)
	for {
		msg := self.Get()
		if n, ok := msg.Content.(int64); ok && n > 0 {
			self.Msgchan <- msg
		}
		time.Sleep(3000 * time.Millisecond)
	}
}

func (self *Queue) Worker(pnum uint, serial bool, ps ...sexredis.Processor) {
	control := make(chan sexredis.Msg, pnum)
	adsl := make(chan sexredis.Msg, pnum)
	go func() {
		self.Consume()
	}()

	go func() {
		for {
			msg := self.Yield()
			control <- msg
			go func() {
				if serial {
					for _, sp := range ps {
						sp.SProcess(&msg)
						if msg.Err != nil {
							break
						}
					}
				} else {
					//并行处理
				}
				//				<-control
				//写入adsl拨号控制通道
				adsl <- msg
			}()
		}
	}()

	//处理adsl挂断和拨号
	go func() {
		for {
			//等待所有任务结束
			for i := 0; uint(i) < pnum; i++ {
				<-adsl
			}
			//挂断adsl
			for {
				if rs, err := self.AdslDisconnect(); err == nil {
					log.Printf("%s %s", self.adslCName, rs)
					break

				} else {
					log.Printf("%s %s %s", self.adslCName, rs, err)
					continue
				}
			}
			time.Sleep(5000 * time.Millisecond)
			//adsl拨号
			for {
				if rs, err := self.AdslConnect(); err == nil && strings.Contains(rs, "已连接") {
					log.Printf("%s %s", self.adslCName, rs)
				} else {
					log.Printf("%s %s %s", self.adslCName, rs, err)
					continue
				}
			}
			//拨号成功清空任务控制通道, 继续下次任务
			for i := 0; uint(i) < pnum; i++ {
				<-control
			}
		}
	}()
}

func (self *Queue) AdslConnect() (string, error) {
	cmd := exec.Command("rasdial", self.adslCName, self.adslUser, self.adslPasswd)
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}
	bytesErr, err := ioutil.ReadAll(errPipe)
	if err != nil {
		return "", err
	}

	if len(bytesErr) != 0 {
		return "", errors.New(string(bytesErr))
	}
	bytesResult, err := ioutil.ReadAll(outPipe)

	if err != nil {
		return "", err
	}
	return string(bytesResult), nil
}

func (self *Queue) AdslDisconnect() (string, error) {
	cmd := exec.Command("rasdial", self.adslCName, "/disconnect")
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}
	bytesErr, err := ioutil.ReadAll(errPipe)
	if err != nil {
		return "", err
	}

	if len(bytesErr) != 0 {
		return "", errors.New(string(bytesErr))
	}
	bytesResult, err := ioutil.ReadAll(outPipe)

	if err != nil {
		return "", err
	}
	return string(bytesResult), nil
}
