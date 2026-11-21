package plugin

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/qmuntal/stateless"
	"gorm.io/gorm"

	"goisekai2/internal/events"
	"goisekai2/internal/store"
)

// Lifecycle states and triggers use the baseline vocabulary.
const (
	// StateDisabled means the plugin is installed but inert.
	StateDisabled = "Disabled"
	// StateEnabled is the superstate allowing calls.
	StateEnabled = "Enabled"
	// StateIdle is the enabled state with no call in flight.
	StateIdle = "Enabled.Idle"
	// StateBusy is the enabled state with one call in flight.
	StateBusy = "Enabled.Busy"
	// StateFailed means the entry could not be loaded.
	StateFailed = "Failed"

	// TriggerEnable activates a plugin.
	TriggerEnable = "Enable"
	// TriggerDisable deactivates a plugin.
	TriggerDisable = "Disable"
	// TriggerCallStarted marks a call as in flight.
	TriggerCallStarted = "CallStarted"
	// TriggerCallFinished completes a successful call.
	TriggerCallFinished = "CallFinished"
	// TriggerCallFailed completes a failed call.
	TriggerCallFailed = "CallFailed"
	// TriggerMarkFailed marks a plugin broken.
	TriggerMarkFailed = "MarkFailed"
)

// IllegalTriggerError reports a trigger the current state does not permit.
type IllegalTriggerError struct {
	State   string
	Trigger string
}

// Error names the offending pair.
func (e *IllegalTriggerError) Error() string {
	return fmt.Sprintf("plugin: illegal trigger %s in state %s", e.Trigger, e.State)
}

// Lifecycle owns one plugin's FSM and its persisted state.
type Lifecycle struct {
	sm       *stateless.StateMachine
	store    *store.Store
	bus      *events.Bus
	pluginID string
	mu       sync.Mutex
	illegal  atomic.Uint64
}

// NewLifecycle builds the machine from the recorded state, defaulting to Disabled.
func NewLifecycle(st *store.Store, bus *events.Bus, pluginID, initial string) *Lifecycle {
	if initial == "" {
		initial = StateDisabled
	}
	l := &Lifecycle{store: st, bus: bus, pluginID: pluginID}
	l.sm = stateless.NewStateMachine(stateless.State(initial))

	l.sm.Configure(stateless.State(StateDisabled)).
		Permit(stateless.Trigger(TriggerEnable), stateless.State(StateEnabled))
	l.sm.Configure(stateless.State(StateEnabled)).
		InitialTransition(stateless.State(StateIdle)).
		Permit(stateless.Trigger(TriggerDisable), stateless.State(StateDisabled)).
		Permit(stateless.Trigger(TriggerMarkFailed), stateless.State(StateFailed))
	l.sm.Configure(stateless.State(StateIdle)).SubstateOf(stateless.State(StateEnabled)).
		Permit(stateless.Trigger(TriggerCallStarted), stateless.State(StateBusy)).
		Permit(stateless.Trigger(TriggerDisable), stateless.State(StateDisabled))
	l.sm.Configure(stateless.State(StateBusy)).SubstateOf(stateless.State(StateEnabled)).
		Permit(stateless.Trigger(TriggerCallFinished), stateless.State(StateIdle)).
		Permit(stateless.Trigger(TriggerCallFailed), stateless.State(StateIdle)).
		Permit(stateless.Trigger(TriggerDisable), stateless.State(StateDisabled))
	l.sm.Configure(stateless.State(StateFailed)).
		Permit(stateless.Trigger(TriggerEnable), stateless.State(StateEnabled))

	l.sm.OnUnhandledTrigger(func(_ context.Context, state, trigger stateless.State, _ []string) error {
		l.illegal.Add(1)
		return &IllegalTriggerError{State: fmt.Sprint(state), Trigger: fmt.Sprint(trigger)}
	})
	return l
}

// Fire applies one trigger, persists the resulting state, commits, then publishes.
func (l *Lifecycle) Fire(ctx context.Context, trigger string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.sm.FireCtx(ctx, stateless.Trigger(trigger)); err != nil {
		return err
	}
	state, err := l.sm.State(ctx)
	if err != nil {
		return err
	}
	value := fmt.Sprint(state)
	if err := l.store.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		if err := tx.WithContext(txCtx).Table("plugin_state").
			Where("plugin_id = ? AND key = ?", l.pluginID, "lifecycle").Delete(nil).Error; err != nil {
			return err
		}
		return tx.WithContext(txCtx).Table("plugin_state").Create(map[string]any{
			"plugin_id": l.pluginID, "key": "lifecycle", "value": value,
		}).Error
	}); err != nil {
		return err
	}
	if l.bus != nil {
		l.bus.Publish(events.PluginLifecycle, events.PluginLifecyclePayload{PluginID: l.pluginID, State: value})
	}
	return nil
}

// State returns the current FSM state.
func (l *Lifecycle) State(ctx context.Context) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	state, err := l.sm.State(ctx)
	return fmt.Sprint(state), err
}

// IllegalCount reports how many illegal triggers were rejected.
func (l *Lifecycle) IllegalCount() uint64 { return l.illegal.Load() }

// IsIllegal reports whether err is an illegal-trigger rejection.
func IsIllegal(err error) bool {
	var illegal *IllegalTriggerError
	return errors.As(err, &illegal)
}
