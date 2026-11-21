package config

import (
	"flag"
	"io"
	"os"
	"strings"
)

// CLI parses command-line flags on top of a Registry.
type CLI struct {
	ConfigPath    string
	DumpRequested bool
	DumpPath      string
	HelpRequested bool

	reg  *Registry
	fs   *flag.FlagSet
	set  map[string]string // flag key -> value, only for flags given on the command line
	help *bool
	h    *bool
}

// dumpValue records the --dump-config argument; a bare flag yields "".
type dumpValue struct{ path *string }

// String reports the current dump path.
func (d dumpValue) String() string {
	if d.path == nil {
		return ""
	}
	return *d.path
}

// Set records the given path; the bare flag yields the empty string.
func (d dumpValue) Set(s string) error {
	*d.path = s
	return nil
}

// flagKey maps an entry to its flag.FlagSet key: FlagName minus the
// display-only leading dashes, since flag strips dashes when parsing.
func flagKey(e Entry) string {
	return strings.TrimPrefix(e.FlagName(), "--")
}

// NewCLI returns a CLI that resolves flags against r.
func NewCLI(r *Registry) *CLI {
	c := &CLI{reg: r, set: map[string]string{}}
	fs := flag.NewFlagSet("goisekai", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&c.ConfigPath, "config", "", "path to the INI configuration file")
	fs.Var(dumpValue{&c.DumpPath}, "dump-config", "write the resolved configuration and exit; optional target path")
	c.help = fs.Bool("help", false, "show usage and exit")
	c.h = fs.Bool("h", false, "alias for -help")
	for _, e := range r.Entries() {
		fs.String(flagKey(e), "", e.Doc)
	}
	c.fs = fs
	return c
}

// Parse parses args. Unknown flags return the flag package's error.
func (c *CLI) Parse(args []string) error {
	// flag demands a value for non-bool flags, but a bare --dump-config must
	// default to "": rewrite it to the = form when no path argument follows.
	normalized := append([]string(nil), args...)
	for i, a := range normalized {
		if (a == "-dump-config" || a == "--dump-config") &&
			(i+1 == len(normalized) || strings.HasPrefix(normalized[i+1], "-")) {
			normalized[i] = a + "="
		}
	}
	c.set = map[string]string{}
	if err := c.fs.Parse(normalized); err != nil {
		return err
	}
	c.fs.Visit(func(f *flag.Flag) { c.set[f.Name] = f.Value.String() })
	c.HelpRequested = *c.help || *c.h
	_, c.DumpRequested = c.set["dump-config"]
	return nil
}

// FlagValue returns a resolver hook serving only flags present on the
// command line; unset flags report false so lower layers win.
func (c *CLI) FlagValue() func(Entry) (string, bool) {
	return func(e Entry) (string, bool) {
		v, ok := c.set[flagKey(e)]
		return v, ok
	}
}

// Snapshot resolves the registry with the process environment and parsed flags.
func (c *CLI) Snapshot(ini *Ini) (*Snapshot, error) {
	return Resolve(c.reg, ini, os.Getenv, c.FlagValue())
}
