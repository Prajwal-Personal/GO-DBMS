package router

import (
	"math"
	"strings"
	"sync"
)

// RoutingPrediction contains the model's recommendation.
type RoutingPrediction struct {
	TargetDatabase string
	TargetDialect  string
	Confidence     float64
}

// RoutingModel is a simple Naive Bayes-like text classifier.
type RoutingModel struct {
	mu sync.RWMutex
	
	// Word occurrences per (Database + "|" + Dialect) class
	wordCounts  map[string]map[string]int 
	classCounts map[string]int
	vocab       map[string]bool
	totalDocs   int
}

// NewRoutingModel initializes an empty model.
func NewRoutingModel() *RoutingModel {
	return &RoutingModel{
		wordCounts:  make(map[string]map[string]int),
		classCounts: make(map[string]int),
		vocab:       make(map[string]bool),
		totalDocs:   0,
	}
}

// tokenize splits a query into normalized terms.
func tokenize(query string) []string {
	q := strings.ToUpper(query)
	// Replace some common punctuation with spaces to extract words
	reps := strings.NewReplacer(",", " ", ";", " ", "(", " ", ")", " ", "=", " ", ">", " ", "<", " ", "'", " ", `"`, " ")
	q = reps.Replace(q)
	
	words := strings.Fields(q)
	return words
}

// Train fits the classifier on a dataset of records.
func (m *RoutingModel) Train(records []TrainingRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Reset model to retrain
	m.wordCounts = make(map[string]map[string]int)
	m.classCounts = make(map[string]int)
	m.vocab = make(map[string]bool)
	m.totalDocs = 0
	
	for _, r := range records {
		m.totalDocs++
		class := r.Database + "|" + r.Dialect
		m.classCounts[class]++
		
		if m.wordCounts[class] == nil {
			m.wordCounts[class] = make(map[string]int)
		}
		
		words := tokenize(r.Query)
		for _, w := range words {
			m.wordCounts[class][w]++
			m.vocab[w] = true
		}
	}
}

// Predict calculates the highest probability class for a query.
func (m *RoutingModel) Predict(query string) RoutingPrediction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.totalDocs == 0 {
		return RoutingPrediction{TargetDatabase: "default", TargetDialect: "SQL", Confidence: 0.0}
	}
	
	words := tokenize(query)
	vocabSize := len(m.vocab)
	
	bestScore := math.Inf(-1)
	bestClass := ""
	
	for class, count := range m.classCounts {
		// P(Class)
		score := math.Log(float64(count) / float64(m.totalDocs))
		
		// P(Word | Class)
		totalWordsInClass := 0
		for _, c := range m.wordCounts[class] {
			totalWordsInClass += c
		}
		
		for _, w := range words {
			wordOccurrence := m.wordCounts[class][w]
			// Laplace smoothing
			prob := float64(wordOccurrence+1) / float64(totalWordsInClass+vocabSize)
			score += math.Log(prob)
		}
		
		if score > bestScore {
			bestScore = score
			bestClass = class
		}
	}
	
	parts := strings.SplitN(bestClass, "|", 2)
	db := "default"
	dialect := "SQL"
	if len(parts) == 2 {
		db = parts[0]
		dialect = parts[1]
	}

	return RoutingPrediction{
		TargetDatabase: db,
		TargetDialect:  dialect,
		Confidence:     bestScore, // Log-prob, useful only relative to others
	}
}
