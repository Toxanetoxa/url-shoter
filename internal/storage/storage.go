package storage

import "errors"

// TODO посмотреть почему сторадж лучше описывать в месте использования
var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url not exist")
)
