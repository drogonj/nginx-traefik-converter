package errors

func (e *ConverterError) Error() string {
	return e.Message
}
