package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCLIParseFlagsAndSources(t *testing.T) {
	iniPath := filepath.Join(t.TempDir(), "a.ini")
	if err := os.WriteFile(iniPath, []byte("[server]\nport = 7000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ini, err := LoadIni(iniPath)
	if err != nil {
		t.Fatal(err)
	}

	reg := testRegistry(t)
	port := Entry{Section: "server", Key: "port", Flag: "port"}

	// Explicit --port wins over the ini value.
	cli := NewCLI(reg)
	if err := cli.Parse([]string{"--config", iniPath, "--port", "9000"}); err != nil {
		t.Fatal(err)
	}
	if cli.ConfigPath != iniPath {
		t.Errorf("ConfigPath = %q, want %q", cli.ConfigPath, iniPath)
	}
	if v, ok := cli.FlagValue()(port); !ok || v != "9000" {
		t.Errorf("flag value = %q, %v; want \"9000\", true", v, ok)
	}
	snap, err := cli.Snapshot(ini)
	if err != nil {
		t.Fatal(err)
	}
	res, ok := snap.Get("server.port")
	if !ok || res.Source != SourceFlag || res.Raw != "9000" {
		t.Errorf("server.port = %+v; want flag source raw 9000", res)
	}

	// Without --port the ini layer stands; no phantom flag sources.
	cli = NewCLI(reg)
	if err := cli.Parse([]string{"--config", iniPath}); err != nil {
		t.Fatal(err)
	}
	if _, ok := cli.FlagValue()(port); ok {
		t.Error("unset --port reported as a flag source")
	}
	snap, err = cli.Snapshot(ini)
	if err != nil {
		t.Fatal(err)
	}
	if res, _ = snap.Get("server.port"); res.Source != SourceIni || res.Raw != "7000" {
		t.Errorf("server.port = %+v; want ini source raw 7000", res)
	}
}

func TestCLIDumpConfig(t *testing.T) {
	reg := testRegistry(t)

	cli := NewCLI(reg)
	if err := cli.Parse([]string{"--dump-config"}); err != nil {
		t.Fatal(err)
	}
	if !cli.DumpRequested || cli.DumpPath != "" {
		t.Errorf("bare --dump-config: requested=%v path=%q; want true, \"\"", cli.DumpRequested, cli.DumpPath)
	}

	for _, args := range [][]string{{"--dump-config", "out.ini"}, {"--dump-config=out.ini"}} {
		cli := NewCLI(reg)
		if err := cli.Parse(args); err != nil {
			t.Fatal(err)
		}
		if !cli.DumpRequested || cli.DumpPath != "out.ini" {
			t.Errorf("%v: requested=%v path=%q; want true, \"out.ini\"", args, cli.DumpRequested, cli.DumpPath)
		}
	}
}

func TestCLIParseGeneratedFlagAndUnknown(t *testing.T) {
	reg := testRegistry(t)
	cli := NewCLI(reg)
	// t.on has no explicit Flag: FlagName is "--t-on".
	if err := cli.Parse([]string{"--t-on", "false"}); err != nil {
		t.Fatal(err)
	}
	entry := Entry{Section: "t", Key: "on"}
	if v, ok := cli.FlagValue()(entry); !ok || v != "false" {
		t.Errorf("t.on flag = %q, %v; want \"false\", true", v, ok)
	}

	cli = NewCLI(reg)
	if err := cli.Parse([]string{"--nope", "1"}); err == nil {
		t.Error("unknown flag: want error")
	}
}

func TestCLIHelp(t *testing.T) {
	reg := testRegistry(t)
	for _, arg := range []string{"--help", "-h"} {
		cli := NewCLI(reg)
		if err := cli.Parse([]string{arg}); err != nil {
			t.Fatal(err)
		}
		if !cli.HelpRequested {
			t.Errorf("%s: HelpRequested = false, want true", arg)
		}
	}

	cli := NewCLI(reg)
	if err := cli.Parse(nil); err != nil {
		t.Fatal(err)
	}
	if cli.HelpRequested {
		t.Error("bare parse: HelpRequested = true, want false")
	}
}
