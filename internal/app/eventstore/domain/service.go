package domain

type Service interface {
	HandleFunc(cmd Command) error
	QueryFunc(cmd Command) ([]Stream, error)
}
