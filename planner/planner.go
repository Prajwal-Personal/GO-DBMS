package planner

import "github.com/unidb/unidb-go/parser"

// GeneratePlan creates a physical execution plan from an AST
func GeneratePlan(ast *parser.QueryAST) (*ExecutionPlan, error) {
	meta := Analyze(ast)
	
	if meta.IsMultiDB {
		return planFederatedQuery(ast, meta)
	}
	
	// Single DB plan
	db := "default"
	queryStr := "SELECT * FROM ..."
	if len(ast.Tables) > 0 {
		db = ast.Tables[0].Database
		if db == "mongodb" {
			// Mock NoSQL translation (Phase 2 & 4 mapping showcase)
			queryStr = `{ "find": "` + ast.Tables[0].Name + `", "filter": { "$gt": 25 } }`
		}
	}

	plan := &ExecutionPlan{
		Steps: []ExecutionStep{
			{
				ID:       1,
				Type:     "SCAN",
				Database: db,
				Query:    queryStr,
			},
		},
	}
	return plan, nil
}
