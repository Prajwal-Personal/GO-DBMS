package parser

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

// Parse converts a raw SQL string into an internal QueryAST
func Parse(query string) (*QueryAST, error) {
	// Check cache first
	if ast, ok := getFromCache(query); ok {
		return ast, nil
	}

	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	var ast *QueryAST

	switch node := stmt.(type) {
	case *sqlparser.Select:
		ast, err = parseSelect(node)
	case *sqlparser.Insert:
		ast, err = parseInsert(node)
	case *sqlparser.Update:
		ast, err = parseUpdate(node)
	case *sqlparser.Delete:
		ast, err = parseDelete(node)
	case *sqlparser.DDL:
		ast, err = parseDDL(node)
	case *sqlparser.DBDDL:
		ast, err = parseDBDDL(node)
	case *sqlparser.Show:
		ast, err = parseShow(node)
	case *sqlparser.Use:
		ast, err = parseUse(node)
	case *sqlparser.Set, *sqlparser.Begin, *sqlparser.Commit, *sqlparser.Rollback:
		ast = &QueryAST{Type: "META"}
	default:
		return nil, ErrUnsupportedQuery
	}

	if err != nil {
		return nil, err
	}

	// Validate
	if len(ast.Tables) == 0 && ast.Type != "DDL" && ast.Type != "META" {
		return nil, ErrInvalidQuery
	}

	ast.RawQuery = query

	// Save to cache
	saveToCache(query, ast)

	return ast, nil
}
