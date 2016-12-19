package main

// Backend is a receiver of metrics
type Backend interface {
	Collect(r *result) error
}
