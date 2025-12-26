package obj

import "time"

type DefaultObj struct {
	GUID       string    `json:"guid"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
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
