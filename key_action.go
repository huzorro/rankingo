package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/huzorro/spfactor/sexredis"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	RANKING_KEYWORD_QUEUE     = "ranking:keyword:queue"
	RANKING_KEYWORD_HASH      = "ranking:keyword:hash"
	RANKING_TASK_RESULT_QUEUE = "ranking:task:result"
)
const (
	RANKING_STATUS_START = iota
	RANKING_STATUS_CANCEL
)

type KeyMsg struct {
	Id          int64  `json:"id"`
	Uid         int64  `json"uid,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Keyword     string `json:"keyword"`
	Destlink    string `json:"destlink"`
	KeyCity     string `json:"keyCity,omitempty"`
	KeyProvince string `json:"keyProvince,omitempty"`
	Status      int64  `json:"status,omitempty`
	Logtime     string `json:"logtime,omitempty"`
}

type RankPay struct {
	User    SpStatUser
	Id      int64
	Balance int64
	Logtime string
}

type RankPayLog struct {
	User    SpStatUser
	Id      int64
	Balance int64
	Remark  string
	Logtime string
}

type RankConsumeLog struct {
	User    SpStatUser
	Keyword NormMsg
	Id      int64
	Balance int64
	Logtime string
}

type Status struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}

type PageResult struct {
	Result
	Norms []*NormMsg
}

type PayLogPageResult struct {
	Result
	RankPayLogs []*RankPayLog
}

type ConsumeLogPageResult struct {
	Result
	RankConsumeLogs []*RankConsumeLog
}

type UserRelation struct {
	User   SpStatUser
	Role   SpStatRole
	Access SpStatAccess
	Pay    RankPay
	KeyN   int64
}

type UserRelationPageResult struct {
	Result
	UserRelations []*UserRelation
}

func logout(r *http.Request, w http.ResponseWriter, log *log.Logger, session sessions.Session) {
	session.Clear()
	http.Redirect(w, r, LOGIN_PAGE_NAME, 301)
}
func loginCheck(r *http.Request, w http.ResponseWriter, log *log.Logger, db *sql.DB, session sessions.Session) (int, string) {
	//cross domain
	w.Header().Set("Access-Control-Allow-Origin", "*")
	un := r.PostFormValue("username")
	pd := r.PostFormValue("password")
	var (
		s Status
	)

	stmtOut, err := db.Prepare(`SELECT a.id, a.username, a.password, a.roleid, b.name, b.privilege, b.menu, 
		a.accessid, c.pri_group, c.pri_rule FROM sp_user a 
		INNER JOIN sp_role b ON a.roleid = b.id 
		INNER JOIN sp_access_privilege c ON a.accessid = c.id 
		WHERE username = ? AND password = ? `)

	if err != nil {
		log.Printf("get login user fails %s", err)
		s = Status{"500", "内部错误导致登录失败."}
		rs, _ := json.Marshal(s)
		return http.StatusOK, string(rs)
	}
	result, err := stmtOut.Query(un, pd)
	defer func() {
		stmtOut.Close()
		result.Close()
	}()
	if err != nil {
		log.Printf("%s", err)
		//		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		s = Status{"500", "内部错误导致登录失败."}
		rs, _ := json.Marshal(s)
		return http.StatusOK, string(rs)
	}
	if result.Next() {
		u := SpStatUser{}
		u.Role = &SpStatRole{}
		u.Access = &SpStatAccess{}
		var g string
		if err := result.Scan(&u.Id, &u.UserName, &u.Password, &u.Role.Id, &u.Role.Name, &u.Role.Privilege, &u.Role.Menu, &u.Access.Id, &g, &u.Access.Rule); err != nil {
			log.Printf("%s", err)
			s = Status{"500", "内部错误导致登录失败."}
			rs, _ := json.Marshal(s)
			return http.StatusOK, string(rs)
		} else {
			u.Access.Group = strings.Split(g, ";")
			//
			uSession, _ := json.Marshal(u)
			session.Set(SESSION_KEY_QUSER, uSession)
			s = Status{"200", "登录成功"}
			rs, _ := json.Marshal(s)
			return http.StatusOK, string(rs)
		}

	} else {
		log.Printf("%s", err)
		s = Status{"403", "登录失败,用户名/密码错误"}
		rs, _ := json.Marshal(s)
		return http.StatusOK, string(rs)
	}

}

