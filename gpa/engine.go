package gpa

import "github.com/jmoiron/sqlx"

var engine *Engine

type Engine struct {
	DB           *sqlx.DB
	tableHashMap map[any]string
}

func (e Engine) SetTableName(entity any, tableName string) {
	e.tableHashMap[entity] = tableName
}

func (e Engine) GetTableName(entity any) (string, bool) {
	val, ok := e.tableHashMap[entity]
	return val, ok
}

func NewEngine(db *sqlx.DB) {
	engine = &Engine{
		DB:           db,
		tableHashMap: make(map[any]string, 0),
	}
}
