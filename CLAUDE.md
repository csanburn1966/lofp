# LoFP — Legends of Future Past

Resurrecting a 1990s MUD from original script files. The original game engine source code is lost; only the content scripts and documentation survive. We reverse-engineer a working game from those.

## Architecture

- **Backend**: Go + gorilla/mux + MongoDB at `engine/`
- **Frontend**: React 19 + TypeScript + Vite + Tailwind 4 at `frontend/`
- **Original Scripts**: `original/scripts/` (read-only reference, 333 .SCR files)
- **Game captures**: 'original/csanburn/' (read-only reference, .TXT files)
- Frontend dev server: port 4992, Backend server: port 4993
- Start both: `./start.sh`
- Start both Windows: './start.bat'

## Build Check

```sh
cd engine && go build ./...
cd frontend && npx tsc --noEmit
```

## Reference Documentation

These original documents are the primary sources for game mechanics and world data:

| Document | Path | Description |
|----------|------|-------------|
| GM Manual | `original/GM Pages/MANUAL.DOC` | Comprehensive GM reference: combat, stats, items, monsters, skills, XP tables, all GM commands |
| GM Script Guide | `original/GMSCRIPT.DOC` | Script language reference: room/item/monster definitions, IFVERB/IFENTRY/IFSAY blocks, variables |
| Player Manual | `original/legends/LEGENDS.DOC` | Player-facing docs: commands, spells, psionics, skills, races, crafting |
| Session Capture | `original/legends/shirla.cap` | 1996 gameplay session capture — invaluable for combat output format, spell interactions, original message text |
| Script Config | `original/scripts/LEGENDS.CFG` | Master config listing all .SCR files to load in order |
| GM Pages | `original/GM Pages/` | Additional GM reference materials (various .DOC files) |

When implementing new features, always cross-reference these documents and the session capture to match original behavior.

## Multi-Machine Coordination

Production is currently a single server, but the engine is designed to run as multiple instances. ALL mutable world state must be coordinated via the MongoDB-backed hub (`engine/internal/hub/`):

- **Messages**: broadcasts, whispers, global announcements → published to `events` collection, delivered via Change Streams
- **Player presence**: WHO list, room occupancy → `presence` collection with TTL heartbeat
- **Room state changes**: item open/close/lock, item drops/pickups, script mutations (vals, itembits) → `room_state_change` events via hub
- **Any new mutable state** added to rooms, items, or global data MUST call `notifyRoomChange()` or publish through the hub, or players on different machines will see inconsistent worlds

## After Server Changes

After making changes to the backend (engine/), restart the Go server:
```sh
kill $(lsof -ti:4993) 2>/dev/null; sleep 1; cd engine && go run cmd/lofp/main.go &
```
Load .env first if needed: `source .env`

## Environments and Deployment

Fly.io is not used; `fly.toml` is left over from upstream.

### Dev: VM 9011 `lofp-dev` (192.168.4.158)

