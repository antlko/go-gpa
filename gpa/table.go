package gpa

import (
	"fmt"
	"github.com/gertd/go-pluralize"
	"reflect"
	"strings"
)

func initTable(obj any, entity interface{}) {
	gpaEntity, ok := entity.(GPAEntity)
	if ok {
		gpaEntity.GPAConfigure(engine)
	}

	tableName, ok := engine.GetTableName(gpaEntity)
	if !ok {
		structName := strings.ToLower(reflect.TypeOf(obj).Name())
		tableName = pluralize.NewClient().Plural(structName)

		if !isTableExists(tableName) {
			createTable(entity, tableName)
		}
	}
	if !isTableExists(tableName) {
		createTable(entity, tableName)
	}
	engine.entityTableNameMap[obj] = tableName
	engine.tableNameEntityMap[tableName] = obj
}

func isTableExists(name string) bool {
	rows, tableCheck := engine.GetInstance().Query("SELECT * FROM " + name + ";")

	if tableCheck != nil {
		return false
	}
	rows.Close()
	return true
}

func createTable(entity interface{}, tableName string) {
	emd := getReflectedData(entity, true)

	fieldsData := ""
	for i := 0; i < len(emd); i++ {
		pgType := getPGType(emd[i])
		fieldsData += fmt.Sprintf("%s %s, ", emd[i].FieldDb, pgType)
	}
	fieldsData = fieldsData[:len(fieldsData)-2]

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, fieldsData)
	rows, err := engine.GetInstance().Queryx(query)
	if err != nil {
		panic(fmt.Sprintf("gpa can't create the table with error:%s", err))
	}
	rows.Close()
}

func getPGType(emd EntityMetadataInfo) string {
	tp := strings.ToLower(emd.FieldType.String())
	null := ""
	if !strings.Contains(tp, "null") && emd.FieldDb != "id" {
		null = " NOT NULL "
	}

	outType := ""
	if emd.FieldDb == "id" {
		outType = " SERIAL "
	} else if strings.Contains(tp, "time") {
		outType = " DATE "
	} else if strings.Contains(tp, "int") {
		outType = " INTEGER "
	} else if strings.Contains(tp, "string") {
		outType = " TEXT "
	} else if strings.Contains(tp, "bool") {
		outType = " BOOL "
	} else {
		// checking custom type
		if emd.FieldEntity != nil {
			nestedMD := emd.FieldEntity.(MetaDataList)
			mappedField := nestedMD.GetDataByDBTag(emd.MetaTags.MappedBy)
			tp = getPGType(mappedField)
			if strings.Contains(tp, "SERIAL") {
				outType = " INTEGER "
			}
		}
	}

	return outType + null
}
