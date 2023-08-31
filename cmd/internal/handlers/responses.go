package handlers

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OkResponse() Response {
	return Response{
		Status: StatusOK,
	}
}

func MakeResponseError(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}
