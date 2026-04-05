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

type SecurityEngine struct {
	aiThreshold float64
}

// NewSecurityEngine creates a new SecurityEngine with the given AI threshold.
func NewSecurityEngine(aiThreshold float64) *SecurityEngine {
	return &SecurityEngine{
		aiThreshold: aiThreshold,
	}
}

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

func detectAIInjection(query string) (float64, bool) {
	q := strings.ToUpper(query)
	score := 0.0

	// Simulated AI feature extraction and scoring
	// Feature 1: Suspicious keywords often used in advanced attacks
	suspiciousKeywords := []string{"SLEEP(", "WAITFOR DELAY", "BENCHMARK(", "PG_SLEEP(", "DBMS_PIPE", "XP_CMDSHELL", "INFORMATION_SCHEMA", "CONCAT(", "CHAR(", "HEX("}
	keywordMatchCount := 0
	for _, kw := range suspiciousKeywords {
		if strings.Contains(q, kw) {
			keywordMatchCount++
		}
	}
	
	if keywordMatchCount > 0 {
		score += float64(keywordMatchCount) * 0.35 // higher weight for suspicious advanced keywords
	}

	// Feature 2: High number of special characters often used in tautologies and bypassing
	specialChars := "';-/*="
	specialCharCount := 0
	for _, char := range query {
		if strings.ContainsRune(specialChars, char) {
			specialCharCount++
		}
	}
	
	if specialCharCount > 5 {
		score += 0.25
	}
	
	// Feature 3: Query length anomalies (extremely long queries could be an attempt to bypass buffers)
	if len(query) > 250 {
		score += 0.15
	}

	return score, score > 0.8
}

func (e *SecurityEngine) Analyze(query string) Decision {
	// Rule-based generic SQL Injection detection
	if detectInjection(query) {
		return Decision{Block: true, Reason: "SQL Injection (Rule-based)"}
	}
	
	// AI-based SQL Injection detection
	if score, isBlocked := detectAIInjection(query); isBlocked {
		return Decision{Block: true, Reason: "AI-based SQL Injection Detected (High Risk Score)"}
	} else if score > 0.5 {
		return Decision{Flag: true, Reason: "AI-based Analysis Flagged Suspicious Activity (Medium Risk Score)"}
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
