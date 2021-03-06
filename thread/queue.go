package thread

import (
	"errors"
	"github.com/huzorro/spfactor/sexredis"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
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
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		// handle error
		msg.Err = errors.New("thread queue request data fails")
		//挂断 重拨
		self.log.Printf("access net fails to disadsl then reconnect %s", msg.Err)
		//挂断adsl
		for {
			self.log.Printf("adsl disconnecting cname:%s", self.adslCName)

			if rs, err := self.AdslDisconnect(); err == nil {
				self.log.Printf("adsl disconnected cname:%s, result:%s", self.adslCName, rs)
				break

			} else {
				self.log.Printf("adsl disconnected fails cname:%s, result:%s, err:%s", self.adslCName, rs, err)
				continue
			}
		}
		time.Sleep(5000 * time.Millisecond * 2)
		//adsl拨号
		for {
			self.log.Printf("adsl connecting cname:%s, user:%s, passwd:%s", self.adslCName, self.adslUser, self.adslPasswd)

			if rs, err := self.AdslConnect(); err == nil {
				self.log.Printf("adsl connected  cname:%s, result:%s", self.adslCName, rs)
				break
			} else {
				self.log.Printf("adsl connect fails cname:%s, result:%s, err:%s", self.adslCName, rs, err)
				continue
			}
		}
		time.Sleep(5000 * time.Millisecond)

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
		self.log.Println("go consume...")
		msg := self.Get()
		self.log.Printf("1%+v", msg)
		if n, ok := msg.Content.(int64); ok && n > 0 {
			self.log.Println(n, ok)
			self.Msgchan <- msg
		}
		self.log.Printf("2%+v", msg)
		time.Sleep(3000 * time.Millisecond)
	}
}

func (self *Queue) Worker(pnum uint, serial bool, ps ...sexredis.Processor) {
	self.log.Printf("%d %s", pnum, self.uri)
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
				self.log.Printf("adsl disconnecting cname:%s", self.adslCName)

				if rs, err := self.AdslDisconnect(); err == nil {
					self.log.Printf("adsl disconnected cname:%s, result:%s", self.adslCName, rs)
					break

				} else {
					self.log.Printf("adsl disconnected fails cname:%s, result:%s, err:%s", self.adslCName, rs, err)
					continue
				}
			}
			time.Sleep(5000 * time.Millisecond * 2)
			//adsl拨号
			for {
				self.log.Printf("adsl connecting cname:%s, user:%s, passwd:%s", self.adslCName, self.adslUser, self.adslPasswd)

				if rs, err := self.AdslConnect(); err == nil {
					self.log.Printf("adsl connected  cname:%s, result:%s", self.adslCName, rs)
					break
				} else {
					self.log.Printf("adsl connect fails cname:%s, result:%s, err:%s", self.adslCName, rs, err)
					continue
				}
			}
			time.Sleep(5000 * time.Millisecond)
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
