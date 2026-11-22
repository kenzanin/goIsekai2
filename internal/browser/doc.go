// Package browser defines the host-side contract for driving a browser
// automation session.
//
// A Browser is one live session. Call dispatches a named tool with arbitrary
// arguments and returns the tool's result as text; Close ends the session.
// Concrete transports implement the interface in their own packages; this
// package carries no dependencies beyond the standard library.
//
// Callers match the sentinels with errors.Is.
package browser
