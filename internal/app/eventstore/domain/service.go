package domain

type Service interface {
	HandleFunc(cmd Command) error
	QueryFunc(cmd Command) ([]Stream, error)
	SubscribeFunc(cmd Command) ([]Entry, error)
}

type HandleFunc func(cmd Command) error
type QueryFunc func(cmd Command) ([]Stream, error)
type SubscribeFunc func(cmd Command) ([]Entry, error)
