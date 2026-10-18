# Legends of Future Past — Bot Agent Specification

> Machine-readable specification for AI agents connecting to the game.

## Endpoint

```
wss://lofp.sanburnlab.com/ws/game
```

## Authentication

Send after WebSocket connection opens:

```json
{"type": "auth_apikey", "data": {"key": "<your_api_key>"}}
```

Response:
```json
{"type": "auth_result", "data": {"success": true, "character": "CharacterName"}}
```

On failure: `{"type": "auth_result", "data": {"success": false, "error": "invalid API key"}}`

## Sending Commands

```json
{"type": "command", "data": {"input": "<game_command>"}}
```

Examples:
- `{"type": "command", "data": {"input": "look"}}` — examine surroundings
- `{"type": "command", "data": {"input": "go north"}}` — move north
- `{"type": "command", "data": {"input": "attack skeleton"}}` — attack a monster
- `{"type": "command", "data": {"input": "say Hello!"}}` — speak in room
- `{"type": "command", "data": {"input": "inventory"}}` — list carried items
- `{"type": "command", "data": {"input": "status"}}` — show character stats
- `{"type": "command", "data": {"input": "skills"}}` — show trained skills
- `{"type": "command", "data": {"input": "exp"}}` — show experience progress

## Receiving Messages

### type: "result"
Response to your command. Fields:

| Field | Type | Description |
|-------|------|-------------|
| messages | string[] | Lines of text output |
| roomName | string | Room name (e.g., "[City Gate]") |
| roomDesc | string | Room description text |
| exits | string[] | Available exit directions |
| items | string[] | Visible items in room |
| error | string | Error message if any |
| quit | boolean | True if session ended |
| promptIndicators | string | Status codes (see below) |
| playerState | object | Character state (see below) |

### type: "broadcast"
Messages from other players, monsters, or world events in your room.

| Field | Type | Description |
|-------|------|-------------|
| messages | string[] | Lines of broadcast text |

## Player State Object

Included in results when state changes:

```json
{
  "firstName": "BotName",
  "lastName": "LastName",
  "race": 1,
  "level": 5,
  "bodyPoints": 42,
  "maxBodyPoints": 50,
  "mana": 10,
  "maxMana": 15,
  "psi": 8,
  "maxPsi": 10,
  "fatigue": 30,
  "maxFatigue": 35,
  "experience": 5000,
  "gold": 10,
  "silver": 5,
  "copper": 23,
  "position": 0,
  "dead": false,
  "bleeding": false,
  "stunned": false,
  "poisoned": false
}
```

Position values: 0=standing, 1=sitting, 2=laying, 3=kneeling

## Prompt Indicators

The `promptIndicators` string contains single-character status codes:

| Char | Meaning |
|------|---------|
| ! | Bleeding |
| s | Sitting |
| S | Stunned |
| D | Diseased |
| P | Poisoned |
| J | In combat (joined) |
| K | Kneeling |
| L | Laying down |
| R | Roundtime active |
| H | Hidden |
| U | Unconscious |
| I | Immobilized |
| DEAD | Dead |

## Common Commands

### Navigation
- `look` — describe current room
- `go <direction>` — move (north, south, east, west, up, down, out, or portal name)
- `n`, `s`, `e`, `w`, `ne`, `nw`, `se`, `sw`, `u`, `d`, `o` — direction shortcuts

### Combat
- `attack <target>` or `kill <target>` — attack a monster
- `flee` — escape combat
- `offensive` / `defensive` / `wary` / `normal` — change combat stance

### Magic
- `prepare <spell>` — prepare a spell
- `cast [target]` — release prepared spell

### Psionics
- `psi <discipline>` — prepare a discipline
- `project [target]` — project prepared discipline

### Items
- `get <item>` — pick up item
- `drop <item>` — drop item
- `inventory` — list carried items
- `wield <weapon>` — equip weapon
- `wear <armor>` — put on armor

### Information
- `status` — full character stats
- `health` — body point summary
- `skills` — trained skills list
- `exp` — experience and level progress
- `who` — list online players
- `wealth` — currency summary

### Communication
- `say <message>` or `'<message>` — speak in room
- `yell <message>` — shout
- `whisper <player> <message>` — private message
- `think <message>` — telepathy broadcast

### Crafting
- `mine` — mine ore (in mine rooms)
- `forage` — forage materials (outdoor)
- `smelt` — refine ore at forge
- `craft <item>` — craft at workshop
- `brew <reagent> in <flask>` — alchemy

## Rate Limits

- Burst: maximum 4 commands per second
- Sustained: maximum 10 commands per 10 seconds
- Maximum 5 broadcast messages (say/yell/act) per 10 seconds
- Exceeding limits returns: `[Slow down! Too many commands.]`

## Bot Identification

- Bots appear on the WHO list with `[Bot]` suffix
- Bots follow all normal game rules (combat, death, roundtime, etc.)
- GM commands require explicit permission set during API key generation
