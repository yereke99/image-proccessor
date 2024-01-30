package domain

import "errors"

var (
	ErrCloudNumberNotDefined = errors.New("cloud number not defined")
	ErrConvertStringToInt    = errors.New("string to number conversion error")
	ErrStagePullTask         = errors.New("error pull task stage")
	ErrEmptyParams           = errors.New("error empty config")
	ErrStatusCode            = errors.New("error status code")
)
