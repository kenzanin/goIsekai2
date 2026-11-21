// Command goisekai2 is the goIsekai daemon entry point.
package main

import (
	"fmt"
	"io"
	"os"

	"goisekai2/internal/config"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes one daemon start attempt and returns the process exit code: 0 on
// success, 2 on a usage error, 1 on any configuration or I/O failure.
func run(args []string, stdout, stderr io.Writer) int {
	reg := config.Knobs()
	cli := config.NewCLI(reg)

	if err := cli.Parse(args); err != nil {
		fmt.Fprintf(stderr, "goisekai2: %v\n\n", err)
		usage(stderr, reg)
		return 2
	}
	if cli.HelpRequested {
		usage(stdout, reg)
		return 0
	}

	if cli.DumpRequested {
		var w io.Writer = stdout
		var f *os.File
		if cli.DumpPath != "" {
			if _, err := os.Stat(cli.DumpPath); err == nil {
				fmt.Fprintf(stderr, "goisekai2: %s exists, refusing to overwrite\n", cli.DumpPath)
				return 1
			}
			var err error
			f, err = os.OpenFile(cli.DumpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
			if err != nil {
				fmt.Fprintf(stderr, "goisekai2: %v\n", err)
				return 1
			}
			defer f.Close()
			w = f
		}
		if err := reg.Dump(w); err != nil {
			fmt.Fprintf(stderr, "goisekai2: %v\n", err)
			return 1
		}
		return 0
	}

	if cli.ConfigPath == "" {
		fmt.Fprintf(stderr, "goisekai2: --config <path> is required\n\n")
		usage(stderr, reg)
		fmt.Fprintf(stderr, "\nrun with --help for the full flag list\n")
		return 2
	}

	ini, err := config.LoadIni(cli.ConfigPath)
	if err != nil {
		fmt.Fprintf(stderr, "goisekai2: %v\n", err)
		return 1
	}
	snap, err := cli.Snapshot(ini)
	if err != nil {
		fmt.Fprintf(stderr, "goisekai2: %v\n", err)
		return 1
	}

	// ponytail: no server yet in this change; print the resolved listen address
	// and exit. Swap for the HTTP shell when internal/http lands.
	fmt.Fprintf(stdout, "goisekai2: config %s ok, would listen on :%d\n",
		cli.ConfigPath, snap.Int(config.KnobServerPort))
	if snap.Int(config.KnobServerPort) <= 0 {
		return 1
	}
	return 0
}

// usage prints the flag list derived from the registry, so a new knob appears in
// --help without a second edit.
func usage(w io.Writer, reg *config.Registry) {
	fmt.Fprintf(w, "usage: goisekai2 --config <path> [flags]\n\nflags:\n")
	fmt.Fprintf(w, "  --config <path>        path to the INI configuration file (required)\n")
	fmt.Fprintf(w, "  --dump-config [path]   write the default configuration; stdout when no path\n")
	fmt.Fprintf(w, "  -h, --help             show this help and exit\n")
	fmt.Fprintf(w, "\nknob flags:\n")
	for _, e := range reg.Entries() {
		fmt.Fprintf(w, "  %-22s %s (default %q)\n", e.FlagName(), e.Doc, e.Default)
	}
}
