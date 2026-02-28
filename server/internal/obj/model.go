package obj

import "time"

type Status int

const (
	StatusFailed  Status = -1
	StatusPending Status = 0
	StatusActive  Status = 1
)

var StatusLabels = map[Status]string{
	StatusFailed:  "failed",
	StatusPending: "pending",
	StatusActive:  "active",
}

type DefaultObj struct {
	GUID       string    `json:"guid"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
	Status     Status    `json:"status"`
}

func NewDefaultObj(guid string) DefaultObj {
	return DefaultObj{
		GUID:   guid,
		Status: StatusActive,
	}
}

type ErrorResult struct {
	IsSuccess bool
	Error     error
}

func ResultSuccess() ErrorResult {
	return ErrorResult{IsSuccess: true, Error: nil}
}

func ResultError(err error) ErrorResult {
	return ErrorResult{IsSuccess: false, Error: err}
}
