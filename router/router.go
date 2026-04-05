package router

import (
	"github.com/unidb/unidb-go/planner"
)

// Route maps a step to a specific database
type Route struct {
	StepID   int
	Database string
}

type Router interface {
	Route(plan *planner.ExecutionPlan) ([]Route, error)
}

type DefaultRouter struct{}

func (r *DefaultRouter) Route(plan *planner.ExecutionPlan) ([]Route, error) {
	var routes []Route

	for _, step := range plan.Steps {
		if step.Type == "SCAN" {
			routes = append(routes, Route{
				StepID:   step.ID,
				Database: step.Database,
			})
		}
	}

	return routes, nil
}
