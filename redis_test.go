package main

import (
	"github.com/gosexy/redis"
	"github.com/huzorro/spfactor/sexredis"
	"github.com/xiam/resp"
	"log"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func TestHScan(t *testing.T) {
	redisPool := &sexredis.RedisPool{make(chan *redis.Client, 10), func() (*redis.Client, error) {
		client := redis.New()
		err := client.Connect("localhost", uint(6379))
		return client, err
	}}
	redisClient, _ := redisPool.Get()
	defer redisPool.Close(redisClient)
	var cur int64 = 0
	mp := make(map[string]string)
	for {
		res, _ := redisClient.HScan(RANKING_KEYWORD_HASH, int64(cur))
		for _, v := range res {
			switch m := v.(type) {
			case []interface{}:
				log.Printf("Got an array of type %s with %d elements (%v).\n",
					reflect.TypeOf(m).Kind(), len(m), string(m[0].(*resp.Message).Interface().([]byte)))
				for i := 0; i < len(m); i += 2 {
					mk := string(m[i].(*resp.Message).Interface().([]byte))
					mv := string(m[i+1].(*resp.Message).Interface().([]byte))
					mp[mk] = mv
					log.Printf("%s:%s", mk, mv)

				}
			case *resp.Message:
				log.Printf("Got value of kind %s (%v), we use the integer part: %d   []bytes : %s\n", reflect.TypeOf(m).Kind(), m, m.Integer, string(m.Bytes))
				cur, _ = strconv.ParseInt(string(m.Bytes), 10, 64)
			}
		}
		if cur == 0 {
			break
		}
	}
}

func TestRegularTasks(t *testing.T) {
	redisPool := &sexredis.RedisPool{make(chan *redis.Client, 10), func() (*redis.Client, error) {
		client := redis.New()
		err := client.Connect("localhost", uint(6379))
		return client, err
	}}

	logger := log.New(os.Stdout, "\r\n", log.Ldate|log.Ltime|log.Lshortfile)
	rt := &RegularTasks{nil, logger, redisPool, nil}
	rt.Handler(rebuild())
}
