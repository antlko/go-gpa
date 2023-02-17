package gpa

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Sign string

type Pagination struct {
	Limit  int64
	Offset int64
}

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

	lazyEntities := getLazyEntitiesMetaData(entity)
	for i := 0; i < len(lazyEntities); i++ {
		lazyEntityMeta := lazyEntities[i]
		lazySingleEntityType := lazyEntityMeta.Type

		generalLazy, ok := engine.GetEntity(lazyEntityMeta.Join)
		if !ok {
			return entity, errors.New("lazy type [" + lazyEntityMeta.Join + "] can't be found or wasn't initialized before")
		}

		lazyEntity := generalLazy
		if lazyEntityMeta.Type.Elem().Kind() == reflect.Slice {
			lazySingleEntityType = lazyEntityMeta.Type.Elem().Elem()
			lazyEntity = reflect.New(lazySingleEntityType).Elem().Interface()
		}

		lazyTable, ok := engine.GetTableName(lazyEntity)
		if !ok {
			return entity, errors.New("lazy type can't be found or wasn't initialized before")
		}
		lmtd := getReflectedData(generalLazy, true)
		joinedTableId := lmtd.GetMappedByMetaJoin(lazyTable)

		inheretedWhere := " " + lazyEntityMeta.MappedBy + " "
		if lazyEntityMeta.Join != lazyTable {
			inheretedWhere = " SELECT " + lazyEntityMeta.FetchBy + " FROM " + lazyEntityMeta.Join + " WHERE " + lazyEntityMeta.MappedBy + fmt.Sprintf(" = $1")
		}

		query := fmt.Sprintf("SELECT * FROM "+lazyTable+" WHERE %s IN (%s)", joinedTableId, inheretedWhere)

		ptr := reflect.New(reflect.SliceOf(reflect.TypeOf(lazyEntity)))
		iface := ptr.Interface()
		if err := engine.GetInstance().Select(iface, query, id); err != nil {
			return entity, err
		}

		val := reflect.ValueOf(iface)
		reflect.Indirect(reflect.ValueOf(&entity)).Field(lazyEntityMeta.Idx).Set(val)
	}
	return entity, nil
}

type MetaLazyEntity struct {
	Idx      int
	Type     reflect.Type
	Join     string
	MappedBy string
	FetchBy  string
}

func getLazyEntitiesMetaData(item interface{}) []MetaLazyEntity {
	t := reflect.TypeOf(item)
	if kind := t.Kind(); kind != reflect.Struct {
		log.Fatalf("should be structure, got %v instead.", kind)
	}

	var lazy []MetaLazyEntity
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("fetch") == "lazy" {
			lazy = append(lazy, MetaLazyEntity{
				Idx:      i,
				Type:     f.Type,
				Join:     f.Tag.Get("join"),
				MappedBy: f.Tag.Get("mappedBy"),
				FetchBy:  f.Tag.Get("fetchBy"),
			})
		}
	}
	return lazy
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

