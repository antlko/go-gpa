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

type Sign string

const (
	Equal     Sign = "="
	MoreEqual Sign = ">="
	LessEqual Sign = "<="
	More      Sign = ">"
	Less      Sign = "<"
)

type Condition string

const (
	AND Condition = "AND"
	OR  Condition = "OR"
)

// F Filter
type F struct {
	FieldName string
	Sign      Sign
	Value     interface{}
	Cond      Condition
}

func (e Entity[entityType]) WithOne(field any) {

}

func (e Entity[entityType]) Get(where string, args ...interface{}) (*entityType, error) {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return nil, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	if where != "" {
		where = " WHERE " + where
	}
	entity := e.entityObj.(entityType)
	if err := engine.DB.Get(&entity, "SELECT * FROM "+tableName+where, args...); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (e Entity[entityType]) Select(where string, args ...interface{}) ([]entityType, error) {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return nil, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	entity := make([]entityType, 0)
	if where != "" {
		where = " WHERE " + where
	}
	if err := engine.DB.Select(&entity, "SELECT * FROM "+tableName+where, args...); err != nil {
		return nil, err
	}
	return entity, nil
}

func (e Entity[entityType]) FindByID(id int64) (*entityType, error) {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return nil, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	entity := e.entityObj.(entityType)
	if err := engine.DB.Get(&entity, "SELECT * FROM "+tableName+" WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (e Entity[entityType]) FindBy(filters []F) ([]entityType, error) {
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

	query := "SELECT * FROM " + tableName + whereElements
	entity := make([]entityType, 0)
	if err := engine.DB.Select(&entity, query, values...); err != nil {
		return nil, err
	}
	return entity, nil
}

func (e Entity[entityType]) FindAll() ([]entityType, error) {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return nil, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	var entities []entityType
	if err := engine.DB.Select(&entities, "SELECT * FROM "+tableName); err != nil {
		return nil, err
	}
	return entities, nil
}

func (e Entity[entityType]) Delete(id int64) error {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id=%d;", tableName, id)

	_, err := engine.DB.Exec(query)
	if err != nil {
		return errors.Wrap(err, "gpa can't remove row with error")
	}
	return nil
}

func (e Entity[entityType]) Insert(item entityType) (int64, error) {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return 0, errors.New(fmt.Sprintf("entity %s wasn't configurate ", reflect.TypeOf(e.entityObj)))
	}

	t := reflect.TypeOf(item)
	if kind := t.Kind(); kind != reflect.Struct {
		log.Panicf("should be struct type, %v instead.", kind)
	}

	fields, _, _ := getReflectedData(item)

	queryStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", tableName, strings.Join(fields, ","), ":"+strings.Join(fields, ", :"))

	stmt, err := engine.DB.PrepareNamed(queryStr)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int64
	err = stmt.QueryRow(item).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (e Entity[entityType]) Inserts(items []entityType) error {
	tableName, ok := engine.GetTableName(e.entityObj)
	if !ok {
		return errors.New(fmt.Sprintf("should be struct type, %v instead.", reflect.TypeOf(e.entityObj)))
	}

	item := items[0]
	fieldsDb, fieldsNames, _ := getReflectedData(item)

	rows := make([]interface{}, 0)
	queryArgs := make([]string, 0)
	queryStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", tableName, strings.Join(fieldsDb, ","))

	for _, i := range items {
		queryArgs = append(queryArgs, "(?)")
		tmp := make([]interface{}, 0)
		for _, name := range fieldsNames {
			v := reflect.ValueOf(i)
			f := reflect.Indirect(v).FieldByName(name)
			tmp = append(tmp, f.Interface())
		}
		rows = append(rows, tmp)
	}
	query, args, err := sqlx.In(queryStr+strings.Join(queryArgs, ", "), rows...)
	if err != nil {
		return err
	}
	_, err = engine.DB.Exec(engine.DB.Rebind(query), args...)
	return err
}

func getReflectedData(item interface{}) ([]string, []string, []string) {
	t := reflect.TypeOf(item)
	if kind := t.Kind(); kind != reflect.Struct {
		log.Fatalf("This program expects to work on a struct; we got a %v instead.", kind)
	}
	fieldsDb := make([]string, 0)
	fieldsNames := make([]string, 0)
	fieldsTypes := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("db")
		_, ok := f.Tag.Lookup("generated")
		if len(tag) > 0 && tag != "id" && !ok {
			fieldsDb = append(fieldsDb, tag)
			fieldsNames = append(fieldsNames, f.Name)
			fieldsTypes = append(fieldsTypes, f.Type.String())
			continue
		}

		// get data from struct nested fields
		if f.Type.Kind() == reflect.Struct && len(f.Tag.Get("db")) == 0 {
			fieldValue := reflect.ValueOf(item).Field(i).Interface()
			dbFields, nameFields, _ := getReflectedData(fieldValue)
			fieldsDb = append(fieldsDb, dbFields...)
			fieldsNames = append(fieldsNames, nameFields...)
			fieldsTypes = append(fieldsNames, f.Type.String())
		}
	}

	return fieldsDb, fieldsNames, fieldsTypes
}