package tools

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"testing"
)

const vrfK1 = "0Ec58JOY3uBzJK9m3zqIOpdlF7UFiax9DmA="
const vrfT1 = "yINlmUNho8VYJT+ibTIP+9ESiULpVEtMOoD6U6lRE0R/xwXo/Xp9NrUgC4cw/Lmo33vUyjUE40kUoEWIr/fxfNNcq2s79ShQ5NhNrFnJ4hXPwOu/SuXzIbuTQKGFvfm08E9jvCfqAtoDqvQq3dVWPQFmJjgvkISBeXY3BgANR+yVnjGbcxZ47d6kLNfZPIayTq3/YGySb1KuVZodWp/WGNAO5pfMcpaK53Hhs0allBszaMaxuouOwdxbwgxIw6YunSsXjI05Yi0j9j4eHKfSXR8Ifo/Od+8iamRfCXTyvm7NGRGYdcQ0ywcK/u6RXhrbcCm4t2eCtrDgQVecJGkQ+A=="
const vrfK2 = "AAdjb1iPY8CiDmq9H34tKTBF8a3oDQ=="
const vrfT2 = "IUFltCxD3Oc2cwCgkJffthaOg9cgPUb0LgW6H/VtfcF0kc5F25t+aWj6JH9VOhOaY0rAFdUxlDnl5BLNvwEJvQtP5qcw7vdb/K+chnbwnspSHT8mz5lqwz41TezG0hkO06FTjJZhsyNuFLDpD2ZZxQj/QIRcF90zpmQ7Byu483WsQqUE0C342HL+JXngRB6fRzxRyVTaKu83h7UYTJ0QMt6ixFh6S3F8gqkKwrGTL3jHNBsD45UnifK8+RGtishQV2K3rujLKEkiZxpr2dYcudFW4oFsDKhad3CLBvuyTqsCo4B7mL5IKQ1vXo/MOOvq1I1d8ar9X6Ttu5KF4fZgiA=="
const vrfK3 = "DELOJgPsVaCcblDtTGMdHzM="
const vrfT3 = "NQHlu1/wVO5EmkwQymF810qqY2xG1k2obcas4Z9mCsPEIFl9pRIjFxbJ7ybMHbBckT5Ton85E0FOeHezbh/mjlEYpmpnlXOS8dgrqeq2KfxImTh1YK9y0PeMNhzA1OQzSY9brYOJq/l2QnE/hwOeZIhPixVSKIUlDb5vLcH6RWKxkIEMuP0bDwIqQ71AJJaEaMJL7A6YtyIwoRT+L5v4aZzodN/0+3nOGsfblFjgxSfPzVDjNFeNl5P26+kEC/8AHgdrpAbt3hHz3HrRN1Y6e+JHgF7ncFWnoF0y3THL1S71WgWGCa6KtSzTCCG58n68nTyj2T3Sshk7utqCtMi/ZQ=="

func vrfStages() []any {
	return []any{
		map[string]any{"iv": float64(0x5A), "key": vrfK1, "tbl": vrfT1},
		map[string]any{"iv": float64(0x35), "key": vrfK2, "tbl": vrfT2},
		map[string]any{"iv": float64(0xBA), "key": vrfK3, "tbl": vrfT3},
	}
}

func TestCryptoBaselineNames(t *testing.T) {
	ts := &ToolSet{}
	md5, err := ts.Dispatch("crypto.md5_hex", "hello")
	if err != nil || md5 != "5d41402abc4b2a76b9719d911017c592" {
		t.Fatalf("md5 = %v, %v", md5, err)
	}
	sha, err := ts.Dispatch("crypto.sha256_hex", "hello")
	if err != nil || sha != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("sha256 = %v, %v", sha, err)
	}
	hmac, err := ts.Dispatch("crypto.hmac_sha256_hex", "hello", "key")
	if err != nil || hmac != "9307b3b915efb5171ff14d8cb55fbcc798c6c0ef1456d66ded1a6aa723a58b7b" {
		t.Fatalf("hmac = %v, %v", hmac, err)
	}
	for _, pair := range [][2]string{
		{"codecs.base64_encode", "crypto.b64enc"},
		{"codecs.hex_encode", "crypto.hexenc"},
	} {
		a, err := ts.Dispatch(pair[0], "hello")
		if err != nil {
			t.Fatal(err)
		}
		b, err := ts.Dispatch(pair[1], "hello")
		if err != nil {
			t.Fatal(err)
		}
		if a != b {
			t.Errorf("%s = %v, %s = %v", pair[0], a, pair[1], b)
		}
	}
}

func TestVrfGoldenVectors(t *testing.T) {
	ts := &ToolSet{}
	cases := []struct {
		path   string
		params map[string]any
		want   string
	}{
		{"/titles", map[string]any{}, "8sK3xtqdFQ"},
		{"/manga", map[string]any{"limit": "50"}, "8qEkqsbjQMrPN82PZEAP"},
		{"/titles", map[string]any{"keyword": "naruto", "limit": "50", "page": "1"}, "8sK3xtqdFZfetBhus6bRAjHyEL3W3G_rnxgOXpPB0_i8tncqWLs"},
		{"/api/titles", map[string]any{"page": "2"}, "8sK3xtqdFZfnfxImVdE"},
		{"/search", map[string]any{"q": "はじめの一歩"}, "8rmP5f_9tOdgg-j3a5w_WQte2iUPEj6pvQYSPw"},
	}
	for _, tc := range cases {
		got, err := ts.Dispatch("crypto.vrf_sign", tc.path, tc.params, vrfStages())
		if err != nil {
			t.Fatalf("%s: %v", tc.path, err)
		}
		if got != tc.want {
			t.Errorf("vrf %s = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestAESGCMRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	nonce := []byte("0123456789ab")
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, len(nonce))
	if err != nil {
		t.Fatal(err)
	}
	sealed := gcm.Seal(nil, nonce, []byte("secret payload"), nil)
	ct, tag := sealed[:len(sealed)-16], sealed[len(sealed)-16:]
	ts := &ToolSet{}
	got, err := ts.Dispatch("crypto.aes_gcm_decrypt", hex.EncodeToString(nonce), hex.EncodeToString(ct), hex.EncodeToString(tag), hex.EncodeToString(key))
	if err != nil || got != "secret payload" {
		t.Fatalf("decrypt = %v, %v", got, err)
	}
	tag[len(tag)-1] ^= 1
	if _, err := ts.Dispatch("crypto.aes_gcm_decrypt", hex.EncodeToString(nonce), hex.EncodeToString(ct), hex.EncodeToString(tag), hex.EncodeToString(key)); err == nil {
		t.Fatal("tampered tag unexpectedly decrypted")
	}
}
