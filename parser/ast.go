package parser

// QueryAST represents the normalized internal AST
type QueryAST struct {
	Type              string // SELECT, INSERT, UPDATE, DELETE
	RawQuery          string
	Tables            []TableNode
	Fields            []FieldNode
	Conditions        []ConditionNode
	Joins             []JoinNode
	Limit             *int
	IsNoSQLCompatible bool
}

type TableNode struct {
	Name     string
	Database string // postgres, mysql, etc.
	Alias    string
}

type FieldNode struct {
	Name      string
	Table     string
	Alias     string
	Aggregate string // COUNT, SUM, etc.
}

type ConditionNode struct {
	Left     string
	Operator string
	Right    any
}

type JoinNode struct {
	Type       string // INNER, LEFT
	LeftTable  string
	RightTable string
	On         ConditionNode
}
