package parser

import (
	"github.com/xwb1989/sqlparser"
)

func parseDDL(ddl *sqlparser.DDL) (*QueryAST, error) {
	// For basic passing through DDL support
	ast := &QueryAST{
		Type: "DDL",
	}
	
	if ddl.Table.Name.String() != "" {
		ast.Tables = []TableNode{
			{Name: ddl.Table.Name.String()},
		}
	} else if ddl.Action == "create" {
		// Possibly CREATE DATABASE, just pass the syntax
		// The AST itself doesn't need to parse down everything for MVP routing
	}
	
	return ast, nil
}

func parseDBDDL(ddldb *sqlparser.DBDDL) (*QueryAST, error) {
	return &QueryAST{
		Type: "DDL",
	}, nil
}

func parseShow(show *sqlparser.Show) (*QueryAST, error) {
	return &QueryAST{
		Type: "META",
	}, nil
}

func parseUse(use *sqlparser.Use) (*QueryAST, error) {
	return &QueryAST{
		Type: "META",
	}, nil
}
