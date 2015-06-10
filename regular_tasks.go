package main

import (
	"database/sql"
	"github.com/huzorro/spfactor/sexredis"
	"github.com/xiam/resp"
	"log"
	"reflect"
	"strconv"
	"time"
)

type RegularTasks struct {
	c   *Cfg
	log *log.Logger
	p   *sexredis.RedisPool
	db  *sql.DB
}

func (self *RegularTasks) Handler(f func(r *RegularTasks)) {
	timer := time.NewTicker(10 * time.Minute)
	for {
		select {
		case <-timer.C:
			go func() {
				f(self)
				if int64(time.Now().Hour()) == self.c.Timing {
					f(self)
				}
			}()
		}
	}
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
		for {
			res, _ := redisClient.HScan(RANKING_KEYWORD_HASH, int64(cur))
			for _, v := range res {
				switch m := v.(type) {
				case []interface{}:
					r.log.Printf("Got an array of type %s with %d elements (%v).\n",
						reflect.TypeOf(m).Kind(), len(m), string(m[0].(*resp.Message).Interface().([]byte)))
					for i := 0; i < len(m); i += 2 {
						mk := string(m[i].(*resp.Message).Interface().([]byte))
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
