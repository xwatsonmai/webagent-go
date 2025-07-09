package event

type ErrorEvent struct {
	err error
}

func NewErrorEvent(err error) IEvent {
	return ErrorEvent{
		err: err,
	}
}

func (e ErrorEvent) Type() Type {
	return TypeError
}

func (e ErrorEvent) Content() string {
	return e.err.Error()
}
