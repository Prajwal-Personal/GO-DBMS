package router

import (
	"os"
	"testing"
	"github.com/unidb/unidb-go/planner"
)

func TestAIRouter(t *testing.T) {
	tmpFile := "test_db.json"
	defer os.Remove(tmpFile)

	err := InitAIRouter(tmpFile)
	if err != nil {
		t.Fatalf("Failed to init AIRouter: %v", err)
	}

	ar, ok := ActiveRouter.(*AIRouter)
	if !ok {
		t.Fatal("ActiveRouter is not of type *AIRouter")
	}

	plan := &planner.ExecutionPlan{
		Steps: []planner.ExecutionStep{
			{ID: 1, Type: "SCAN", Query: "SELECT * FROM users WHERE ROWNUM <= 10"},
			{ID: 2, Type: "SCAN", Query: "SELECT JSON_EXTRACT(data, '$.name') FROM configs"},
			{ID: 3, Type: "SCAN", Query: "{ \"find\": \"users\", \"filter\": { \"age\": { \"$gt\": 25 } } }"},
		},
	}

	routes, err := ar.Route(plan)
	if err != nil {
		t.Fatalf("Failed to route: %v", err)
	}

	if len(routes) != 3 {
		t.Fatalf("Expected 3 routes, got %d", len(routes))
	}

	// Verify the AI essentially picked up the seed data
	if routes[0].Database != "oracle" {
		t.Errorf("Expected oracle for ROWNUM query, got %s", routes[0].Database)
	}
	if routes[1].Database != "mysql" {
		t.Errorf("Expected mysql for JSON_EXTRACT query, got %s", routes[1].Database)
	}
	if routes[2].Database != "mongodb" {
		t.Errorf("Expected mongodb for JSON format query, got %s", routes[2].Database)
	}
}
