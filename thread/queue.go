package thread

import (
	"errors"
	"io/ioutil"
	"net/http"
)

type Queue struct {
	uri     string
	msgchan chan Msg
}

type Msg struct {
	Content interface{}
	Err     error
}

type Processor interface {
	SProcess(msg *Msg)
}

func New() *Queue {
	q := new(Queue)
	return q
}

func (self *Queue) SetRequestUri(uri string) {
	self.uri = uri
}

func (self *Queue) get() Msg {
	var (
		msg Msg
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
			msg.Content = string(body)
		}
	}
	return msg
}

/*
use channel implement like python yield
*/
func (self *Queue) consume() {
	self.msgchan = make(chan Msg)

	for {
		msg := self.get()
		if n, ok := msg.Content.(int64); ok && n > 0 {
			self.msgchan <- msg
		}
	}
}

func (self *Queue) yield() (msg Msg) {
	return <-self.msgchan
}

func (self *Queue) Worker(pnum uint, serial bool, ps ...Processor) {
	control := make(chan Msg, pnum)
	go func() {
		self.consume()
	}()

	go func() {
		for {
			msg := self.yield()
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
