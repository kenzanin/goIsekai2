//go:build tools

// Package tools pins dependencies that do not have a production import yet.
// It is never compiled into the daemon; `go mod tidy` keeps the requirements
// listed here until the store change imports them directly.
package tools

import (
	_ "github.com/glebarez/sqlite"
	_ "gorm.io/gorm"
)