- Runs in Docker Compose from the repo clone at `~/lofp`: `deploy/docker-compose.yml` (project name `lofp-deploy`), with secrets in `deploy/.env` (gitignored; template in `deploy/.env.example`).
- Two containers: `lofp` (built from the repo's `Dockerfile`) and `mongo:7` as a single-node replica set `rs0` (the hub needs change streams).
- Ports: web 8080, telnet 4000, SSH game access 4022.
- SSH from aiportal: `ssh -i ~/.ssh/claude_proxmox claude@192.168.4.158`
- Update to the latest pushed code:
  ```sh
  cd ~/lofp && git pull && cd deploy && sudo docker compose up -d --build
  ```
- No email service, Google OAuth, or Turnstile keys are configured. Email/password login works in the web client. Telnet/SSH logins need a verified email, so mark accounts verified by hand:
  ```sh
  cd ~/lofp/deploy && sudo docker compose exec mongo mongosh lofp --eval 'db.accounts.updateOne({email:"you@example.com"},{$set:{emailVerified:true}})'
  ```
- The VM's CPU type is `host` because MongoDB 5+ needs AVX.

### Production: `lofp` (192.168.4.123), lofp.sanburnlab.com

- Not Docker. Source tree at `/home/csanburn/lofp`, MongoDB runs natively (`mongod.service`). Two systemd services, both as user `csanburn`:
  - `lofp`: the Go engine, `ExecStart=/home/csanburn/lofp/engine/cmd/lofp/lofp`, `WorkingDirectory=/home/csanburn/lofp/engine`, `EnvironmentFile=/home/csanburn/lofp/.env`. It listens on 4993 (web API/WebSocket), 4000 (telnet), and 4022 (SSH).
  - Uses `config/prod.yaml` via `.env`: `LOFP_CONFIG=config/prod.yaml`, `LOFP_PORT=4993` (Vite proxies to 4993), `LOFP_SCRIPTS_DIR=../original/scripts`, `LOFP_CONFIG_FILE=../original/scripts/LEGENDS.CFG`, plus `MONGODB_URI`, `FRONTEND_URL`, `EMAIL_FROM`, and `JWT_SECRET`. `prod.yaml` has no default for `JWT_SECRET`; it must be at least 32 characters or the engine panics at startup, and if empty, auth (web login, SSH) is disabled. Generate it at a shell prompt: `echo "JWT_SECRET=$(openssl rand -hex 32)" >> .env`.
  - `dev.yaml` on production is no longer used, so copying the repo's `dev.yaml` over it is harmless.
  - After a crash loop, `sudo systemctl reset-failed lofp` before restarting.
  - `lofp-frontend`: the **Vite dev server** (`npx vite --port 4992`) serving `frontend/` straight from source and proxying `/api/`, `/healthz`, and `/ws` to `localhost:4993` (see `frontend/vite.config.ts`, which allows the host `lofp.sanburnlab.com`). There is no frontend build step: copied `.ts`/`.tsx` changes are picked up live. A new npm dependency still needs `npm install` in `frontend/` and a restart of `lofp-frontend`.
- Deploys are manual: changed and new source files are copied to the server with WinSCP, then on the server (only needed for engine changes):
  ```sh
  cd ~/lofp/engine/cmd/lofp && sudo systemctl stop lofp && go build -o lofp . && sudo systemctl start lofp
  ```
- Because files are copied by hand, when a change is ready for production, list every added, modified, and deleted file (e.g. `git diff --name-status <last-deployed-commit>..HEAD`) so nothing is missed. Deleted files have to be removed on the server by hand.

### Frontend build-time variables

`VITE_GOOGLE_CLIENT_ID` and `VITE_TURNSTILE_SITE_KEY` are read by Vite when it starts: baked into the bundle at build time for Docker (the `Dockerfile` takes them as build args), or read at startup by the dev server on production (restart `lofp-frontend` after changing them). If they're missing, the Google sign-in button and the registration CAPTCHA don't render. `TURNSTILE_SECRET_KEY` must be set server-side at runtime, or `VerifyCaptcha` (`engine/internal/auth/turnstile.go`) silently skips verification.

**Never commit `.env`**: it contains secrets and is in `.gitignore`.

## MUD Client Protocols

Reference: https://wiki.mudlet.org/w/Manual:Supported_Protocols

Telnet server (`engine/internal/api/telnet.go`) implements these protocols:

### GMCP (option 201)
- **Core.Hello** — sent on connect: `{"client":"LoFP","version":<engine.GameVersion>}` (the same constant the in-game `VERSION` command reports)
- **Core.Supports.Set** — handled from client to track subscribed packages
- **Char.Vitals** — sent after every command: `bp`, `maxbp`, `mana`, `maxmana`, `psi`, `maxpsi`, `fatigue`, `maxfatigue`, `position`, `conditions`
- **Char.Status** — sent on login: `name`, `fullname`, `race`, `gender`, `level`, `experience`, `gold`, `silver`, `copper`
- **Char.Stats** — sent on login: `strength`, `agility`, `quickness`, `constitution`, `perception`, `willpower`, `empathy`
- **Room.Info** — sent on every room change (powers Mudlet automapper): `num` (int), `name` (string), `area` (string), `environment` (string), `exits` (map direction→room number)

### MXP (option 91)
- Line mode `\033[1z` = secure line (allows `<send>` tags until newline). NOT `\033[4z`.
- Used for clickable exits: `\033[1z<send href="north">north</send>, <send href="east">east</send>`
- `<send>` is a SECURE tag — only works in secure line mode (`\033[1z`)
- MXP output is gated on `tc.mxpEnabled` — plain clients see normal text

### MCCP2 (option 86)
- After negotiation, sends `IAC SB 86 IAC SE` uncompressed, then ALL subsequent output goes through zlib
- **IMPORTANT**: When MCCP2 is active, ALL data including IAC sequences (like WILL ECHO for password mode) MUST go through `t.write()` (the compressor), NOT `t.conn.Write()` (raw). Sending raw bytes corrupts the zlib stream and causes Mudlet to disconnect.

### MSSP (option 70)
- Sends game metadata (name, player count, website, genre, etc.) for MUD directory crawlers

### MSDP (option 69)
- Variable reporting for TinTin++ compatibility
- Handles REPORT subscriptions, pushes CHARACTER_NAME, HEALTH, MANA, ROOM, etc.

### Password Echo Suppression
- `WILL ECHO` / `WONT ECHO` toggle for password fields
- Must be sent through the compressor when MCCP2 is active
- `enterPasswordMode()` before prompt, `exitPasswordMode()` after reading

### NAWS (option 31)
- Window size negotiation, updates `t.width` for `wordWrap()`

## Script Language

The game world is defined in a custom scripting language (documented in `original/GMSCRIPT.DOC`):
- **Rooms**: `NUMBER`, `NAME`, `*DESCRIPTION_START/END`, `EXIT`, `ITEM`, terrain, lighting
- **Items**: `INUMBER`, `NAME` (noun ref), type, weight, volume, substance, worn slots
- **Monsters**: `MNUMBER`, body parts, stats, AI strategy, weapons, spells
- **Events**: `IFVERB/IFPREVERB/IFSAY/IFENTRY/IFVAR...ENDIF` conditional blocks
- **Variables**: Named variables + internal vars (stats, time, flags, item vals)
- Config file: `original/scripts/LEGENDS.CFG` lists all scripts to load in order

## Current State (v11.5)

- Script parser loads 2273+ rooms, 1990+ items, 297 monsters with case-insensitive file loading
- Full combat system: original [ToHit/Roll] format, weapon crits/slayers, fatigue, weapon clash
- 60+ spells across 5 schools, 30+ psionic disciplines
- Complete crafting: mining, smelting, forging, weaving, dyeing, foraging, alchemy (32 recipes)
- 36 skills with build point costs, prerequisites, and mechanical effects
- Treasure system: coin drops, weapon/armor/scroll/chest drops scaled by monster TREASURE level
- Monster AI: hostile aggro, flee behavior, special attacks, guard, demand-based spawning
- ELSE branches in script conditionals, all portal types supported
- 150+ emotes, race-specific, submit-gated interactions
- Character soft-delete with admin recovery, unique first names
- Security: rate limiting, connection caps, chat flood protection, HTML sanitization
- WebSocket-based real-time multiplayer
- Admin panel for rooms/items/monsters/players/logs
- Feedback pipeline: REPORT command forwards to VibeCtl for triage and issue tracking
- Production: single server at lofp.sanburnlab.com (see Environments and Deployment)
  - One instance to avoid multi-instance state desync (monsters, combat, player visibility)
  - Scale to 2+ machines only after implementing full hub-based monster/combat coordination

## Units

| Unit | Code | Path | Description |
|------|------|------|-------------|
| Engine | ENG | `engine` | Go backend: script parser, game engine, command interpreter, MongoDB persistence |
| Frontend | FRONT | `frontend` | React + Tailwind: player UI (text client) and admin interface |
| Scripts | SCR | `original` | Original game script files and documentation (read-only reference) |
