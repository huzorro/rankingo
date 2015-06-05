package area

import (
	"github.com/huzorro/rankingo/task"
	"github.com/huzorro/spfactor/sexredis"
)

//处理带有地域需求的队列
//需要获取代理、验证代理、提交数据到任务队列的处理器
type Queue struct {
	Task *task.Queue
	Area *sexredis.Queue
}

func New() *Queue {
	q := new(Queue)
	return q
}

func (self *Queue) Worker(pnum uint, serial bool, ps ...sexredis.Processor) {
	control := make(chan sexredis.Msg, pnum)
	go func() {
		self.Task.Consume()
	}()
	go func() {
		self.Area.Consume()
	}()

	go func() {
		for {
			self.Task.Yield()
			area := self.Area.Yield()
			msg := sexredis.Msg{area.Content, nil}
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
