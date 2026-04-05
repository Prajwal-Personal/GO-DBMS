package planner

import "github.com/unidb/unidb-go/parser"

// ExecutionPlan represents physical execution steps
type ExecutionPlan struct {
	Steps []ExecutionStep
}

// ExecutionStep is an individual query or operation to perform
type ExecutionStep struct {
	ID        int
	Type      string // "SCAN", "JOIN", "FILTER"
	Database  string
	Query     string
	DependsOn []int
}

// QueryMetadata is the logical info extracted by analyzer
type QueryMetadata struct {
	Tables    []parser.TableNode
	IsMultiDB bool
	HasJoin   bool
	Filters   []parser.ConditionNode
}
