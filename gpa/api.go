package gpa

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"log"
	"reflect"
	"strings"
)

type Entity[entityType any] struct {
	entityObj any
}

func (e *Entity[entityType]) Get(where string, args ...interface{}) (entityType, error) {
	entity := e.entityObj.(entityType)

	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return entity, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	if where != "" {
		where = " WHERE " + where
	}
	if err := engine.GetInstance().Get(&entity, "SELECT * FROM "+tableName+where, args...); err != nil {
		return entity, err
	}
	return entity, nil
}

func (e *Entity[entityType]) Select(where string, args ...interface{}) ([]entityType, error) {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return nil, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	entity := make([]entityType, 0)
	if where != "" {
		where = " WHERE " + where
	}
	if err := engine.GetInstance().Select(&entity, "SELECT * FROM "+tableName+where, args...); err != nil {
		return nil, err
	}
	return entity, nil
}

func (e *Entity[entityType]) FindByID(id int64) (entityType, error) {
	entity := e.entityObj.(entityType)
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return entity, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	if err := engine.GetInstance().Get(&entity, "SELECT * FROM "+tableName+" WHERE id = $1", id); err != nil {
		return entity, err
	}

	return e.withsLazy(entity)
}

func (e *Entity[entityType]) FindBy(filters []F, p *Pagination) ([]entityType, error) {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return nil, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	whereElements := " WHERE "
	paramsCounter := 1
	values := make([]interface{}, 0)
	for i := 0; i < len(filters); i++ {
		filter := filters[i]
		where := fmt.Sprintf(" %s %s ", filter.FieldName, filter.Sign)

		whereElements += fmt.Sprintf("%s$%d %s ", where, paramsCounter, filter.Cond)
		values = append(values, filter.Value)
		paramsCounter++
	}

	query := "SELECT * FROM " + tableName + whereElements + e.getPagQuery(p)
	entity := make([]entityType, 0)
	if err := engine.GetInstance().Select(&entity, query, values...); err != nil {
		return nil, err
	}
	return entity, nil
}

func (e *Entity[entityType]) FindOneBy(filters []F, p *Pagination) (entityType, error) {
	entity := *new(entityType)

	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return entity, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	whereElements := " WHERE "
	paramsCounter := 1
	values := make([]interface{}, 0)
	for i := 0; i < len(filters); i++ {
		filter := filters[i]
		where := fmt.Sprintf(" %s %s ", filter.FieldName, filter.Sign)

		whereElements += fmt.Sprintf("%s$%d %s ", where, paramsCounter, filter.Cond)
		values = append(values, filter.Value)
		paramsCounter++
	}

	query := "SELECT * FROM " + tableName + whereElements + e.getPagQuery(p)
	if err := engine.GetInstance().Get(&entity, query, values...); err != nil {
		return entity, err
	}
	return entity, nil
}

func (e *Entity[entityType]) FindAll(p *Pagination) ([]entityType, error) {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return nil, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	var entities []entityType
	if err := engine.GetInstance().Select(&entities, "SELECT * FROM "+tableName+e.getPagQuery(p)); err != nil {
		return nil, err
	}
	return e.withLazies(entities)
}

func (e *Entity[entityType]) Delete(id int64) error {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id=%d;", tableName, id)

	_, err := engine.GetInstance().Exec(query)
	if err != nil {
		return errors.Wrap(err, "gpa can't remove row with error")
	}
	return nil
}

func (e *Entity[entityType]) Update(entity entityType) error {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	t := reflect.TypeOf(entity)
	if kind := t.Kind(); kind != reflect.Struct {
		log.Panicf("should be struct type, %v instead.", kind)
	}

	fields := getReflectedData(entity, false)
	values := make([]string, 0)
	for _, f := range fields {
		values = append(values, fmt.Sprintf("%s = :%s", f.FieldDb, f.FieldDb))
	}

	queryStr := fmt.Sprintf("UPDATE %s SET %v WHERE id = :id", tableName, strings.Join(values, ","))

	stmt, err := engine.GetInstance().PrepareNamed(queryStr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(entity)
	return err
}

func (e *Entity[entityType]) Insert(item interface{}) error {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	t := reflect.TypeOf(item)
	if kind := t.Kind(); kind != reflect.Struct {
		log.Panicf("should be struct type, %v instead.", kind)
	}

	mdl := getReflectedData(item, false)
	queryStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(mdl.GetFieldsDb(), ","), ":"+strings.Join(mdl.GetFieldsDb(), ", :"))

	stmt, err := engine.GetInstance().PrepareNamed(queryStr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(item)
	return err
}

func (e *Entity[entityType]) Inserts(items []entityType) error {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return errors.New(fmt.Sprintf("should be struct type, %v instead.", reflect.TypeOf(e.entityObj)))
	}

	item := items[0]
	mdl := getReflectedData(item, false)

	rows := make([]interface{}, 0)
	queryArgs := make([]string, 0)
	queryStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", tableName, strings.Join(mdl.GetFieldsDb(), ","))

	for _, i := range items {
		queryArgs = append(queryArgs, "(?)")
		tmp := make([]interface{}, 0)
		for _, name := range mdl.GetFieldsName() {
			f := reflect.Indirect(reflect.ValueOf(i)).FieldByName(name)
			tmp = append(tmp, f.Interface())
		}
		rows = append(rows, tmp)
	}
	query, args, err := sqlx.In(queryStr+strings.Join(queryArgs, ", "), rows...)
	if err != nil {
		return err
	}
	_, err = engine.GetInstance().Exec(engine.GetInstance().Rebind(query), args...)
	return err
}
