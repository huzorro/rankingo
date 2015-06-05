package main

import (
	"database/sql"
	"flag"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gosexy/redis"
	"github.com/huzorro/rankingo/area"
	"github.com/huzorro/rankingo/norm"
	"github.com/huzorro/rankingo/task"
	"github.com/huzorro/rankingo/thread"
	"github.com/huzorro/spfactor/sexredis"
	"github.com/huzorro/woplus/tools"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"log"
	"net/http"
	"os"
)

type Cfg struct {
	//获取任务的api
	TaskUri string `json:"taskUri"`
	//获取任务数量的api
	TaskNUri string `json:"taskNUri"`
	//提交任务执行结果的api
	SubmitUri string `json:"submitUri"`
	//rank进程路径
	RankPath string `json:"rankPath"`
	//rank进程参数
	RankParam string `json:"rankParam"`
	//任务执行进程数量
	ThreadN int64 `json:"threadN"`
	//任务执行超时时间
	Timeout int64 `json:"timeout"`
	//提交任务执行结果的认证字符串
	Authorization string `json:"Authorization"`
	ContentType   string `json:"Content-Type"`
	Accept        string `json:"Accept"`
	//指数排名的百分比
	OIRatio float64 `json:"oiRatio"`
	//无指数关键字的基础百分比
	NIBase int64 `json:"nIBase"`
	//数据库类型
	Dbtype string `json:"dbtype"`
	//数据库连接uri
	Dburi string `json:"dburi"`
	//定时更新数据间隔
	Interval int64 `json:"interval"`
	//符合预期的时间点
	Timing int64 `json:"timing"`
	//页宽
	PageSize int64 `json:"pageSize"`
	//欠费容忍度 例如余额小于 < -1000
	//计费规则 当月在首页次数超过的天数 > 3 则扣除当月费用
	//价格计算规则 起步价开始的指数 例如200指数以内的设定一个起步价
	//价格计算规则 超过起步价指数的单价

	OrderApi    string `json:"orderApi"`
	OrderApiKey string `json:"orderApiKey"`
	IndexApi    string `json:"indexApi"`
	IndexApiKey string `json:"indexApiKey"`
	ProxyApi    string `json:"proxyApi"`
	CheckApi    string `json:"checkApi"`
}

