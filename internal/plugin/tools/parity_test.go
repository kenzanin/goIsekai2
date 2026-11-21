package tools

import "testing"

func TestDispatchStorageParity(t *testing.T) {
	ts := &ToolSet{Store: testStore(t), PluginID: "parity"}

	if _, err := ts.Dispatch("storage.set", "greeting", "hello world"); err != nil {
		t.Fatal(err)
	}
	got, err := ts.Dispatch("storage.get", "greeting")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello world" {
		t.Fatalf("storage.get roundtrip: got %v, want hello world", got)
	}

	if _, err := ts.Dispatch("storage.set", "greeting", "updated"); err != nil {
		t.Fatal(err)
	}
	got, err = ts.Dispatch("storage.get", "greeting")
	if err != nil {
		t.Fatal(err)
	}
	if got != "updated" {
		t.Fatalf("storage.set overwrite: got %v, want updated", got)
	}

	if _, err := ts.Dispatch("storage.get", "missing-key"); err == nil {
		t.Fatal("storage.get of absent key: want error, got nil")
	}
}

func TestDispatchURLEncodeParity(t *testing.T) {
	got, err := ts("url_encode", "a b&c=d")
	if err != nil {
		t.Fatal(err)
	}
	if want := "a+b%26c%3Dd"; got != want {
		t.Fatalf("url_encode: got %v, want %s", got, want)
	}
}

func TestDispatchCryptoHashParity(t *testing.T) {
	for _, tc := range []struct{ alg, want string }{
		{"sha1", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"sha256", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
	} {
		got, err := ts("crypto.hash", tc.alg, "hello")
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("crypto.hash %s: got %v, want %s", tc.alg, got, tc.want)
		}
	}
	if _, err := ts("crypto.hash", "md5", "hello"); err == nil {
		t.Fatal("crypto.hash md5: want unsupported-algorithm error, got nil")
	}
}

// ts routes one call through Dispatch on a bare ToolSet (no store needed
// for the pure tools; storage tests build their own).
func ts(name string, args ...any) (any, error) {
	return (&ToolSet{PluginID: "parity"}).Dispatch(name, args...)
}
