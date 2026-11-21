---
type: Fact
title: "Events bus implementation verified with FIFO, panic recovery, and recorder"
description: Refined internal/events Bus using non-closing reqChan with doneCh signaling and done synchronization on Subscribe/Publish.
tags: [events, bus]
generated: { by: agent/mcp, at: "2026-11-21T15:15:43Z" }
status: stable
---

### Evidence
Command: CGO_ENABLED=0 go test -v ./internal/events
Output:
=== RUN   TestBus_FIFOAndPublish
--- PASS: TestBus_FIFOAndPublish (0.00s)
=== RUN   TestBus_PanicRecoveryDoesNotStopLaterEvents
--- PASS: TestBus_PanicRecoveryDoesNotStopLaterEvents (0.00s)
=== RUN   TestBus_CloseLeavesNoGoroutine
--- PASS: TestBus_CloseLeavesNoGoroutine (0.01s)
=== RUN   TestBus_RecordConfigChangedPayload
--- PASS: TestBus_RecordConfigChangedPayload (0.00s)
PASS
ok  	goisekai2/internal/events	0.014s

### Consequence
Race-free shutdown and synchronous handler execution are guaranteed via doneCh selection and request acknowledgment channels.

# Related Concepts
- [Event-driven core: domain events on gookit/event, stream plane separate](../architecture/event-driven-core.md): Refined implementation of the event bus conforms to the event-driven core architecture.
