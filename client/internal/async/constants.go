package async

type AsyncResultState int

const (
	ReadyAsyncResultState AsyncResultState = iota
	PendingAsyncResultState
	DoneAsyncResultState
)
