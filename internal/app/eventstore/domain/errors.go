package domain

// error codes
const (
	ErrReadStreamFailed = iota + 9101
	ErrConcurrentChange
	ErrMisconfiguration
	ErrListen
)
