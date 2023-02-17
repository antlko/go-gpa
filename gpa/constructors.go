package gpa

type Pagination struct {
	Limit  int64
	Offset int64
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
