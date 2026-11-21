# Conventions
* [Godoc comment rule for all Go source](godoc.md) - Mandatory godoc conventions for goIsekai2 Go code: package comments, name-first doc comments on every exported symbol, behavior over implementation, no banner comments, examples as tests, and the review gate that enforces it.
* [English for all artifacts, Indonesian only in chat](language.md) - Every artifact goIsekai2 produces is written in English; Indonesian is allowed only in chat replies to the user.
* [Raw SQL boundary](sql-boundary.md) - All raw SQL lives in the internal/store/sql module; scripts/sql-boundary-check enforces it in make lint.