type TaskResultMsg struct {
	TaskMsg   TaskMsg   `json:"taskMsg"`
	ResultMsg ResultMsg `json:"resultMsg"`
}
type ResultMsg struct {
	EndTime     int64  `json:"endTime`
	ExecTime    int64  `json:"execTime"`
	CostTime    int64  `json:"costTime"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func main() {

	portPtr := flag.String("port", ":10086", "service port")
	redisIdlePtr := flag.Int("redis", 20, "redis idle connections")
	dbMaxPtr := flag.Int("db", 10, "max db open connections")
	//config path
	cfgPathPtr := flag.String("config", "config.json", "config path name")
	//web
	webPtr := flag.Bool("web", false, "web sm start")
	apiPtr := flag.Bool("api", false, "rest api start")
	//key handler
	keyHandlerPtr := flag.Bool("key", false, "key handler start")
	normHandlerPtr := flag.Bool("norm", false, "common norm handler start")
	areaHandlerPtr := flag.Bool("area", false, "area norm handler start")
	proxyHandlerPtr := flag.Bool("proxy", false, "common proxy handler start")
	threadHandlerPtr := flag.Bool("thread", false, "thread control handler start")

	flag.Parse()

	//json config
	var cfg Cfg
	if err := tools.Json2Struct(*cfgPathPtr, &cfg); err != nil {
		log.Printf("load json config fails %s", err)
		panic(err.Error())
	}

	logger := log.New(os.Stdout, "\r\n", log.Ldate|log.Ltime|log.Lshortfile)

	redisPool := &sexredis.RedisPool{make(chan *redis.Client, *redisIdlePtr), func() (*redis.Client, error) {
		client := redis.New()
		err := client.Connect("localhost", uint(6379))
		return client, err
	}}
	db, err := sql.Open(cfg.Dbtype, cfg.Dburi)
	db.SetMaxOpenConns(*dbMaxPtr)

	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}

	mtn := martini.Classic()

	mtn.Map(logger)
	mtn.Map(redisPool)
	mtn.Map(db)

	cache := &Cache{db, redisPool}
	mtn.Map(cache)
	//	load rbac node
	if nMap, err := cache.RbacNodeToMap(); err != nil {
		logger.Printf("rbac node to map fails %s", err)
	} else {
		mtn.Map(nMap)
	}
	//load rbac menu
	if ms, err := cache.RbacMenuToSlice(); err != nil {
		logger.Printf("rbac menu to slice fails %s", err)
	} else {
		mtn.Map(ms)
	}
	//session
	store := sessions.NewCookieStore([]byte("secret123"))
	mtn.Use(sessions.Sessions("Qsession", store))
	//render
	rOptions := render.Options{}
	rOptions.Extensions = []string{".tmpl", ".html"}
	mtn.Use(render.Renderer(rOptions))

	mtn.Map(&cfg)

	if *webPtr {
		mtn.Get("/login", func(r render.Render) {
			r.HTML(200, "login", "")
		})
		mtn.Get("/logout", logout)
		mtn.Post("/login/check", loginCheck)
		mtn.Post("/key/add", keyAddAction)
		mtn.Post("/key/update", keyUpdateAction)
		mtn.Post("/key/one", keyOneAction)
		mtn.Get("/key/show", keyShowAction)
	}

	if *webPtr || *apiPtr {
		go http.ListenAndServe(*portPtr, mtn)
	}

	if *keyHandlerPtr {
		rc, err := redisPool.Get()
		defer redisPool.Close(rc)
		if err != nil {
			log.Printf("get redis connection fails %s", err)
			return
		}
		queue := sexredis.New()
		queue.SetRClient(RANKING_KEYWORD_QUEUE, rc)
		queue.Worker(2, true, &Order{&cfg, logger, redisPool}, &Index{&cfg, logger, redisPool},
			&PutIn{&cfg, logger, redisPool}, &Recoder{&cfg, logger, db})
	}

	if *normHandlerPtr {
		normRc, err := redisPool.Get()
		defer redisPool.Close(normRc)
		if err != nil {
			log.Printf("get redis connection fails %s", err)
			return
		}

		normQueue := sexredis.New()
		normQueue.SetRClient(RANKING_COMMON_NORM_QUEUE, normRc)

		proxyRc, err := redisPool.Get()
		defer redisPool.Close(proxyRc)
		if err != nil {
			log.Printf("get redis connection fails %s", err)
			return
		}
		proxyQueue := sexredis.New()
		proxyQueue.SetRClient(RANKING_PROXY_QUEUE, proxyRc)

		queue := norm.New()
		queue.Norm = normQueue
		queue.Proxy = proxyQueue
		queue.Worker(2, true, &Task{&cfg, logger, redisPool})
	}
	if *areaHandlerPtr {
		taskRc, err := redisPool.Get()
		defer redisPool.Close(taskRc)
		if err != nil {
			log.Printf("get redis connection fails %s", err)
			return
		}

		taskQueue := task.New()
		taskQueue.SetRClient(RANKING_TASK_QUEUE, taskRc)

		areaRc, err := redisPool.Get()
		defer redisPool.Close(areaRc)
		if err != nil {
			log.Printf("get redis connection fails %s", err)
			return
		}
		areaQueue := sexredis.New()
		areaQueue.SetRClient(RANKING_AREA_NORM_QUEUE, areaRc)

		queue := area.New()
		queue.Task = taskQueue
		queue.Area = areaQueue
		queue.Worker(2, true, &AreaProxy{&cfg, logger, redisPool})
	}

	if *proxyHandlerPtr {
		taskRc, err := redisPool.Get()
		defer redisPool.Close(taskRc)
		if err != nil {
			log.Printf("get redis connection fails %s", err)
			return
		}
		taskQueue := task.New()
		taskQueue.SetRClient(RANKING_TASK_QUEUE, taskRc)
		taskQueue.Worker(2, true, &ProxyGet{&cfg, logger, redisPool})
	}
	if *threadHandlerPtr {
		rc, err := redisPool.Get()
		defer redisPool.Close(rc)
		if err != nil {
			log.Printf("get redis connection fails %s", err)
			return
		}
		queue := thread.New()
		queue.SetRequestUri(cfg.TaskNUri)
		queue.Worker(2, true, &Control{&cfg, logger, redisPool}, &Submit{&cfg, logger, redisPool})
	}
	done := make(chan bool)
	<-done
}