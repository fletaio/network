package network

//mock network errors
var (
	ErrDialTimeout = &timeoutError{}
)

type timeoutError struct {
	error
}

func (t *timeoutError) Error() string {
	return "Dial timeout error"
}

func (t *timeoutError) Timeout() bool {
	return true
}
func (t *timeoutError) Temporary() bool {
	return false
}
