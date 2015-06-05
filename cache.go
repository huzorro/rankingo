package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/huzorro/spfactor/sexredis"
)

type Cache struct {
	db   *sql.DB
	pool *sexredis.RedisPool
}

func (self *Cache) RbacNodeToMap() (map[string]*SpStatNode, error) {
	stmtOut, err := self.db.Prepare("SELECT id, name, node FROM sp_node_privilege")
	defer stmtOut.Close()
	if err != nil {
		return nil, err
	}
	result, err := stmtOut.Query()
	defer result.Close()
	if err != nil {
		return nil, err
	}
	nMap := make(map[string]*SpStatNode)
	for result.Next() {
		node := &SpStatNode{}

		if err := result.Scan(&node.Id, &node.Name, &node.Node); err != nil {
			return nil, err
		} else {
			nMap[node.Node] = node
		}
	}
	return nMap, nil
}

func (self *Cache) RbacMenuToSlice() ([]*SpStatMenu, error) {
	stmtOut, err := self.db.Prepare("SELECT id, title, name FROM sp_menu_template")
	defer stmtOut.Close()
	if err != nil {
		return nil, err
	}
	result, err := stmtOut.Query()
	defer result.Close()
	if err != nil {
		return nil, err
	}
	var ms []*SpStatMenu
	for result.Next() {
		menu := &SpStatMenu{}

		if err := result.Scan(&menu.Id, &menu.Title, &menu.Name); err != nil {
			return nil, err
		} else {
			ms = append(ms, menu)
		}
	}
	return ms, nil
}
