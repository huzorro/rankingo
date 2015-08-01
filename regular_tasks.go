package main

import (
	"database/sql"
	"github.com/huzorro/spfactor/sexredis"
	"github.com/robfig/cron"
	"github.com/xiam/resp"
	"log"
	"reflect"
	"strconv"
)

type RegularTasks struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
	db  *sql.DB
}

//func (self *RegularTasks) Handler(f func(r *RegularTasks)) {
//	timer := time.NewTicker(10 * time.Minute)
//	go func() {
//		for {
//			select {
//			case <-timer.C:
//				go func() {
//					if int64(time.Now().Hour()) == self.c.Timing {
//						f(self)

//					}
//				}()
//			}
//		}
//	}()
//}
func (self *RegularTasks) Handler() {
	c := cron.New()
	c.AddFunc(self.c.Crontab, func() {
		redisClient, err := self.p.Get()
		defer self.p.Close(redisClient)
		if err != nil {
			self.log.Printf("get redis connection fails %s", err)
		}
		var cur int64 = 0
		//		mp := make(map[string]string)
		//清空task queue
		if n, err := redisClient.Del(RANKING_TASK_QUEUE); err != nil {
			self.log.Printf("%s del fails %s", RANKING_TASK_QUEUE, err)
			return
		} else {
			self.log.Printf("%s del successed N:%d", RANKING_TASK_QUEUE, n)
		}
		for {
			res, _ := redisClient.HScan(RANKING_KEYWORD_HASH, int64(cur))
			for _, v := range res {
				switch m := v.(type) {
				case []interface{}:
					self.log.Printf("Got an array of type %s with %d elements (%v).\n",
						reflect.TypeOf(m).Kind(), len(m), string(m[0].(*resp.Message).Interface().([]byte)))
					for i := 0; i < len(m); i += 2 {
						//						mk := string(m[i].(*resp.Message).Interface().([]byte))
						mv := string(m[i+1].(*resp.Message).Interface().([]byte))
						//mp[mk] = mv
						redisClient.RPush(RANKING_KEYWORD_QUEUE, mv)
						//log.Printf("%s:%s", mk, mv)
					}
				case *resp.Message:
					self.log.Printf("Got value of kind %s (%v), we use the integer part: %d   []bytes : %s\n", reflect.TypeOf(m).Kind(), m, m.Integer, string(m.Bytes))
					cur, _ = strconv.ParseInt(string(m.Bytes), 10, 64)
				}
			}
			if cur == 0 {
				break
			}
		}

	})
	c.Start()
}
func rebuild() func(r *RegularTasks) {
	return func(r *RegularTasks) {
		redisClient, err := r.p.Get()
		defer r.p.Close(redisClient)
		if err != nil {
			r.log.Printf("get redis connection fails %s", err)
		}
		var cur int64 = 0
		//		mp := make(map[string]string)
		//清空task queue
		if n, err := redisClient.Del(RANKING_TASK_QUEUE); err != nil {
			r.log.Printf("%s del fails %s", RANKING_TASK_QUEUE, err)
			return
		} else {
			r.log.Printf("%s del successed N:%d", RANKING_TASK_QUEUE, n)
		}
		for {
			res, _ := redisClient.HScan(RANKING_KEYWORD_HASH, int64(cur))
			for _, v := range res {
				switch m := v.(type) {
				case []interface{}:
					r.log.Printf("Got an array of type %s with %d elements (%v).\n",
						reflect.TypeOf(m).Kind(), len(m), string(m[0].(*resp.Message).Interface().([]byte)))
					for i := 0; i < len(m); i += 2 {
						//						mk := string(m[i].(*resp.Message).Interface().([]byte))
						mv := string(m[i+1].(*resp.Message).Interface().([]byte))
						//mp[mk] = mv
						redisClient.RPush(RANKING_KEYWORD_QUEUE, mv)
						//log.Printf("%s:%s", mk, mv)
					}
				case *resp.Message:
					r.log.Printf("Got value of kind %s (%v), we use the integer part: %d   []bytes : %s\n", reflect.TypeOf(m).Kind(), m, m.Integer, string(m.Bytes))
					cur, _ = strconv.ParseInt(string(m.Bytes), 10, 64)
				}
			}
			if cur == 0 {
				break
			}
		}
	}
}
