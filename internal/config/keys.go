package config

// Configuration knob keys, addressed by dotted path.

const (
	// KnobServerPort is the HTTP listen port of the daemon.
	KnobServerPort = "server.port"
	// KnobProxy is the upstream proxy URL used for plugin fetches.
	KnobProxy = "server.proxy"
	// KnobLogLevel is the minimum level emitted by the daemon logger.
	KnobLogLevel = "log.level"
)
