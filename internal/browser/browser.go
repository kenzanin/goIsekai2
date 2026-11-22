package browser

import "errors"

// Browser is one live browser automation session.
type Browser interface {
	// Call invokes the named tool with args and returns the tool's result
	// as text.
	Call(tool string, args any) (string, error)

	// Close ends the session.
	Close() error
}

// ErrUnknownTool is returned by Call when tool does not name a known tool.
var ErrUnknownTool = errors.New("browser: unknown tool")

// ErrSessionExpired is returned when the session lifetime has ended and a
// new session must be started.
var ErrSessionExpired = errors.New("browser: session expired")

// ErrDead is returned when the browser process is gone and cannot be
// reached again.
var ErrDead = errors.New("browser: dead")
