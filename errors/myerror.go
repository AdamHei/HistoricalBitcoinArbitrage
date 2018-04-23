// Package errors defines a custom error type to include http errors
package errors

type MyError struct {
	Err string
	ErrorCode int
}

func (e *MyError) Error() string {
	return e.Err
}
