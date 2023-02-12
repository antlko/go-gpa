package gpa

import (
	"fmt"
	pluralize "github.com/gertd/go-pluralize"
	"github.com/jackc/pgconn"
	"reflect"
	"strings"
)

type GPAEntity interface {
	GPAConfigure(e *Engine)
}

func From[entityType any]() Entity[entityType] {
	entityObject := *new(entityType)
	_, ok := engine.tableHashMap[entityObject]
	if !ok {
		initTable(engine, entityObject, entityObject)
	}

	return Entity[entityType]{
		entityObj: entityObject,
	}
}

func initTable(e *Engine, obj any, entity interface{}) {
	gpaEntity, ok := entity.(GPAEntity)
	if ok {
		gpaEntity.GPAConfigure(e)
	}

	tableName, ok := e.GetTableName(gpaEntity)
	if !ok {
		structName := strings.ToLower(reflect.TypeOf(obj).Name())
		tableName = pluralize.NewClient().Plural(structName)

		if !isTableExists(e, tableName) {
			createTable(e, entity, tableName)
		}
	}
	e.tableHashMap[obj] = tableName
}

func isTableExists(e *Engine, name string) bool {
	row := e.DB.QueryRow("SELECT 1 FROM " + name)
	switch row.Err().(type) {
	case *pgconn.PgError:
		if row.Err().(*pgconn.PgError).Code == "42P01" {
			return false
		}
	}
	if row.Err() != nil {
		panic(row.Err())
	}
	return true
}

func createTable(e *Engine, entity interface{}, tableName string) {
	names, _, types := getReflectedData(entity)

	fieldsData := ""
	for i := 0; i < len(names); i++ {
		pgType := getPGType(types[i])
		fieldsData += fmt.Sprintf("%s %s,", names[i], pgType)
	}
	fieldsData = fieldsData[:len(fieldsData)-1]

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id SERIAL,%s);", tableName, fieldsData)
	row := e.DB.QueryRow(query)
	if row.Err() != nil {
		panic(fmt.Sprintf("gpa can't create the table with error:%s", row.Err().Error()))
	}
}

func getPGType(tp string) string {
	tp = strings.ToLower(tp)
	null := ""
	if !strings.Contains(tp, "null") {
		null = " NOT NULL "
	}

	outType := ""
	if strings.Contains(tp, "time") {
		outType = " DATE "
	}
	if strings.Contains(tp, "int") {
		outType = " INTEGER "
	}
	if strings.Contains(tp, "string") {
		outType = " TEXT "
	}
	return outType + null
}
