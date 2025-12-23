package autho

import "time"

type DefaultObj struct {
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type ErrorResult struct {
	IsSuccess bool
	Error     error
}

type User struct {
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Password    string     `json:"password"`
	DefaultObj  DefaultObj `json:"objInfo"`
	ErrorResult ErrorResult
}

func ResultSuccess() ErrorResult {
	return ErrorResult{IsSuccess: true, Error: nil}
}

func ResultError(err error) ErrorResult {
	return ErrorResult{IsSuccess: false, Error: err}
}
