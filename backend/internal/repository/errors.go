package repository

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrTicketNotFound       = errors.New("ticket not found")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
)
