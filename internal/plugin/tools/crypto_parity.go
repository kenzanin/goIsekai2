package tools

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// AESGCMDecrypt decrypts hex-encoded iv/ciphertext/tag/key using AES-GCM.
// Only 32-byte keys and 16-byte tags are accepted, matching the baseline's
// pinned AES-256 contract.
func AESGCMDecrypt(ivHex, ctHex, tagHex, keyHex string) (string, error) {
	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		return "", fmt.Errorf("aes_gcm_decrypt: iv: %w", err)
	}
	ct, err := hex.DecodeString(ctHex)
	if err != nil {
		return "", fmt.Errorf("aes_gcm_decrypt: ciphertext: %w", err)
	}
	tag, err := hex.DecodeString(tagHex)
	if err != nil {
		return "", fmt.Errorf("aes_gcm_decrypt: tag: %w", err)
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("aes_gcm_decrypt: key: %w", err)
	}
	if len(key) != 32 {
		return "", fmt.Errorf("aes_gcm_decrypt: key must be 32 bytes")
	}
	if len(tag) != 16 {
		return "", fmt.Errorf("aes_gcm_decrypt: tag must be 16 bytes")
	}
	if len(iv) == 0 {
		return "", fmt.Errorf("aes_gcm_decrypt: iv is empty")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return "", err
	}
	plain, err := gcm.Open(nil, iv, append(append([]byte{}, ct...), tag...), nil)
	if err != nil {
		return "", fmt.Errorf("aes_gcm_decrypt: authentication failed: %w", err)
	}
	return string(plain), nil
}

func asStringMap(v any) (map[string]any, bool) {
	switch x := v.(type) {
	case map[string]any:
		return x, true
	case map[string]string:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[k] = val
		}
		return out, true
	default:
		return nil, false
	}
}

func asAnySlice(v any) ([]any, bool) {
	switch x := v.(type) {
	case []any:
		return x, true
	case []map[string]any:
		out := make([]any, 0, len(x))
		for _, item := range x {
			out = append(out, item)
		}
		return out, true
	default:
		return nil, false
	}
}

func scalarString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case bool:
		return strconv.FormatBool(x), true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case int:
		return strconv.Itoa(x), true
	default:
		return "", false
	}
}

// VRFSign implements the mangafire staged transform from cppIsekai. Params are
// sorted by key and stages are applied in order, each with its own iv, key and
// 256-byte substitution table. The result is unpadded base64url.
func VRFSign(apiPath string, params map[string]any, stages []any) (string, error) {
	if apiPath == "" {
		return "", fmt.Errorf("vrf_sign: api_path is empty")
	}
	if len(stages) == 0 {
		return "", fmt.Errorf("vrf_sign: stages must be a non-empty array")
	}

	sign := strings.TrimPrefix(apiPath, "/api")
	if len(params) > 0 {
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			v, ok := scalarString(params[k])
			if !ok {
				return "", fmt.Errorf("vrf_sign: param %q must be scalar", k)
			}
			parts = append(parts, k+"="+v)
		}
		sign += "?" + strings.Join(parts, "&")
	}

	data := []byte(sign)
	for i, raw := range stages {
		stage, ok := raw.(map[string]any)
		if !ok {
			return "", fmt.Errorf("vrf_sign: stage %d must be an object", i+1)
		}
		ivAny, ok := stage["iv"].(float64)
		if !ok || ivAny < 0 || ivAny > 255 || ivAny != float64(int(ivAny)) {
			return "", fmt.Errorf("vrf_sign: stage %d: iv must be integer 0-255", i+1)
		}
		keyText, ok := stage["key"].(string)
		if !ok {
			return "", fmt.Errorf("vrf_sign: stage %d: key must be a string", i+1)
		}
		tblText, ok := stage["tbl"].(string)
		if !ok {
			return "", fmt.Errorf("vrf_sign: stage %d: tbl must be a string", i+1)
		}
		key, err := base64.StdEncoding.DecodeString(keyText)
		if err != nil || len(key) == 0 {
			return "", fmt.Errorf("vrf_sign: stage %d: key: invalid base64", i+1)
		}
		tbl, err := base64.StdEncoding.DecodeString(tblText)
		if err != nil || len(tbl) < 256 {
			return "", fmt.Errorf("vrf_sign: stage %d: tbl: invalid base64 (need 256 bytes)", i+1)
		}
		out := make([]byte, len(data))
		prev := byte(int(ivAny))
		for j, b := range data {
			x := b ^ key[j%len(key)] ^ prev
			prev = tbl[x]
			out[j] = prev
		}
		data = out
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
