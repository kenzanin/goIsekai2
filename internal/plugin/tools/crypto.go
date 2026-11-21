package tools

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
)

// hashers maps supported Hash/HMAC algorithm names to constructors.
var hashers = map[string]func() hash.Hash{
	"sha1":   sha1.New,
	"sha256": sha256.New,
	"sha512": sha512.New,
}

// Hash returns the lowercase hex digest of value using the named
// algorithm (sha1, sha256 or sha512); any other name is an error.
func Hash(algorithm, value string) (string, error) {
	newHash, ok := hashers[algorithm]
	if !ok {
		return "", fmt.Errorf("unsupported hash algorithm %q", algorithm)
	}
	h := newHash()
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// HMAC returns the lowercase hex HMAC of value under key using the named
// algorithm (sha1, sha256 or sha512); any other name is an error.
func HMAC(algorithm, key, value string) (string, error) {
	newHash, ok := hashers[algorithm]
	if !ok {
		return "", fmt.Errorf("unsupported hash algorithm %q", algorithm)
	}
	m := hmac.New(newHash, []byte(key))
	m.Write([]byte(value))
	return hex.EncodeToString(m.Sum(nil)), nil
}

// Base64Encode returns the standard padded base64 form of s.
func Base64Encode(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

// Base64Decode reverses Base64Encode; malformed input is an error.
func Base64Decode(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	return string(b), err
}

// HexEncode returns the lowercase hex form of s.
func HexEncode(s string) string { return hex.EncodeToString([]byte(s)) }

// HexDecode reverses HexEncode; malformed input is an error.
func HexDecode(s string) (string, error) {
	b, err := hex.DecodeString(s)
	return string(b), err
}

// Sign returns the lowercase hex HMAC-SHA256 of payload under key.
func Sign(payload, key string) (string, error) { return HMAC("sha256", key, payload) }
