package main

import (
	"database/sql"
	"encoding/json"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/martini-contrib/sessions"
	"log"
	"net/http"
)

type SpStatUser struct {
	Id       int64
	UserName string
	Password string
	Role     *SpStatRole
	Access   *SpStatAccess
}

type SpStatRole struct {
	Id        int64
	Name      string
	Privilege int64
	Menu      int64
}

type SpStatAccess struct {
	Id    int64
	Group []string
	Rule  int64
}

type SpStatNode struct {
	Id   int64
	Name string
	Node string
}

type SpStatMenu struct {
	Id    int64
	Title string
	Name  string
}

type IFilter interface {
	Filter(params ...interface{})
}

type RBAC struct{}

const (
	SESSION_KEY_QUSER = "Quser"
	ERROR_PAGE_NAME   = "/500.html"
	LOGIN_PAGE_NAME   = "/login"
	NODE_LOGIN_PRI    = 1 << 0
	NODE_RLOGIN_PRI   = 1 << 2

	ROLE_GUESS_ID = 1 << 0

	ROLE_GUESS_PRI = NODE_LOGIN_PRI | NODE_RLOGIN_PRI
)
const (
	GROUP_PRI_ALL = 1 << iota
	GROUP_PRI_ALLOW
	GROUP_PRI_BAN
)

func (self *RBAC) Filter() martini.Handler {
	return func(r *http.Request, w http.ResponseWriter, log *log.Logger, db *sql.DB, session sessions.Session, nMap map[string]*SpStatNode) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		path := r.URL.Path

		user := SpStatUser{}

		value := session.Get(SESSION_KEY_QUSER)
		if value == nil {
			role := SpStatRole{}
			role.Id = ROLE_GUESS_ID
			role.Privilege = ROLE_GUESS_PRI
			user.Role = &role
		} else {

			if v, ok := value.([]byte); ok {
				json.Unmarshal(v, &user)
			}
		}
		if v, ok := nMap[path]; ok {
			if (user.Role.Privilege & v.Id) != v.Id {
				http.Redirect(w, r, LOGIN_PAGE_NAME, 301)
				return
			}
			return
		}
		log.Printf("access unauth page")
		http.Redirect(w, r, LOGIN_PAGE_NAME, 301)
	}
}
