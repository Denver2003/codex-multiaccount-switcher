package domain

import "errors"

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrUsage          = errors.New("usage error")

	ErrAuthNotFound     = errors.New("active auth file not found")
	ErrAuthUnreadable   = errors.New("active auth file unreadable")
	ErrInvalidAuth      = errors.New("invalid auth")
	ErrMetadataNotFound = errors.New("metadata not found")
	ErrInvalidMetadata  = errors.New("invalid metadata")
	ErrProfileNotFound  = errors.New("profile not found")
)
