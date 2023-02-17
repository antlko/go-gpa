package gpa

type GPAEntity interface {
	GPAConfigure(e *Engine)
}

func From[entityType any]() *Entity[entityType] {
	entityObject := *new(entityType)
	_, ok := engine.entityTableNameMap[entityObject]
	if !ok {
		initTable(entityObject, entityObject)
	}

	return &Entity[entityType]{
		entityObj: entityObject,
	}
}
