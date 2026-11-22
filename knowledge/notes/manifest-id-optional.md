---
type: Fact
title: Manifest id is optional; folder is the fallback identity
description: "cppIsekai plugin loaders default a main.json without id to the folder name, and goIsekai2 LoadManifest rejects only a declared id that differs from the folder."
tags: [plugin, manifest, abi, baseline, parity]
generated: { by: agent/mcp, at: "2026-11-22T05:33:09Z" }
status: stable
---

cppIsekai sets the runtime identity in plugin_load.cpp scan_dir via `rt.info.id = m.id.value_or(folder)`: a manifest without `id` is valid and inherits its folder name, and the loader applies no folder/id equality check.

goIsekai2 mirrors this in internal/plugin/manifest.go LoadManifest: the folder/id rejection fires only when the manifest declares an id, so id-less manifests keep the baseline folder-identity behavior; discovery additionally keys Entry.ID by folder (internal/plugin/discovery.go).

# Related Concepts
- [First site corpus tranche and parity tracking](../decisions/site-corpus-scope.md): Manifest id validation evidence for the site corpus tranche
