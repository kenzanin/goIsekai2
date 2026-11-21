package plugin

import (
	"context"
	"fmt"
	"io"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// wasmRequestOffset is where the host writes the JSON request in linear memory.
const wasmRequestOffset uint32 = 1024

// WasmRuntime executes a WASI preview1 module exporting handle.
type WasmRuntime struct {
	module []byte
}

// NewWasmRuntime validates the module and its handle export.
func NewWasmRuntime(module []byte) (*WasmRuntime, error) {
	ctx := context.Background()
	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigCompiler())
	defer r.Close(ctx)
	compiled, err := r.CompileModule(ctx, module)
	if err != nil {
		return nil, fmt.Errorf("plugin: wasm compile: %w", err)
	}
	if _, ok := compiled.ExportedFunctions()["handle"]; !ok {
		return nil, fmt.Errorf("plugin: wasm: missing handle export")
	}
	return &WasmRuntime{module: module}, nil
}

// Call writes request bytes into linear memory and invokes handle(len). The
// i64 result packs response pointer in the high 32 bits and length in the low.
func (r *WasmRuntime) Call(ctx context.Context, request []byte) ([]byte, error) {
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigCompiler().WithCloseOnContextDone(true))
	defer rt.Close(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	mod, err := rt.InstantiateWithConfig(ctx, r.module, wazero.NewModuleConfig().
		WithStdout(io.Discard).WithStderr(io.Discard))
	if err != nil {
		return nil, fmt.Errorf("plugin: wasm instantiate: %w", err)
	}
	handle := mod.ExportedFunction("handle")
	if handle == nil {
		return nil, fmt.Errorf("plugin: wasm: missing handle export")
	}
	mem := mod.Memory()
	if mem == nil {
		return nil, fmt.Errorf("plugin: wasm: no exported memory")
	}
	if !mem.Write(wasmRequestOffset, request) {
		return nil, fmt.Errorf("plugin: wasm: request too large")
	}
	results, err := handle.Call(ctx, uint64(len(request)))
	if err != nil {
		return nil, fmt.Errorf("plugin: wasm call: %w", err)
	}
	packed := results[0]
	ptr, size := uint32(packed>>32), uint32(packed)
	out, ok := mem.Read(ptr, size)
	if !ok {
		return nil, fmt.Errorf("plugin: wasm: response out of bounds")
	}
	return append([]byte(nil), out...), nil
}