func (e *Entity[entityType]) getPagQuery(p *Pagination) string {
	var pagQuery = ""
	if p != nil && p.Limit != 0 {
		pagQuery += " LIMIT " + strconv.FormatInt(p.Limit, 10)
	}
	if p != nil && p.Offset != 0 {
		pagQuery += " OFFSET " + strconv.FormatInt(p.Offset, 10)
	}
	return pagQuery
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
	return entities, nil
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

	fields := getReflectedData(entity, false).GetFieldsName()
	values := make([]string, 0)
	for _, f := range fields {
		values = append(values, fmt.Sprintf("%s = :%s", f, f))
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

type EntityMetadataInfo struct {
	FieldDb    string
	FieldName  string
	FieldType  reflect.Type
	FieldValue interface{}

	FieldEntity interface{}

	MetaTags MetaTags
}

type MetaTags struct {
	Join     string
	MappedBy string
	Fetch    string
}

type MetaDataList []EntityMetadataInfo

func (m MetaDataList) GetDataByDBTag(dbTag string) EntityMetadataInfo {
	for _, v := range m {
		if v.FieldDb == dbTag {
			return v
		}
	}
	return EntityMetadataInfo{}
}

func (m MetaDataList) GetMappedByMetaJoin(tableNameMeta string) string {
	for _, v := range m {
		if v.MetaTags.Join == tableNameMeta {
			return v.MetaTags.MappedBy
		}
	}
	return ""
}

func (m MetaDataList) GetFieldsDb() []string {
	var arr []string
	for _, v := range m {
		arr = append(arr, v.FieldDb)
	}
	return arr
}

func (m MetaDataList) GetFieldsName() []string {
	var arr []string
	for _, v := range m {
		arr = append(arr, v.FieldName)
	}
	return arr
}

func (m MetaDataList) GetFieldsType() []reflect.Type {
	var arr []reflect.Type
	for _, v := range m {
		arr = append(arr, v.FieldType)
	}
	return arr
}

func (m MetaDataList) GetFieldValues() interface{} {
	var arr []interface{}
	for _, v := range m {
		arr = append(arr, v.FieldValue)
	}
	return arr
}

func (m MetaDataList) GetFieldEntity() []interface{} {
	var arr []interface{}
	for _, v := range m {
		arr = append(arr, v.FieldEntity)
	}
	return arr
}

func getOperationReflectedData(item interface{}, withId bool) MetaDataList {
	t := reflect.TypeOf(item)
	if kind := t.Kind(); kind != reflect.Struct {
		log.Fatalf("should be structure, got %v instead.", kind)
	}

	metaDataFields := make(MetaDataList, 0)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("db")

		metaData := EntityMetadataInfo{}
		join := f.Tag.Get("join")
		mappedBy := f.Tag.Get("mappedBy")
		fetch := f.Tag.Get("fetch")

		if len(tag) > 0 && tag != "id" || (withId && tag == "id") {
			metaData.FieldDb = tag
			metaData.FieldName = f.Name
			metaData.FieldType = f.Type
			metaData.FieldValue = reflect.ValueOf(item).Field(i).Interface()

			metaData.MetaTags = MetaTags{
				Join:     join,
				MappedBy: mappedBy,
				Fetch:    fetch,
			}

			if f.Type.Kind() == reflect.Struct {
				reflMetaData := getOperationReflectedData(metaData.FieldValue, true)
				metaData.FieldEntity = reflMetaData
				metaData.FieldValue = reflMetaData.GetDataByDBTag(metaData.MetaTags.MappedBy).FieldValue
			}
			metaDataFields = append(metaDataFields, metaData)
		}
	}

	return metaDataFields
}

func getReflectedData(item interface{}, withId bool) MetaDataList {
	t := reflect.TypeOf(item)
	if kind := t.Kind(); kind != reflect.Struct {
		log.Fatalf("should be structure, got %v instead.", kind)
	}

	metaDataFields := make(MetaDataList, 0)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("db")

		metaData := EntityMetadataInfo{}
		join := f.Tag.Get("join")
		mappedBy := f.Tag.Get("mappedBy")
		fetch := f.Tag.Get("fetch")

		if len(tag) > 0 && tag != "id" || (withId && tag == "id") {
			metaData.FieldDb = tag
			metaData.FieldName = f.Name
			metaData.FieldType = f.Type
			metaData.MetaTags = MetaTags{
				Join:     join,
				MappedBy: mappedBy,
				Fetch:    fetch,
			}

			if f.Type.Kind() == reflect.Struct {
				fieldValue := reflect.ValueOf(item).Field(i).Interface()
				reflMetaData := getReflectedData(fieldValue, withId)
				metaData.FieldEntity = reflMetaData
			}
			metaDataFields = append(metaDataFields, metaData)
		}
	}

	return metaDataFields
}
