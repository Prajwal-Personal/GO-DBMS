package planner

import "github.com/unidb/unidb-go/parser"

func planFederatedQuery(ast *parser.QueryAST, meta QueryMetadata) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{}
	stepID := 1
	var dependsOn []int

	for _, tbl := range ast.Tables {
		db := tbl.Database
		if db == "" {
			db = "default"
		}
		step := ExecutionStep{
			ID:       stepID,
			Type:     "SCAN",
			Database: db,
			Query:    "SELECT * FROM " + tbl.Name, // minimal reconstruction
		}
		plan.Steps = append(plan.Steps, step)
		dependsOn = append(dependsOn, stepID)
		stepID++
	}

	if meta.HasJoin {
		plan.Steps = append(plan.Steps, ExecutionStep{
			ID:        stepID,
			Type:      "JOIN",
			DependsOn: dependsOn,
		})
	}

	return plan, nil
}
