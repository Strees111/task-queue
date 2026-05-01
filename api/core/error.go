package core

import "errors"

var ErrBadArguments = errors.New("arguments are not acceptable")
var ErrAlreadyExists = errors.New("resource or task already exists")
var ErrNotFound = errors.New("resource is not found")
var ErrPoolClosed = errors.New("worker pool is closed")
var ErrPoolFull = errors.New("worker pool queue is full")
var ErrInvalidTask = errors.New("invalid task: nil")
var ErrTaskFailed = errors.New("task encountered temporary failure")
