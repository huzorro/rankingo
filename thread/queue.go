package thread

import (
	"errors"
	"github.com/huzorro/spfactor/sexredis"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

//通过任务队列控制子进程的数量
//需要启动子进程、等待子进程执行完毕、获取执行结果回写到任务执行结果队列的处理器
type Queue struct {
	uri string
	sexredis.Queue
}

func New() *Queue {
	return new(Queue)
}
func (self *Queue) SetRequestUri(uri string) {
	self.uri = uri
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
	self.Msgchan = make(chan sexredis.Msg)
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
				<-control
			}()
		}
	}()
}
