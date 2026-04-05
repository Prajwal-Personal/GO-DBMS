package circuit

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open due to repeated failures")

// State represents the status of the CircuitBreaker.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker acts as a protective shield against failing backend dependencies.
type CircuitBreaker struct {
	mu           sync.RWMutex
	state        State
	failures     int
	threshold    int
	timeout      time.Duration
	lastFailure  time.Time
}

// NewCircuitBreaker creates a new circuit breaker mechanism.
func NewCircuitBreaker(failureThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     StateClosed,
		threshold: failureThreshold,
		timeout:   timeout,
	}
}

// Execute runs the given operation through the circuit breaker.
func (cb *CircuitBreaker) Execute(operation func() error) error {
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.timeout {
			// Testing recovery 
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mu.Unlock()

	err := operation()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.threshold {
			cb.state = StateOpen
		}
		return err
	}

	// Success resolves the state
	cb.failures = 0
	cb.state = StateClosed
	return nil
}
