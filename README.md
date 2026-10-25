# Legends of Future Past

### **[Play at lofp.sanburnlab.com](https://lofp.sanburnlab.com)**

*Legends of Future Past* was a commercial text MUD that ran from 1992 to 1999 on CompuServe, dial-up and packet networks. The original engine source is lost, but the game's script files survived. In 2026 Jon Radoff rebuilt the engine from those scripts and the original documentation; this repository is a fork of [metavert-io/lofp](https://github.com/metavert-io/lofp) that runs at lofp.sanburnlab.com and continues development: restored spells, skills, crafting, guilds, GM tools and many fixes. See [CHANGELOG.md](CHANGELOG.md), or **Version Notes** in the game's main menu.

For the history of the original game and how it was brought back, see the [upstream README](https://github.com/metavert-io/lofp#readme) and [Resurrecting a 1992 MUD with Agentic AI](https://meditations.metavert.io/p/resurrecting-a-1992-mud-with-agentic).

## Connecting

- **Web client:** [lofp.sanburnlab.com](https://lofp.sanburnlab.com)
- **Bots:** see [bots/README.md](bots/README.md) for the WebSocket bot API

## Architecture

| Component | Technology | Path |
|-----------|-----------|------|
| Backend | Go + gorilla/mux + MongoDB | `engine/` |
| Frontend | React 19 + TypeScript + Vite + Tailwind 4 | `frontend/` |
| Scripts | Original 1992-1999 game data | `original/scripts/` |
| Documentation | GM Manual, player docs, session captures | `original/` |

## Running Your Own Server

Prerequisites: **Go** 1.25+, **Node.js** 22+, and **MongoDB**. Run MongoDB as a replica set (even a single node); the engine's messaging hub uses change streams and logs errors without one.

```bash
git clone https://github.com/csanburn1966/lofp.git
cd lofp
cp .env.example .env   # set MONGODB_URI and JWT_SECRET (at least 32 characters: openssl rand -hex 32)
./start.sh             # start.bat on Windows
```

The frontend runs at http://localhost:4992 and the backend at http://localhost:4993. Telnet is on port 4000 and SSH on 4022.

To run it in Docker instead, build the included `Dockerfile` (pass `VITE_GOOGLE_CLIENT_ID` and `VITE_TURNSTILE_SITE_KEY` as build args if you use Google sign-in or the registration CAPTCHA) and run it alongside MongoDB with `MONGODB_URI` and `JWT_SECRET` set. The container uses `config/prod.yaml` and serves the web client on port 8080.

## Credits

**Original Game (1992-1999)** — Copyright (c) 1992-1999 Inner Circle Software / NovaLink USA Corp

| Role | Name |
|------|------|
| Created & Programmed by | Jon Radoff |
| Additional Programming | Ichiro Lambe |
| Co-Producer | Angela Bull |
| Legends Manager | Gary Whitten |
| World Building | Gary Whitten, David Goodman, Tony Spataro, Stacy Jannis, Kevin Jepson, Daniel Brainerd, Michael Hjerppe |
| Documentation | Gary Whitten |
| Quality Assurance | David Goodman, Stacy Jannis |

**2026 Resurrection** — Reimplemented from original script files and documentation by [Jon Radoff](https://metavert.io) using [Claude Code](https://claude.ai/claude-code).

Special thanks to **David Goodman** for preserving the original game materials that made this reconstruction possible, and to **Chris Sanburn** for providing over 7MB of original session logs from the 1990s that helped refine the gameplay and restore authentic output.


**Fork (2026–)** — Continued development by Chris Sanburn.

## License

Released under the [MIT License](LICENSE) — Copyright (c) 2026 Metavert LLC.
