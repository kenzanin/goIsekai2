---
type: Fact
title: "tools crypto: lowercase-hex digests; Sign = HMAC-SHA256 alias; vectors openssl-verified"
description: "internal/plugin/tools crypto helpers pass the CGO_ENABLED=0 gate with lowercase-hex digests, stdlib-only implementations and openssl-cross-checked known vectors in crypto_test.go."
tags: [internal/plugin/tools, plugin-host, crypto]
generated: { by: agent/mcp, at: "2026-11-21T19:04:29Z" }
status: stable
---

## Evidence

```
CGO_ENABLED=0 go test ./internal/plugin/tools
ok  	goisekai2/internal/plugin/tools	0.032s
go run scripts/godoc-check.go  # clean; go vet clean; gofmt clean
```

Known-vector tests cross-checked against openssl
(`printf 'hello' | openssl dgst -sha512`, `openssl dgst -sha256 -hmac secret`,
`openssl dgst -sha256 -hmac key` over the classic fox sentence):
- sha1("hello")   = aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
- sha256("hello") = 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
- sha512("hello") = 9b71d2...c043 (full value in crypto_test.go)
- HMAC-SHA256(key, fox) = f7bc83f430...3cd8; HMAC-SHA1 likewise
- Sign("hello","secret") = 88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b

## Consequence

Digest/HMAC outputs are lowercase hex everywhere in the plugin ABI; `Sign` is a
one-line alias over `HMAC("sha256", ...)`. Unknown algorithm names are errors, so
new algorithms are a one-entry map change in `internal/plugin/tools/crypto.go`.

# Related Concepts
- [tools.ToolSet: storage delegates to plugin_state, missing keys return gorm not-found](plugin-tools-surface.md): crypto helpers are part of the same host tool surface; storage/http/text/regex/cookies facts hang off it too.