func keyOneAction(r *http.Request, w http.ResponseWriter, db *sql.DB,
	log *log.Logger, session sessions.Session, render render.Render) (int, string) {
	var (
		nm   *NormMsg
		js   []byte
		user SpStatUser
		con  string
		s    Status
	)
	r.ParseForm()
	value := session.Get(SESSION_KEY_QUSER)
	if v, ok := value.([]byte); ok {
		json.Unmarshal(v, &user)
	} else {
		log.Printf("session stroe type error")
		s = Status{"500", "操作失败"}
		rs, _ := json.Marshal(s)
		return http.StatusOK, string(rs)
	}

	switch user.Access.Rule {
	case GROUP_PRI_ALL:
	case GROUP_PRI_ALLOW:
		con = "uid IN(" + strings.Join(user.Access.Group, ",") + ") AND "
	case GROUP_PRI_BAN:
		con = "uid NOT IN(" + strings.Join(user.Access.Group, ",") + ") AND "
	default:
		log.Printf("group private erros")
	}
	stmtOut, err := db.Prepare(`SELECT id, owner, keyword, destlink, history_order, 
	 current_order, history_index, current_index, city_key, province_key, cost, status, logtime 
	FROM ranking_detail WHERE ` + con + " id = ?")
	defer stmtOut.Close()
	id, _ := strconv.Atoi(r.PostFormValue("Id"))
	rows, err := stmtOut.Query(id)
	defer rows.Close()
	if err != nil {
		log.Printf("%s", err)
		s = Status{"500", "操作失败"}
		rs, _ := json.Marshal(s)
		return http.StatusOK, string(rs)
	}
	if rows.Next() {
		nm = &NormMsg{}
		if err := rows.Scan(&nm.KeyMsg.Id, &nm.KeyMsg.Owner, &nm.KeyMsg.Keyword, &nm.KeyMsg.Destlink,
			&nm.HOrder, &nm.COrder, &nm.HIndex, &nm.CIndex, &nm.KeyMsg.KeyCity,
			&nm.KeyMsg.KeyProvince, &nm.Cost, &nm.KeyMsg.Status, &nm.KeyMsg.Logtime); err != nil {
			log.Printf("%s", err)
			s = Status{"500", "操作失败"}
			rs, _ := json.Marshal(s)
			return http.StatusOK, string(rs)
		}
		if nm.KeyMsg.Status == RANKING_STATUS_CANCEL {
			nm.Cancel = true
		}
	}

	if js, err = json.Marshal(nm); err != nil {
		log.Printf("json Marshal fails %s", err)
		s = Status{"500", "内部错误导致登录失败."}
		rs, _ := json.Marshal(s)
		return http.StatusOK, string(rs)
	}
	return http.StatusOK, string(js)
}
func keyShowAction(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg, session sessions.Session, ms []*SpStatMenu, render render.Render) {
	var (
		nm     *NormMsg
		nms    []*NormMsg
		menu   []*SpStatMenu
		user   SpStatUser
		con    string
		totalN int64
		pr     *PageResult
		destPn int64
		money  int64
	)
	path := r.URL.Path
	r.ParseForm()
	value := session.Get(SESSION_KEY_QUSER)

	if v, ok := value.([]byte); ok {
		json.Unmarshal(v, &user)
	} else {
		log.Printf("session stroe type error")
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}

	switch user.Access.Rule {
	case GROUP_PRI_ALL:
		if uid := r.FormValue("uid"); uid != "" {
			con = "WHERE uid IN(" + uid + ")"
		}
	case GROUP_PRI_ALLOW:
		con = "WHERE uid IN(" + strings.Join(user.Access.Group, ",") + ")"
	case GROUP_PRI_BAN:
		con = "WHERE uid NOT IN(" + strings.Join(user.Access.Group, ",") + ")"
	default:
		log.Printf("group private erros")
	}

	for _, elem := range ms {
		if (user.Role.Menu & elem.Id) == elem.Id {
			menu = append(menu, elem)
		}
	}
	stmtOut, err := db.Prepare("SELECT COUNT(*) FROM ranking_detail " + con)
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	row := stmtOut.QueryRow()
	if err = row.Scan(&totalN); err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	//page
	if r.URL.Query().Get("p") != "" {
		destPn, _ = strconv.ParseInt(r.URL.Query().Get("p"), 10, 64)
	} else {
		destPn = 1
	}
	details := make(Details, totalN)
	result := Result{Data: make(Details, cfg.PageSize)}
	details.Page(int(destPn), &result)

	stmtOut, err = db.Prepare(`SELECT id, owner, keyword, destlink, history_order, 
	 current_order, history_index, current_index, city_key, province_key, cost, status, logtime 
	FROM ranking_detail ` + con + " ORDER BY logtime DESC LIMIT ?, ?")

	defer stmtOut.Close()
	rows, err := stmtOut.Query(cfg.PageSize*(destPn-1), cfg.PageSize)
	defer rows.Close()
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	for rows.Next() {
		nm = &NormMsg{}
		if err := rows.Scan(&nm.KeyMsg.Id, &nm.KeyMsg.Owner, &nm.KeyMsg.Keyword, &nm.KeyMsg.Destlink,
			&nm.HOrder, &nm.COrder, &nm.HIndex, &nm.CIndex, &nm.KeyMsg.KeyCity,
			&nm.KeyMsg.KeyProvince, &nm.Cost, &nm.KeyMsg.Status, &nm.KeyMsg.Logtime); err != nil {
			log.Printf("%s", err)
			http.Redirect(w, r, ERROR_PAGE_NAME, 301)
			return
		}
		if nm.KeyMsg.Status == RANKING_STATUS_CANCEL {
			nm.Cancel = true
		}
		nms = append(nms, nm)
	}
	pr = &PageResult{}
	pr.Result = result
	pr.Norms = make([]*NormMsg, pr.CurrentTotal)
	if totalN > 0 {
		copy(pr.Norms, nms)
	}
	paginator := NewPaginator(r, cfg.PageSize, totalN)
	//余额
	stmtOutPay, err := db.Prepare("SELECT IFNULL(SUM(balance),0) FROM ranking_pay " + con)
	defer stmtOutPay.Close()
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	row = stmtOutPay.QueryRow()
	if err := row.Scan(&money); err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	ret := struct {
		Menu      []*SpStatMenu
		Result    *PageResult
		Paginator *Paginator
		User      *SpStatUser
		Money     int64
	}{menu, pr, paginator, &user, money}

	if path == "/" {
		render.HTML(200, menu[0].Name, ret)
	} else {
		index := strings.LastIndex(path, "/")
		render.HTML(200, path[index+1:], ret)
	}
}

