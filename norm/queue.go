package norm

import (
	"github.com/huzorro/spfactor/sexredis"
)

//处理普通数据队列
//需要提交数据到任务队列的处理器
type Queue struct {
	Norm  *sexredis.Queue
	Proxy *sexredis.Queue
}

func New() *Queue {
	q := new(Queue)
	return q
}

func (self *Queue) Worker(pnum uint, serial bool, ps ...sexredis.Processor) {
	control := make(chan sexredis.Msg, pnum)
	go func() {
		self.Norm.Consume()
	}()
	go func() {
		self.Proxy.Consume()
	}()

	go func() {
		for {
			proxy := self.Proxy.Yield()
			norm := self.Norm.Yield()
			msgs := []sexredis.Msg{norm, proxy}
			msg := sexredis.Msg{msgs, nil}
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
