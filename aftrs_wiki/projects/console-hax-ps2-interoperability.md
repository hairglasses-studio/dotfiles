# Console‑Hax PS2 Interoperability

A unified workflow for building and running PS2 homebrew across repos in `~/Docs/console-hax`.

## Shared scripts

Located at `~/Docs/console-hax/scripts/`:

- docker_build.sh — Build current repo in `ghcr.io/ps2dev/ps2dev:latest`
- run_ps2link.sh — Run an ELF on real PS2 via `ps2client execee host:<elf>`
- fetch_irx.sh — Copy common IRX modules from `PS2SDK` into a local folder
- common.sh — Logging helpers (uses `gum` when available)

Make sure they are executable: `chmod +x ~/Docs/console-hax/scripts/*.sh`.

## Makefile targets convention

Each PS2 project should expose:

```
.PHONY: docker-build
docker-build:
	../scripts/docker_build.sh --target all

.PHONY: run-ps2link
run-ps2link: $(EE_BIN)
	PS2_IP?=192.168.1.42
	../scripts/run_ps2link.sh $(EE_BIN)
```

- Build: `make docker-build`
- Run on PS2: `PS2_IP=<ip> make run-ps2link`

## IRX modules

Use `fetch_irx.sh` to populate `iop/` with IRX modules from your SDK install:

```
./tools/fetch_irx.sh    # or ../scripts/fetch_irx.sh iop
```

When not using hostfs, keep IRX near your ELF and set `NET_IRX_PATH` accordingly.

## Notes

- Prefer `ghcr.io/ps2dev/ps2dev:latest` for reproducible builds.
- Ensure `PS2DEV`, `PS2SDK`, and `GSKIT` are exported when building natively.
- For network runs, provide uppercase `IPCONFIG.DAT` and aligned IRX versions.


