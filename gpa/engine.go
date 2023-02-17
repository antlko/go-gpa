package gpa

import "github.com/jmoiron/sqlx"

var engine *Engine

type DbProviderI interface {
	sqlx.Ext
	sqlx.Preparer
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
}

type Engine struct {
	entityTableNameMap map[any]string
	tableNameEntityMap map[string]any
	db                 *sqlx.DB
	t                  *sqlx.Tx
}

func (e *Engine) GetInstance() DbProviderI {
	if e.t != nil {
		return e.t
	}
	return e.db
}

func (e *Engine) SetTableName(entity any, tableName string) {
	e.entityTableNameMap[entity] = tableName
	e.tableNameEntityMap[tableName] = entity
}

func (e *Engine) GetTableName(entity any) (string, bool) {
	val, ok := e.entityTableNameMap[entity]
	return val, ok
}

func (e *Engine) GetEntity(tableName string) (any, bool) {
	val, ok := e.tableNameEntityMap[tableName]
	return val, ok
}

func NewEngine(db *sqlx.DB) {
	engine = &Engine{
		entityTableNameMap: make(map[any]string, 0),
		tableNameEntityMap: make(map[string]any, 0),
		db:                 db,
	}
}
