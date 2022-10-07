package crate

import "context"

type MutationType uint8

const (
	Inserting MutationType = 1
	Updating  MutationType = 2
)

type BeforeMutation interface {
	BeforeMutation(ctx context.Context, mutationType MutationType) error
}

type AfterMutation interface {
	AfterMutation(ctx context.Context, mutationType MutationType)
}

type IsZeroer interface {
	IsZero() bool
}
