# Tasks: atsumaru-newest-release

## 1. Plugin

- [x] 1.1 Keep the newest duplicate chapter release in `plugins/lua/atsumaru/main.lua` — verify `/api/chapters` for `AJDs` returns `AJDs:2V2CqE` for Chapter 1.

## 2. Live verification

- [x] 2.1 Restart the daemon, confirm the reported reader URL shows the upstream page set.

Gate: `make test lint`.