func keyUpdateAction(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg, session sessions.Session) (int, string) {
	var (
		key    KeyMsg
		oneKey KeyMsg
		js     []byte
		user   SpStatUser
	)
	r.ParseForm()
	rType := reflect.TypeOf(&key).Elem()
	rValue := reflect.ValueOf(&key).Elem()
	for i := 0; i < rType.NumField(); i++ {
		fN := rType.Field(i).Name
		p, _ := url.QueryUnescape(strings.TrimSpace(r.PostFormValue(fN)))
		switch rType.Field(i).Type.Kind() {
		case reflect.String:
			rValue.FieldByName(fN).SetString(p)
		case reflect.Int64:
			in, _ := strconv.ParseInt(p, 10, 64)
			rValue.FieldByName(fN).SetInt(in)
		case reflect.Float64:
			fl, _ := strconv.ParseFloat(p, 64)
			rValue.FieldByName(fN).SetFloat(fl)
		default:
			log.Printf("unknow type %s", p)
		}
	}
	//get session
	value := session.Get(SESSION_KEY_QUSER)

	if v, ok := value.([]byte); ok {
		json.Unmarshal(v, &user)
	} else {
		log.Printf("session stroe type error")
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}

	stmtIn, err := db.Prepare("UPDATE ranking_detail SET owner = ?, status = ? WHERE id = ? AND uid = ?")
	defer stmtIn.Close()
	if err != nil {
		log.Printf("%s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if _, err := stmtIn.Exec(key.Owner, key.Status, key.Id, user.Id); err != nil {
		log.Printf("%s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}

	//update ranking:keyword:hash
	redisClient, err := redisPool.Get()
	defer redisPool.Close(redisClient)
	if err != nil {
		log.Printf("get connection of redis pool %s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}

	onejs, err := redisClient.HGet(RANKING_KEYWORD_HASH, "id:"+strconv.FormatInt(key.Id, 10))
	if err != nil {
		log.Printf("get one %s %s fails %s", RANKING_KEYWORD_HASH, "id:"+strconv.FormatInt(key.Id, 10), err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if err := json.Unmarshal([]byte(onejs), &oneKey); err != nil {
		log.Printf("json Unmarshal fails %s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}

	oneKey.Owner = key.Owner
	oneKey.Status = key.Status

	if js, err = json.Marshal(oneKey); err != nil {
		log.Printf("json Marshal fails %s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}

	if _, err = redisClient.HMSet(RANKING_KEYWORD_HASH, "id:"+strconv.FormatInt(key.Id, 10), string(js)); err != nil {
		log.Printf("update in  %s fails %s", RANKING_KEYWORD_HASH, err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	js, _ = json.Marshal(Status{"200", "操作成功"})
	return http.StatusOK, string(js)
}

func keyAddAction(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg, session sessions.Session) (int, string) {
	var (
		key  KeyMsg
		user SpStatUser
		n    int64
		js   []byte
	)
	r.ParseForm()
	rType := reflect.TypeOf(&key).Elem()
	rValue := reflect.ValueOf(&key).Elem()
	for i := 0; i < rType.NumField(); i++ {
		fN := rType.Field(i).Name
		p, _ := url.QueryUnescape(strings.TrimSpace(r.PostFormValue(fN)))
		switch rType.Field(i).Type.Kind() {
		case reflect.String:
			rValue.FieldByName(fN).SetString(p)
		case reflect.Int64:
			in, _ := strconv.ParseInt(p, 10, 64)
			rValue.FieldByName(fN).SetInt(in)
		case reflect.Float64:
			fl, _ := strconv.ParseFloat(p, 64)
			rValue.FieldByName(fN).SetFloat(fl)
		default:
			log.Printf("unknow type %s", p)
		}
	}
	// get session

	value := session.Get(SESSION_KEY_QUSER)
	if v, ok := value.([]byte); ok {
		json.Unmarshal(v, &user)
	} else {
		log.Printf("session stroe type error")
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	key.Uid = user.Id

	stmtOut, err := db.Prepare("SELECT COUNT(*) FROM ranking_detail WHERE keyword = ? AND destlink = ?")
	defer stmtOut.Close()
	if err != nil {
		log.Printf("%s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	row := stmtOut.QueryRow(key.Keyword, key.Destlink)

	if err := row.Scan(&n); err != nil {
		log.Printf("%s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if n > 0 {
		js, _ = json.Marshal(Status{"201", "操作失败, 添加了重复数据"})
		return http.StatusOK, string(js)
	}

	stmtOutSec, err := db.Prepare("SELECT COUNT(*) FROM ranking_keyword WHERE msg = ?")
	defer stmtOutSec.Close()
	if err != nil {
		log.Printf("%s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if js, err = json.Marshal(key); err != nil {
		log.Printf("json Marshal fails %s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	row = stmtOutSec.QueryRow(js)
	if err := row.Scan(&n); err != nil {
		log.Printf("%s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if n > 0 {
		js, _ = json.Marshal(Status{"201", "操作失败, 添加了重复数据"})
		return http.StatusOK, string(js)
	}

	stmtIn, err := db.Prepare("INSERT INTO ranking_keyword (msg) VALUES(?)")
	defer stmtIn.Close()
	if err != nil {
		log.Printf("%s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	//	key.Logtime = time.Now().Format("2006-01-02 15:04:05")

	//	if js, err = json.Marshal(key); err != nil {
	//		log.Printf("json Marshal fails %s", err)
	//		js, _ = json.Marshal(Status{"201", "操作失败"})
	//		return http.StatusOK, string(js)
	//	}

	result, err := stmtIn.Exec(js)
	if err != nil {
		log.Printf("%s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	key.Id, _ = result.LastInsertId()
	redisClient, err := redisPool.Get()
	defer redisPool.Close(redisClient)
	if err != nil {
		log.Printf("get connection of redis pool %s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if js, err = json.Marshal(key); err != nil {
		log.Printf("json Marshal fails %s", err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}

	if _, err = redisClient.HMSet(RANKING_KEYWORD_HASH, "id:"+strconv.FormatInt(key.Id, 10), string(js)); err != nil {
		log.Printf("put in  %s fails %s", RANKING_KEYWORD_HASH, err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}

	queue := sexredis.New()
	queue.SetRClient(RANKING_KEYWORD_QUEUE, redisClient)

	if _, err := queue.Put(js); err != nil {
		log.Printf("put in %s fails %s", RANKING_KEYWORD_QUEUE, err)
		js, _ = json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	js, _ = json.Marshal(Status{"200", "操作成功"})
	return http.StatusOK, string(js)
}

func taskNumberApi(r *http.Request, w http.ResponseWriter, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg) (int, string) {
	redisClient, err := redisPool.Get()
	defer redisPool.Close(redisClient)
	if err != nil {
		log.Printf("get connection of redis pool fails %s", err)
		return http.StatusOK, "0"
	}
	if n, err := redisClient.LLen(RANKING_TASK_QUEUE); err != nil {
		return http.StatusOK, "0"
	} else {
		return http.StatusOK, strconv.FormatInt(n, 10)
	}

}
func taskOnexApi(r *http.Request, w http.ResponseWriter, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg) (int, string) {
	redisClient, err := redisPool.Get()
	defer redisPool.Close(redisClient)
	if err != nil {
		log.Printf("get connection of redis pool %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	m, err := redisClient.LPop(RANKING_TASK_QUEUE)
	if err != nil {
		log.Printf("get out of queue elem fails %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if m != "" {
		//出队后, 分时数据递减, 重新放入task队列
		var msg TaskMsg
		if err := json.Unmarshal([]byte(m), &msg); err != nil {
			log.Printf("json Unmarshal fails %s", err)
			js, _ := json.Marshal(Status{"201", "操作失败"})
			return http.StatusOK, string(js)
		}
		h := fmt.Sprint(time.Now().Hour())

		msg.NormMsg.Hour[h] = msg.NormMsg.Hour[h] - 1
		if msg.NormMsg.Hour[h] < 0 {
			js, _ := json.Marshal(Status{"202", "该时段任务达标"})
			return http.StatusOK, string(js)
		}
		msg.InitTime = time.Now().UnixNano() / (1000 * 1000)
		js, _ := json.Marshal(msg)
		if _, err := redisClient.RPush(RANKING_TASK_QUEUE, js); err != nil {
			log.Printf("put end of the queue fails %s", err)
		}
		return http.StatusOK, string(m)
	} else {
		log.Printf("not found elem in queue %s", err)
		js, _ := json.Marshal(Status{"404", "没有发现任务"})
		return http.StatusOK, string(js)
	}

}
func taskOneApi(r *http.Request, w http.ResponseWriter, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg) (int, string) {
	redisClient, err := redisPool.Get()
	defer redisPool.Close(redisClient)
	if err != nil {
		log.Printf("get connection of redis pool %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	m, err := redisClient.LPop(RANKING_TASK_QUEUE)
	if err != nil {
		log.Printf("get out of queue elem fails %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if m != "" {
		//出队后, 分时数据递减, 取出NormMsg放入norm队列, 等待新的代理数据, 重新生成task任务
		var msg TaskMsg
		if err := json.Unmarshal([]byte(m), &msg); err != nil {
			log.Printf("json Unmarshal fails %s", err)
			js, _ := json.Marshal(Status{"201", "操作失败"})
			return http.StatusOK, string(js)
		}
		h := fmt.Sprint(time.Now().Hour())

		msg.NormMsg.Hour[h] = msg.NormMsg.Hour[h] - 1

		if msg.NormMsg.Hour[h] < 0 {
			js, _ := json.Marshal(Status{"202", "该时段任务达标"})
			return http.StatusOK, string(js)
		}

		js, _ := json.Marshal(msg.NormMsg)
		if msg.NormMsg.KeyMsg.KeyCity != "" || msg.NormMsg.KeyMsg.KeyProvince != "" {
			if _, err := redisClient.RPush(RANKING_AREA_NORM_QUEUE, js); err != nil {
				log.Printf("put end of the queue fails %s", err)
			}
		} else {
			if _, err := redisClient.RPush(RANKING_COMMON_NORM_QUEUE, js); err != nil {
				log.Printf("put end of the queue fails %s", err)
			}
		}

		return http.StatusOK, string(m)
	} else {
		log.Printf("not found elem in queue %s", err)
		js, _ := json.Marshal(Status{"404", "没有发现任务"})
		return http.StatusOK, string(js)
	}

}

func taskResultApi(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg, session sessions.Session) (int, string) {
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("json read fails %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	redisClient, err := redisPool.Get()
	defer redisPool.Close(redisClient)
	if err != nil {
		log.Printf("get connection of redis pool %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if _, err := redisClient.RPush(RANKING_TASK_RESULT_QUEUE, data); err != nil {
		log.Printf("put in %s fails %s", RANKING_TASK_RESULT_QUEUE, err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	} else {
		js, _ := json.Marshal(Status{"200", "操作成功"})
		return http.StatusOK, string(js)
	}
}

func proxyUaApi(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger) (int, string) {
	log.Printf("%s | %s", r.RemoteAddr, r.UserAgent())
	return http.StatusOK, r.RemoteAddr + ";" + r.UserAgent()
}
func payAdminAction(r *http.Request, w http.ResponseWriter, db *sql.DB,
	log *log.Logger, cfg *Cfg, session sessions.Session) (int, string) {
	r.ParseForm()
	if r.PostFormValue("balance") == "" || r.PostFormValue("uid") == "" || r.PostFormValue("remark") == "" {
		log.Printf("pay balance or uid is empty")
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	balance, err := strconv.Atoi(r.PostFormValue("balance"))
	uid, err := strconv.Atoi(r.PostFormValue("uid"))
	if err != nil {
		log.Printf("pay balance conversion failed %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	tx, err := db.Begin()

	stmtInLog, err := tx.Prepare("INSERT INTO ranking_pay_log (uid, balance, remark) VALUES(?, ?, ?)")
	defer stmtInLog.Close()
	if err != nil {
		log.Printf("insert into ranking pay log fails %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	_, err = stmtInLog.Exec(uid, balance*100, r.PostFormValue("remark"))
	if err != nil {
		tx.Rollback()
		log.Printf("insert into ranking pay log fails %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	stmtInPay, err := tx.Prepare("INSERT INTO ranking_pay (uid, balance) VALUES(?, ?) ON DUPLICATE KEY UPDATE balance = balance + VALUES(balance)")
	defer stmtInPay.Close()
	if err != nil {
		log.Printf("insert into ranking pay fails %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if _, err := stmtInPay.Exec(uid, balance*100); err != nil {
		tx.Rollback()
		log.Printf("insert into ranking pay fails %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Printf("tx commit fails %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	log.Printf("%s: %d succeed", r.PostFormValue("remark"), balance)
	js, _ := json.Marshal(Status{"200", "操作成功"})
	return http.StatusOK, string(js)

}

func payLogAction(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg, session sessions.Session, ms []*SpStatMenu, render render.Render) {
	var (
		paylog  *RankPayLog
		paylogs []*RankPayLog
		menu    []*SpStatMenu
		user    SpStatUser
		con     string
		totalN  int64
		pr      *PayLogPageResult
		destPn  int64
		money   int64
	)
	path := r.URL.Path
	r.ParseForm()
	value := session.Get(SESSION_KEY_QUSER)

	if v, ok := value.([]byte); ok {
		json.Unmarshal(v, &user)
	} else {
		log.Printf("session stroe type error")
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}

	switch user.Access.Rule {
	case GROUP_PRI_ALL:
		if uid := r.FormValue("uid"); uid != "" {
			con = "WHERE uid IN(" + uid + ")"
		}
	case GROUP_PRI_ALLOW:
		con = "WHERE uid IN(" + strings.Join(user.Access.Group, ",") + ")"
	case GROUP_PRI_BAN:
		con = "WHERE uid NOT IN(" + strings.Join(user.Access.Group, ",") + ")"
	default:
		log.Printf("group private erros")
	}

	for _, elem := range ms {
		if (user.Role.Menu & elem.Id) == elem.Id {
			menu = append(menu, elem)
		}
	}
	stmtOut, err := db.Prepare("SELECT COUNT(*) FROM ranking_pay_log " + con)
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	row := stmtOut.QueryRow()
	if err = row.Scan(&totalN); err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	//page
	if r.URL.Query().Get("p") != "" {
		destPn, _ = strconv.ParseInt(r.URL.Query().Get("p"), 10, 64)
	} else {
		destPn = 1
	}
	details := make(Details, totalN)
	result := Result{Data: make(Details, cfg.PageSize)}
	details.Page(int(destPn), &result)

	stmtOut, err = db.Prepare(`SELECT id, uid, balance, remark, logtime FROM ranking_pay_log ` + con + " ORDER BY id DESC LIMIT ?, ?")
	defer stmtOut.Close()
	rows, err := stmtOut.Query(cfg.PageSize*(destPn-1), cfg.PageSize)
	defer rows.Close()
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	for rows.Next() {
		paylog = &RankPayLog{}
		if err := rows.Scan(&paylog.Id, &paylog.User.Id, &paylog.Balance, &paylog.Remark, &paylog.Logtime); err != nil {
			log.Printf("%s", err)
			http.Redirect(w, r, ERROR_PAGE_NAME, 301)
			return
		}
		paylogs = append(paylogs, paylog)
	}
	pr = &PayLogPageResult{}
	pr.Result = result
	pr.RankPayLogs = make([]*RankPayLog, pr.CurrentTotal)
	if totalN > 0 {
		copy(pr.RankPayLogs, paylogs)
	}
	paginator := NewPaginator(r, cfg.PageSize, totalN)
	//余额
	stmtOutPay, err := db.Prepare("SELECT IFNULL(SUM(balance), 0) FROM ranking_pay " + con)
	defer stmtOutPay.Close()
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	row = stmtOutPay.QueryRow()
	if err := row.Scan(&money); err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	ret := struct {
		Menu      []*SpStatMenu
		Result    *PayLogPageResult
		Paginator *Paginator
		User      *SpStatUser
		Money     int64
	}{menu, pr, paginator, &user, money}

	index := strings.LastIndex(path, "/")
	render.HTML(200, path[index+1:], ret)
}

func consumeLogAction(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg, session sessions.Session, ms []*SpStatMenu, render render.Render) {
	var (
		consumelog  *RankConsumeLog
		consumelogs []*RankConsumeLog
		menu        []*SpStatMenu
		user        SpStatUser
		con         string
		totalN      int64
		pr          *ConsumeLogPageResult
		destPn      int64
		money       int64
	)
	path := r.URL.Path
	r.ParseForm()
	value := session.Get(SESSION_KEY_QUSER)

	if v, ok := value.([]byte); ok {
		json.Unmarshal(v, &user)
	} else {
		log.Printf("session stroe type error")
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}

	switch user.Access.Rule {
	case GROUP_PRI_ALL:
		if uid := r.FormValue("uid"); uid != "" {
			con = "WHERE uid IN(" + uid + ")"
		}
	case GROUP_PRI_ALLOW:
		con = "WHERE uid IN(" + strings.Join(user.Access.Group, ",") + ")"
	case GROUP_PRI_BAN:
		con = "WHERE uid NOT IN(" + strings.Join(user.Access.Group, ",") + ")"
	default:
		log.Printf("group private erros")
	}

	for _, elem := range ms {
		if (user.Role.Menu & elem.Id) == elem.Id {
			menu = append(menu, elem)
		}
	}
	stmtOut, err := db.Prepare("SELECT COUNT(*) FROM ranking_consume_log " + con)
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	row := stmtOut.QueryRow()
	if err = row.Scan(&totalN); err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	//page
	if r.URL.Query().Get("p") != "" {
		destPn, _ = strconv.ParseInt(r.URL.Query().Get("p"), 10, 64)
	} else {
		destPn = 1
	}
	details := make(Details, totalN)
	result := Result{Data: make(Details, cfg.PageSize)}
	details.Page(int(destPn), &result)

	stmtOut, err = db.Prepare(`SELECT id, uid, kid, balance, logtime FROM ranking_consume_log ` + con + " ORDER BY id DESC LIMIT ?, ?")
	defer stmtOut.Close()
	rows, err := stmtOut.Query(cfg.PageSize*(destPn-1), cfg.PageSize)
	defer rows.Close()
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	for rows.Next() {
		consumelog = &RankConsumeLog{}
		if err := rows.Scan(&consumelog.Id, &consumelog.User.Id, &consumelog.Keyword.KeyMsg.Id, &consumelog.Balance, &consumelog.Logtime); err != nil {
			log.Printf("%s", err)
			http.Redirect(w, r, ERROR_PAGE_NAME, 301)
			return
		}
		consumelogs = append(consumelogs, consumelog)
	}
	pr = &ConsumeLogPageResult{}
	pr.Result = result
	pr.RankConsumeLogs = make([]*RankConsumeLog, pr.CurrentTotal)
	if totalN > 0 {
		copy(pr.RankConsumeLogs, consumelogs)
	}
	paginator := NewPaginator(r, cfg.PageSize, totalN)
	//余额
	stmtOutPay, err := db.Prepare("SELECT IFNULL(SUM(balance), 0) FROM ranking_pay " + con)
	defer stmtOutPay.Close()
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	row = stmtOutPay.QueryRow()
	if err := row.Scan(&money); err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	ret := struct {
		Menu      []*SpStatMenu
		Result    *ConsumeLogPageResult
		Paginator *Paginator
		User      *SpStatUser
		Money     int64
	}{menu, pr, paginator, &user, money}

	index := strings.LastIndex(path, "/")
	render.HTML(200, path[index+1:], ret)
}

func viewUsersAction(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg, session sessions.Session, ms []*SpStatMenu, render render.Render) {
	var (
		userRelation  *UserRelation
		userRelations []*UserRelation
		menu          []*SpStatMenu
		user          SpStatUser
		con           string
		totalN        int64
		pr            *UserRelationPageResult
		destPn        int64
		money         int64
	)
	path := r.URL.Path
	r.ParseForm()
	value := session.Get(SESSION_KEY_QUSER)

	if v, ok := value.([]byte); ok {
		json.Unmarshal(v, &user)
	} else {
		log.Printf("session stroe type error")
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}

	switch user.Access.Rule {
	case GROUP_PRI_ALL:
	case GROUP_PRI_ALLOW:
		con = "WHERE id IN(" + strings.Join(user.Access.Group, ",") + ")"
	case GROUP_PRI_BAN:
		con = "WHERE id NOT IN(" + strings.Join(user.Access.Group, ",") + ")"
	default:
		log.Printf("group private erros")
	}

	for _, elem := range ms {
		if (user.Role.Menu & elem.Id) == elem.Id {
			menu = append(menu, elem)
		}
	}
	stmtOut, err := db.Prepare("SELECT COUNT(*) FROM sp_user " + con)
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	row := stmtOut.QueryRow()
	if err = row.Scan(&totalN); err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	//page
	if r.URL.Query().Get("p") != "" {
		destPn, _ = strconv.ParseInt(r.URL.Query().Get("p"), 10, 64)
	} else {
		destPn = 1
	}
	details := make(Details, totalN)
	result := Result{Data: make(Details, cfg.PageSize)}
	details.Page(int(destPn), &result)

	stmtOut, err = db.Prepare(`SELECT a.id, a.username, a.password, a.roleid, a.accessid,  
			b.name, c.pri_rule, IFNULL(d.balance, 0) AS money, 
			COUNT(e.id) AS number FROM sp_user a LEFT JOIN sp_role b ON a.roleid = b.id 
			LEFT JOIN sp_access_privilege c  ON a.accessid = c.id 
			LEFT JOIN ranking_pay d ON a.id = d.uid LEFT JOIN 
			ranking_detail e ON a.id = e.uid ` + con + " GROUP BY a.id ORDER BY id DESC LIMIT ?, ?")
	defer stmtOut.Close()
	rows, err := stmtOut.Query(cfg.PageSize*(destPn-1), cfg.PageSize)
	defer rows.Close()
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	for rows.Next() {
		userRelation = &UserRelation{}
		if err := rows.Scan(&userRelation.User.Id, &userRelation.User.UserName, &userRelation.User.Password,
			&userRelation.Role.Id, &userRelation.Access.Id, &userRelation.Role.Name,
			&userRelation.Access.Rule, &userRelation.Pay.Balance, &userRelation.KeyN); err != nil {
			log.Printf("%s", err)
			http.Redirect(w, r, ERROR_PAGE_NAME, 301)
			return
		}
		userRelations = append(userRelations, userRelation)
	}

	pr = &UserRelationPageResult{}
	pr.Result = result
	pr.UserRelations = make([]*UserRelation, pr.CurrentTotal)
	if totalN > 0 {
		copy(pr.UserRelations, userRelations)
	}
	paginator := NewPaginator(r, cfg.PageSize, totalN)
	//余额
	stmtOutPay, err := db.Prepare("SELECT IFNULL(SUM(balance), 0) FROM ranking_pay " + con)
	defer stmtOutPay.Close()
	if err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	row = stmtOutPay.QueryRow()
	if err := row.Scan(&money); err != nil {
		log.Printf("%s", err)
		http.Redirect(w, r, ERROR_PAGE_NAME, 301)
		return
	}
	ret := struct {
		Menu      []*SpStatMenu
		Result    *UserRelationPageResult
		Paginator *Paginator
		User      *SpStatUser
		Money     int64
	}{menu, pr, paginator, &user, money}

	index := strings.LastIndex(path, "/")
	render.HTML(200, path[index+1:], ret)
}

func editUserAction(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg, session sessions.Session, ms []*SpStatMenu, render render.Render) (int, string) {
	var (
		user SpStatUser
	)
	r.ParseForm()
	id, err := url.QueryUnescape(strings.TrimSpace(r.PostFormValue("id")))
	user.Id, _ = strconv.ParseInt(id, 10, 64)
	user.UserName, err = url.QueryUnescape(strings.TrimSpace(r.PostFormValue("userName")))
	user.Password, err = url.QueryUnescape(strings.TrimSpace(r.PostFormValue("password")))
	if err != nil {
		log.Printf("post param parse fails %s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	stmtIn, err := db.Prepare("UPDATE sp_user SET username=?, password=? WHERE id = ?")
	defer stmtIn.Close()
	if _, err = stmtIn.Exec(user.UserName, user.Password, user.Id); err != nil {
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	js, _ := json.Marshal(Status{"200", "操作成功"})
	return http.StatusOK, string(js)
}

func viewUserAction(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg, session sessions.Session, ms []*SpStatMenu, render render.Render) (int, string) {
	var (
		user SpStatUser
	)
	r.ParseForm()
	id, err := url.QueryUnescape(strings.TrimSpace(r.PostFormValue("id")))
	user.Id, _ = strconv.ParseInt(id, 10, 64)
	stmtOut, err := db.Prepare("SELECT username, password FROM sp_user WHERE id = ?")
	defer stmtOut.Close()
	row := stmtOut.QueryRow(user.Id)
	err = row.Scan(&user.UserName, &user.Password)
	if err != nil {
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if js, err := json.Marshal(user); err != nil {
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	} else {
		return http.StatusOK, string(js)
	}
}

func addUserAction(r *http.Request, w http.ResponseWriter, db *sql.DB, log *log.Logger,
	redisPool *sexredis.RedisPool, cfg *Cfg, session sessions.Session, ms []*SpStatMenu, render render.Render) (int, string) {
	r.ParseForm()
	userName, err := url.QueryUnescape(strings.TrimSpace(r.PostFormValue("userName")))
	password, err := url.QueryUnescape(strings.TrimSpace(r.PostFormValue("password")))
	tx, err := db.Begin()
	if err != nil {
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	stmtIn, err := tx.Prepare("INSERT INTO sp_user (username, password) VALUES(?, ?)")
	defer stmtIn.Close()

	if err != nil {
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}

	result, err := stmtIn.Exec(userName, password)
	if err != nil {
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	stmtInAccess, err := tx.Prepare("INSERT INTO sp_access_privilege(pri_group) VALUES(?)")
	defer stmtInAccess.Close()
	if err != nil {
		tx.Rollback()
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	result, err = stmtInAccess.Exec(id)
	if err != nil {
		tx.Rollback()
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	accessId, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	stmtInUpdate, err := tx.Prepare("UPDATE sp_user SET accessid = ? WHERE id = ?")
	defer stmtInUpdate.Close()
	if err != nil {
		tx.Rollback()
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	_, err = stmtInUpdate.Exec(accessId, id)

	if err != nil {
		tx.Rollback()
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		log.Printf("%s", err)
		js, _ := json.Marshal(Status{"201", "操作失败"})
		return http.StatusOK, string(js)
	}
	js, _ := json.Marshal(Status{"200", "操作成功"})
	return http.StatusOK, string(js)
}
