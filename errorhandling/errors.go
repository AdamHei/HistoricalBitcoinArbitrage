package errorhandling

type MyError struct {
	Err string
	ErrorCode int
}

func (e *MyError) Error() string {
	return e.Err
}
