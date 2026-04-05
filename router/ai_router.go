package router

import (
	"log"
	"github.com/unidb/unidb-go/planner"
)

// AIRouter uses a trained Navie Bayes text AI model to predict
// the correct database and the dialect of queries.
type AIRouter struct {
	model *RoutingModel
	db    *CommonDB
}

// NewAIRouter creates an AIRouter and attempts to load and train the model.
func NewAIRouter(dbFilePath string) (*AIRouter, error) {
	commonDb, err := NewCommonDB(dbFilePath)
	if err != nil {
		return nil, err
	}
	
	model := NewRoutingModel()
	records := commonDb.GetAllRecords()
	if len(records) > 0 {
		model.Train(records)
		log.Printf("AIRouter initialized: trained on %d records\n", len(records))
	} else {
		// Insert some initial seed data if DB is empty
		seedData := []TrainingRecord{
			{Query: "SELECT * FROM users LIMIT 10", Database: "postgres", Dialect: "PostgreSQL"},
			{Query: "SELECT * FROM users WHERE ROWNUM <= 10", Database: "oracle", Dialect: "OracleSQL"},
			{Query: "SELECT * FROM users ORDER BY id DESC LIMIT 5", Database: "mysql", Dialect: "MySQL"},
			{Query: "SELECT JSON_EXTRACT(data, '$.name') FROM configs", Database: "mysql", Dialect: "MySQL"},
			{Query: "SELECT data->>'name' FROM configs", Database: "postgres", Dialect: "PostgreSQL"},
			{Query: "{ \"find\": \"users\", \"filter\": { \"age\": { \"$gt\": 25 } } }", Database: "mongodb", Dialect: "MQL"},
		}
		for _, s := range seedData {
			commonDb.AddRecord(s)
		}
		model.Train(commonDb.GetAllRecords())
		log.Printf("AIRouter initialized: seeded and trained on %d records\n", len(seedData))
	}

	return &AIRouter{
		model: model,
		db:    commonDb,
	}, nil
}

// Route predicts the best target database and dialect based on ML model
func (ar *AIRouter) Route(plan *planner.ExecutionPlan) ([]Route, error) {
	var routes []Route

	for _, step := range plan.Steps {
		if step.Type == "SCAN" || step.Type == "JOIN" || step.Type == "FILTER" || step.Type == "EXEC" {
			prediction := ar.model.Predict(step.Query)
			
			routes = append(routes, Route{
				StepID:         step.ID,
				Database:       prediction.TargetDatabase,
				CommandDialect: prediction.TargetDialect,
			})
			log.Printf("AIRouter predicted Step [%d] -> DB: %s, Dialect: %s (Confidence: %.2f)\n", 
				step.ID, prediction.TargetDatabase, prediction.TargetDialect, prediction.Confidence)
		}
	}

	return routes, nil
}
