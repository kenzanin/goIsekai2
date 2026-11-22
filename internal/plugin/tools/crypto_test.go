package tools

import "testing"

func TestHash(t *testing.T) {
	cases := []struct{ alg, in, want string }{
		{"md5", "hello", "5d41402abc4b2a76b9719d911017c592"},
		{"sha1", "hello", "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"},
		{"sha256", "hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"sha512", "hello", "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
	}
	for _, c := range cases {
		got, err := Hash(c.alg, c.in)
		if err != nil || got != c.want {
			t.Errorf("Hash(%q,%q) = %q,%v want %q", c.alg, c.in, got, err, c.want)
		}
	}
	if _, err := Hash("bogus", "x"); err == nil {
		t.Error("Hash(unknown) should error")
	}
}

func TestHMAC(t *testing.T) {
	got, err := HMAC("sha256", "key", "The quick brown fox jumps over the lazy dog")
	if err != nil || got != "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8" {
		t.Errorf("HMAC sha256 vector = %q,%v", got, err)
	}
	got, err = HMAC("sha1", "key", "The quick brown fox jumps over the lazy dog")
	if err != nil || got != "de7c9b85b8b78aa6bc8a7a36f70a90701c9db4d9" {
		t.Errorf("HMAC sha1 vector = %q,%v", got, err)
	}
	if _, err := HMAC("bogus", "k", "v"); err == nil {
		t.Error("HMAC(unknown) should error")
	}
}

func TestSign(t *testing.T) {
	got, err := Sign("hello", "secret")
	if err != nil || got != "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b" {
		t.Errorf("Sign vector = %q,%v", got, err)
	}
}

func TestBase64HexRoundTrip(t *testing.T) {
	if got := Base64Encode("hello"); got != "aGVsbG8=" {
		t.Errorf("Base64Encode = %q", got)
	}
	if got, _ := Base64Decode("aGVsbG8="); got != "hello" {
		t.Errorf("Base64Decode = %q", got)
	}
	if _, err := Base64Decode("!!!"); err == nil {
		t.Error("Base64Decode(invalid) should error")
	}
	if got := HexEncode("hello"); got != "68656c6c6f" {
		t.Errorf("HexEncode = %q", got)
	}
	if got, _ := HexDecode("68656c6c6f"); got != "hello" {
		t.Errorf("HexDecode = %q", got)
	}
	if _, err := HexDecode("zz"); err == nil {
		t.Error("HexDecode(invalid) should error")
	}
}
