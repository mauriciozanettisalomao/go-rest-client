package client

import "context"

// Requester defines the interface for a client that can make HTTP requests.
type Requester interface {
	Do(ctx context.Context, input interface{}, output interface{}) (int64, error)
}

// NewRequester returns a new Requester
func NewRequester(r Requester) Requester {
	return r
}
