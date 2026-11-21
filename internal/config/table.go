package config

// Defaults and knob names below mirror cppIsekai `src/config/registry.hpp`; the
// CPPISEKAI_* environment prefix becomes GOISEKAI_*.

// Knobs is the first registry set. Every knob is declared once here; defaults,
// validation and the dump are derived from this table only.
func Knobs() *Registry {
	r, err := New(
		Entry{
			Section: "server", Key: "port", Kind: KindInt,
			Default: "8090", Min: 1, Max: 65535,
			Flag: "port", Env: "GOISEKAI_PORT",
			Doc: "TCP listen port",
		},
		Entry{
			Section: "server", Key: "proxy", Kind: KindString,
			Flag: "proxy", Env: "GOISEKAI_PROXY",
			Doc: "Outbound HTTP(S) proxy URL",
		},
		Entry{
			Section: "plugin", Key: "plugins", Kind: KindString,
			Default: "data/plugins",
			Flag:    "plugins-dir", Env: "GOISEKAI_PLUGINS_DIR",
			Doc: "Plugin directory",
		},
		Entry{
			Section: "plugin", Key: "fetch_profile", Kind: KindString,
			Default: "chrome116",
			Flag:    "fetch-profile", Env: "GOISEKAI_FETCH_PROFILE",
			Doc: "curl-impersonate target; dropdown list from [plugin] fetch_profiles",
		},
		Entry{
			Section: "plugin", Key: "fetch_timeout_ms", Kind: KindInt64,
			Default: "30000", Min: 1,
			Flag: "fetch-timeout-ms", Env: "GOISEKAI_FETCH_TIMEOUT_MS",
			Doc: "HTTP fetch timeout (ms)",
		},
		Entry{
			Section: "cache", Key: "cache_max_mb", Kind: KindInt64,
			Default: "2048", Min: 1,
			Flag: "cache-max-mb", Env: "GOISEKAI_CACHE_MAX_MB",
			Doc: "Image cache cap (MiB); 2048 MiB = 2 GiB, 4096 MiB = 4 GiB",
		},
		Entry{
			Section: "cache", Key: "image_format", Kind: KindEnum,
			Default: "webp", Allowed: []string{"webp", "avif", "jxl", "none"},
			Flag: "image-format", Env: "GOISEKAI_IMAGE_FORMAT",
			Doc: "Stored image format (webp|avif|jxl|none)",
		},
		Entry{
			Section: "cache", Key: "cover_max_px", Kind: KindInt64,
			Default: "800", Min: 0,
			Flag: "cover-max-px", Env: "GOISEKAI_COVER_MAX_PX",
			Doc: "Cover longest-side cap (px)",
		},
		Entry{
			Section: "paths", Key: "data_dir", Kind: KindString,
			Default: "data",
			Flag:    "data-dir", Env: "GOISEKAI_DATA_DIR",
			Doc: "Data root",
		},
		Entry{
			Section: "paths", Key: "library", Kind: KindString,
			Default: "data/library.db",
			Flag:    "db", Env: "GOISEKAI_DB",
			Doc: "SQLite library path",
		},
		Entry{
			Section: "paths", Key: "log_file", Kind: KindString,
			Default: "data/goisekai.log",
			Flag:    "log-file", Env: "GOISEKAI_LOG_FILE",
			Doc: "Log file path",
		},
		Entry{
			Section: "log", Key: "level", Kind: KindEnum,
			Default: "info", Allowed: []string{"trace", "debug", "info", "warning", "error"},
			Flag: "log-level", Env: "GOISEKAI_LOG_LEVEL",
			Doc: "Log level (trace|debug|info|warning|error)",
		},
	)
	if err != nil {
		panic("config: invalid knob table: " + err.Error())
	}
	return r
}
