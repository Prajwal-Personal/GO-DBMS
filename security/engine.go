package security

import (
	"context"
	"strings"
	"time"

	"github.com/unidb/unidb-go/api"
	"github.com/unidb/unidb-go/internal"
	"github.com/unidb/unidb-go/parser"
)

type Decision struct {
	Block  bool
	Flag   bool
	Reason string
}

type SecurityEngine struct{}

var injectionPatterns = []string{
	" OR 1=1",
	"--",
	"/*",
	"UNION SELECT",
}

func detectInjection(query string) bool {
	q := strings.ToUpper(query)
	for _, pattern := range injectionPatterns {
		if strings.Contains(q, strings.ToUpper(pattern)) {
			return true
		}
	}
	return false
}

var sensitiveTables = []string{
	"users",
	"passwords",
	"tokens",
}

func detectExfiltration(ast *parser.QueryAST) bool {
	if ast == nil {
		return false
	}
	for _, table := range ast.Tables {
		for _, sens := range sensitiveTables {
			if table.Name == sens && ast.Limit == nil {
				return true
			}
		}
	}
	return false
}

func detectAnomaly(query string) bool {
	// naive anomaly detection mock
	if len(query) > 5000 {
		return true
	}
	return false
}

func (e *SecurityEngine) Analyze(query string) Decision {
	if detectInjection(query) {
		return Decision{Block: true, Reason: "SQL Injection"}
	}

	ast, err := parser.Parse(query)
	if err == nil {
		if detectExfiltration(ast) {
			return Decision{Flag: true, Reason: "Data Exfiltration Risk: Full scan on sensitive table without LIMIT."}
		}
	}

	if detectAnomaly(query) {
		return Decision{Flag: true, Reason: "Anomaly"}
	}

	return Decision{}
}

// SecurityMiddleware creates a unified API middleware
func SecurityMiddleware(engine *SecurityEngine) api.Middleware {
	return func(next api.Handler) api.Handler {
		return func(ctx context.Context, query string, args ...any) (internal.Result, error) {
			decision := engine.Analyze(query)

			// logging mock
			_ = SecurityLog{
				Query:     query,
				Timestamp: time.Now(),
				Decision:  "Proceed", // default
				Reason:    decision.Reason,
			}

			if decision.Block {
				return nil, api.ErrBlockedQuery
			}
			if decision.Flag {
				// We just flag/log it but proceed
				_ = "Flagged: " + decision.Reason
			}

			return next(ctx, query, args...)
		}
	}
}

type SecurityLog struct {
	Query     string
	Timestamp time.Time
	Decision  string
	Reason    string
}
