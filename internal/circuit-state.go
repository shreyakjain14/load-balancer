package internal

type CircuitState int 

const (
	Closed CircuitState = iota 
	Open
	HalfOpen
)