package crate

type MutationType uint8

const (
	Inserting MutationType = 1
	Updating  MutationType = 2
)

type BeforeMutation interface {
	BeforeMutation(MutationType) error
}

type AfterMutation interface {
	AfterMutation(MutationType)
}

type IsZeroer interface {
	IsZero() bool
}
