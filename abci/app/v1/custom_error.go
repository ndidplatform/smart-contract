package app

type ApplicationError struct {
	Code    uint32
	Message string
}

func (e *ApplicationError) Error() string {
	return e.Message
}
