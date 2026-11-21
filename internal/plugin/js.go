package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"modernc.org/quickjs"

	"goisekai2/internal/plugin/tools"
)

// JSRuntime executes plugin scripts on an embedded QuickJS VM. A script sees
// the core language only — there are no fs, process or network globals; every
// host interaction travels through the JSON request value passed to the
// script's handle function.
type JSRuntime struct {
	source string

	// Tools, when non-nil, is bound into every call VM as host.*.
	Tools *tools.ToolSet
}

// SetTools attaches the host tool set exposed to scripts; Call binds it into
// each fresh VM.
func (r *JSRuntime) SetTools(ts *tools.ToolSet) { r.Tools = ts }

// NewJSRuntime returns a runtime for source, rejecting scripts that do not
// compile.
func NewJSRuntime(source string) (*JSRuntime, error) {
	vm, err := quickjs.NewVM()
	if err != nil {
		return nil, fmt.Errorf("plugin: js: %w", err)
	}
	defer vm.Close()
	if _, err := vm.Eval(source, 0); err != nil {
		return nil, fmt.Errorf("plugin: js source: %w", err)
	}
	return &JSRuntime{source: source}, nil
}

// Call evaluates handle(<request>) on a fresh VM and returns the JSON result
// bytes. The eval is bounded by the context deadline, when one is set. A
// handle returning a Promise is rejected with "promise return unsupported".
func (r *JSRuntime) Call(ctx context.Context, request []byte) ([]byte, error) {
	if !json.Valid(request) {
		return nil, fmt.Errorf("plugin: js: invalid request JSON")
	}
	vm, err := quickjs.NewVM()
	if err != nil {
		return nil, fmt.Errorf("plugin: js: %w", err)
	}
	defer vm.Close()

	var timeout time.Duration
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
	vm.SetEvalTimeout(timeout)

	if r.Tools != nil {
		if err := bindJS(vm, r.Tools); err != nil {
			return nil, fmt.Errorf("plugin: js: %w", err)
		}
	}

	if _, err := vm.Eval(r.source, 0); err != nil {
		return nil, fmt.Errorf("plugin: js source: %w", err)
	}

	// ponytail: per-call VM and deadline-only cancellation; add a
	// vm.Interrupt watcher goroutine when mid-eval ctx.Done latency matters.
	res, err := vm.Eval(fmt.Sprintf(`(function(){var r=handle(%s);if(r instanceof Promise)throw"promise return unsupported";return JSON.stringify(r)})()`, request), 0)
	if err != nil {
		return nil, fmt.Errorf("plugin: js call: %w", err)
	}
	s, ok := res.(string)
	if !ok {
		return nil, fmt.Errorf("plugin: js: handle returned no JSON (%T)", res)
	}
	return []byte(s), nil
}
