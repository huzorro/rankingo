package task

import (
	"errors"
	"github.com/huzorro/spfactor/sexredis"
	"time"
)

//通过任务队列控制代理数据的获取时机
type Queue struct {
	sexredis.Queue
}

func New() *Queue {
	return new(Queue)
}
func (self *Queue) Get() sexredis.Msg {
	if n, err := self.LLen(); err == nil {
		return sexredis.Msg{n, nil}
	} else {
		return sexredis.Msg{nil, errors.New("get size of queue fails")}
	}
}

func (self *Queue) Consume() {
	self.Msgchan = make(chan sexredis.Msg)
	for {
		msg := self.Get()
		if n, ok := msg.Content.(int64); ok && n == 0 {
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
