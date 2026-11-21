// Command goisekai2 is the goIsekai daemon entry point.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goisekai2/internal/config"
	"goisekai2/internal/events"
	httpserver "goisekai2/internal/http"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}

// run executes one daemon start attempt and returns the process exit code: 0 on
// success, 2 on a usage error, 1 on any configuration or I/O failure.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
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

	levelVar := new(slog.LevelVar)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: levelVar})))
	config.RegisterLiveApplier("log.level", func(v string) error {
		var level slog.Level
		if err := level.UnmarshalText([]byte(v)); err != nil {
			return err
		}
		levelVar.Set(level)
		return nil
	})
	config.ApplySnapshot(snap)

	bus := events.New()
	defer bus.Close()
	srv := httpserver.New(reg, cli.ConfigPath, bus)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", snap.Int(config.KnobServerPort)))
	if err != nil {
		fmt.Fprintf(stderr, "goisekai2: %v\n", err)
		return 1
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()
	fmt.Fprintf(stdout, "goisekai2: listening on %s\n", ln.Addr())

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Fprintf(stderr, "goisekai2: shutdown: %v\n", err)
		return 1
	}
	if err := <-serveErr; err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(stderr, "goisekai2: serve: %v\n", err)
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
