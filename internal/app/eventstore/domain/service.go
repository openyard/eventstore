package domain

type Service interface {
	HandleFunc(cmd Command) error
	QueryFunc(cmd Command) ([]Stream, error)
}

type HandleFunc func(cmd Command) error
type QueryFunc func(cmd Command) ([]Stream, error)
