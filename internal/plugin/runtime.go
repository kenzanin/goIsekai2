package plugin

import (
	"context"
	"encoding/json"
	"fmt"
)

// HostAPI is the envelope contract version every plugin request must carry.
const HostAPI = 1

// Envelope is the JSON request frame the host sends to a plugin.
type Envelope struct {
	API    int             `json:"api"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// Runtime is implemented by every plugin engine.
type Runtime interface {
	// Call decodes one request envelope, executes it, and returns the raw
	// response bytes.
	Call(ctx context.Context, request []byte) ([]byte, error)
}

// AllowedMethods is the fixed plugin API surface accepted by ValidateMethod.
var AllowedMethods = []string{"search", "detail", "chapters", "pages", "genres"}

// ValidateMethod rejects any method outside AllowedMethods.
func ValidateMethod(method string) error {
	for _, allowed := range AllowedMethods {
		if allowed == method {
			return nil
		}
	}
	return fmt.Errorf("plugin: unknown method %q", method)
}
