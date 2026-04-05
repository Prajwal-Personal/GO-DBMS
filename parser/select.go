package parser

import (
	"github.com/xwb1989/sqlparser"
)

func parseSelect(sel *sqlparser.Select) (*QueryAST, error) {
	ast := &QueryAST{
		Type: "SELECT",
	}

	// Extract fields
	for _, expr := range sel.SelectExprs {
		switch e := expr.(type) {
		case *sqlparser.AliasedExpr:
			switch subE := e.Expr.(type) {
			case *sqlparser.ColName:
				ast.Fields = append(ast.Fields, FieldNode{
					Name:  subE.Name.String(),
					Table: subE.Qualifier.Name.String(),
					Alias: e.As.String(),
				})
			case *sqlparser.FuncExpr:
				ast.Fields = append(ast.Fields, FieldNode{
					Name:      sqlparser.String(subE),
					Aggregate: subE.Name.String(),
					Alias:     e.As.String(),
				})
			}
		case *sqlparser.StarExpr:
			ast.Fields = append(ast.Fields, FieldNode{
				Name:  "*",
				Table: e.TableName.Name.String(),
			})
		}
	}

	// Extract tables
	for _, from := range sel.From {
		extractTables(from, ast)
	}

	// Conditions (minimal implementation for MVP)
	if sel.Where != nil {
		switch expr := sel.Where.Expr.(type) {
		case *sqlparser.ComparisonExpr:
			ast.Conditions = append(ast.Conditions, ConditionNode{
				Left:     sqlparser.String(expr.Left),
				Operator: expr.Operator,
				Right:    sqlparser.String(expr.Right),
			})
		}
	}

	// Limit
	if sel.Limit != nil {
		if val, err := sqlparser.String(sel.Limit.Rowcount), error(nil); err == nil {
			// convert val to int, mocked for MVP
			_ = val
			ast.Limit = new(int) // just marking non-nil
		}
	}

	return ast, nil
}

func extractTables(expr sqlparser.TableExpr, ast *QueryAST) {
	switch t := expr.(type) {
	case *sqlparser.AliasedTableExpr:
		switch te := t.Expr.(type) {
		case sqlparser.TableName:
			ast.Tables = append(ast.Tables, TableNode{
				Database: te.Qualifier.String(),
				Name:     te.Name.String(),
				Alias:    t.As.String(),
			})
		}
	case *sqlparser.JoinTableExpr:
		ast.Joins = append(ast.Joins, JoinNode{
			Type: t.Join,
		})
		// Recursively extract from Left and Right
		extractTables(t.LeftExpr, ast)
		extractTables(t.RightExpr, ast)
	}
}

