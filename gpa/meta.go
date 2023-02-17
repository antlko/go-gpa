package gpa

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"reflect"
	"strconv"
)

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

// EntityMetadataInfo MetaData info structure for saving information about field
// Contains also nested field FieldEntity, which should be the same  EntityMetadataInfo type
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

func (e *Entity[entityType]) withLazies(entities []entityType) ([]entityType, error) {
	lzs := make([]entityType, 0)
	for i := 0; i < len(entities); i++ {
		lz, err := e.withsLazy(entities[i])
		if err != nil {
			return nil, errors.Wrap(err, "error fetching lazy entity")
		}
		lzs = append(lzs, lz)
	}
	return lzs, nil
}

func (e *Entity[entityType]) withsLazy(entity entityType) (entityType, error) {
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
		currentTable, ok := engine.GetTableName(e.entityObj)
		if !ok {
			return entity, errors.New("current table can't be found or wasn't initialized before")
		}

		lmtd := getReflectedData(generalLazy, true)
		joinedTableId := lmtd.GetMappedByMetaJoin(lazyTable)
		joinedTableCurrentId := lmtd.GetMappedByMetaJoin(currentTable)

		var joinValue interface{}
		entityTypeOf := reflect.TypeOf(entity)
		for i := 0; i < entityTypeOf.NumField(); i++ {
			f := entityTypeOf.Field(i)
			if f.Tag.Get("db") == joinedTableCurrentId {
				joinValue = reflect.ValueOf(entity).Field(i).Interface()
				break
			}
		}

		inheretedWhere := " " + lazyEntityMeta.MappedBy + " "
		if lazyEntityMeta.Join != lazyTable {
			whereId := " WHERE " + lazyEntityMeta.MappedBy + fmt.Sprintf(" = $1")
			inheretedWhere = " SELECT " + lazyEntityMeta.FetchBy + " FROM " + lazyEntityMeta.Join + whereId
		}

		query := fmt.Sprintf("SELECT * FROM "+lazyTable+" WHERE %s IN (%s)", joinedTableId, inheretedWhere)

		ptr := reflect.New(reflect.SliceOf(reflect.TypeOf(lazyEntity)))
		iface := ptr.Interface()
		if err := engine.GetInstance().Select(iface, query, joinValue); err != nil {
			return entity, err
		}

		val := reflect.ValueOf(iface)
		reflect.Indirect(reflect.ValueOf(&entity)).Field(lazyEntityMeta.Idx).Set(val)
	}
	return entity, nil
}
