package federation

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/unidb/unidb-go/internal"
	"github.com/unidb/unidb-go/planner"
)

// Row representation for in-memory merging
type Row map[string]any

type IntermediateResult struct {
	StepID int
	Rows   []Row
}

type FederationEngine struct {
	// dependencies on connections would go here
	// for MVP, we mock the execute execution
}

func (f *FederationEngine) Execute(ctx context.Context, plan *planner.ExecutionPlan) (internal.Result, error) {
	var wg sync.WaitGroup
	resChan := make(chan IntermediateResult, len(plan.Steps))
	errChan := make(chan error, len(plan.Steps))

	for _, step := range plan.Steps {
		if step.Type == "SCAN" {
			wg.Add(1)
			go func(s planner.ExecutionStep) {
				defer wg.Done()
				rows, err := f.executeStep(ctx, s)
				if err != nil {
					errChan <- err
					return
				}
				resChan <- IntermediateResult{StepID: s.ID, Rows: rows}
			}(step)
		}
	}

	wg.Wait()
	close(resChan)
	close(errChan)

	if len(errChan) > 0 {
		return nil, <-errChan
	}

	results := make(map[int][]Row)
	for r := range resChan {
		results[r.StepID] = r.Rows
	}

	for _, step := range plan.Steps {
		if step.Type == "JOIN" {
			// Extract right and left from depends on
			if len(step.DependsOn) < 2 {
				return nil, fmt.Errorf("join requires 2 dependencies")
			}
			lID, rID := step.DependsOn[0], step.DependsOn[1]
			leftRows, rightRows := results[lID], results[rID]
			
			// perform hash join
			merged := HashJoin(leftRows, rightRows, "id", "user_id") // Mocked keys
			results[step.ID] = merged
		}
	}

	// Figure out the final step's result and return a wrapped internal.Result
	finalRows := results[plan.Steps[len(plan.Steps)-1].ID]
	
	return &FederationResult{rows: finalRows, current: -1}, nil
}

func (f *FederationEngine) executeStep(ctx context.Context, step planner.ExecutionStep) ([]Row, error) {
	// Mock fetching from DB
	// We would normally `conn := db.connections[step.Database]; conn.Query(...)`
	return []Row{}, nil
}

// HashJoin implements an in-memory hash join (MVP)
func HashJoin(leftRows, rightRows []Row, leftKey, rightKey string) []Row {
	hash := make(map[any][]Row)
	for _, row := range rightRows {
		key := row[rightKey]
		hash[key] = append(hash[key], row)
	}

	var result []Row
	for _, l := range leftRows {
		key := l[leftKey]
		matches := hash[key]
		for _, r := range matches {
			merged := mergeRows(l, r)
			result = append(result, merged)
		}
	}
	return result
}

func mergeRows(a, b Row) Row {
	res := make(Row)
	for k, v := range a {
		if !strings.Contains(k, ".") {
			k = "a." + k
		}
		res[k] = v
	}
	for k, v := range b {
		if !strings.Contains(k, ".") {
			k = "b." + k
		}
		res[k] = v
	}
	return res
}

// FederationResult implements internal.Result
type FederationResult struct {
	rows    []Row
	current int
}

func (fr *FederationResult) Columns() []string {
	if len(fr.rows) == 0 {
		return []string{}
	}
	var cols []string
	for k := range fr.rows[0] {
		cols = append(cols, k)
	}
	return cols
}

func (fr *FederationResult) Next() bool {
	fr.current++
	return fr.current < len(fr.rows)
}

func (fr *FederationResult) Scan(dest ...any) error {
	if fr.current < 0 || fr.current >= len(fr.rows) {
		return fmt.Errorf("scan out of bounds")
	}
	// For MVP, simplistic scan mock
	return nil
}

func (fr *FederationResult) Close() error {
	return nil
}
