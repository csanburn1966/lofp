# New Havarth Mining Company: "The Surveyor's Commission" — GM Runbook

Four surveying assignments for the New Havarth Mining Company, upstairs from the ore counter in the North Bazaar (room 394). Each assignment unlocks when the previous report is paid. Everything is scripted. No GM is needed, and the Havarths are room messages, not NPCs.

Design source: `New Havarth Quest suggestion.docx`.

## Files

| File | Purpose |
|------|---------|
| `scripts/HAVARTHITEMS.SCR` | The surveyor's instrument and the Havarth Masterwork Mining Axe, plus their words. **Permanent** once live: add it to `LEGENDS.CFG` so items players own keep working. |
| `scripts/HAVARTH.SCR` | Everything else: room 394 (staircase, Bram), the three office rooms, the Shartan prospect rooms, the dialogue and survey macros, ambient lines, and the newsboard poster. |

Engine changes these rely on (deploy these files): `engine/internal/engine/crafting.go`, `engine.go`, `scripts.go`.

- `SURVEY` lets a carried item with an `IFPREVERB SURVEY -1` script handle the command first, for anyone, before the Mining skill check. Without the instrument, SURVEY works as before.
- `INCOMBAT` script variable: 1 while a living monster in the room is targeting the player, or the player is targeting one.
- `MINE WITH <tool>` (or `MINE <tool>`) picks which carried tool to use. Plain `MINE` uses the first mining tool in INVENTORY order. A `MINETOOL` item's `PARAMETER3` adds that many ranks of Mining while it's the tool in use (the axe: +2 ranks, which is +10% to find ore and +1 purity).

## Testing on lofp-dev (GM script area)

1. Deploy the engine changes.
2. GM panel → GM Scripts: upload `HAVARTHITEMS.SCR` with a **higher priority** than `HAVARTH.SCR` (the items and words must load first). Save both as drafts, then **Activate** `HAVARTHITEMS.SCR` and then `HAVARTH.SCR`.
3. **Disable** `HAVARTH.SCR` to switch the quest off. Room 394 reverts, the office and prospect rooms disappear, and SURVEY with the instrument falls back to the normal mining survey. Players standing in a removed room should be moved first.

## Going live

Add both files to `LEGENDS.CFG` (after `FESTITEMS.SCR`), restart, then delete the GM script copies. At minimum `HAVARTHITEMS.SCR` must be permanent. Don't leave `HAVARTH.SCR` both in `LEGENDS.CFG` and as a live GM script.

## Walkthrough

Talk to the Havarths with `SAY`. The keywords are matched anywhere in what's said.

| Where | Who | Keywords |
|-------|-----|----------|
| 394 ore counter (`GO DOOR` from Bazaar North 207) | Bram (pays out, points upstairs) | work, job, survey, Edric |
| 9400 main office (`UP` from 394) | Edric (assignments, accepts reports, pays) | work, job, assignment, survey, report, ready, complete, finished |
| 9401 surveying room (`WEST` of 9400) | Mara (compiles reports, directions, replacement instruments) | report, survey, readings, complete, finished, done · where, directions, help · instrument, lost, replace |
| 9402 workshop (`EAST` of 9400) | Tessa (flavor; made the axe) | axe, pick, tool, work |

Flow for each assignment: Edric gives the work → `SURVEY` at each spot (5 s round time, then "A geological reading has been successfully recorded.") → tell Mara the readings are complete → tell Edric about your report → paid. `LOOK INSTRUMENT` shows the survey journal.

| # | Assignment | Survey spots (room) | Reward |
|---|-----------|---------------------|--------|
| 1 | Inner Sea caves (in through the opening at 529) | cave entrance 978 · dripping tunnel 981 · springs confluence 953 (down the hole at 982) | 100 gold crowns, 1,000 XP |
| 2 | Abandoned Dig (3400, ride the platform) | upper shaft 3402 · middle 3403 · lower 3404 · bottom 3405 | 250 gold crowns, 2,000 XP |
| 3 | Matriarch Tree (Grymwood) | foot of the tree 110 · basement 109 · cranny 103 | 500 gold crowns, 4,000 XP |
| 4 | Teeth of Shartan | map 1: ledge path 627, Teeth of Shartan 636 · map 2: landslide 712, dam base 732 · map 3: the prospect 9413 | 1,000 gold crowns, 8,000 XP, masterwork mining axe |

Assignment 4's prospect has no entrance from the mountain. Once the first four Shartan readings are taken, `SURVEY` on the Lakeside Road at room **740** points at a cleft behind the boulders and takes the surveyor into 9410. Before that, SURVEY there says it needs more readings. From 9410: N 9411 (narrow passage), NE 9412 (banded gallery), D 9413 (deep chamber, MINEA, the final reading). 9411 and 9412 spawn Lakeside monster group 26. `OUT` from 9410 returns to 740. Players who've finished can always get back in the same way, so the prospect stays a grade A mining spot for them.

Rules:
- One assignment at a time. Readings, the report and the stage are stored on the character and survive logout, disconnect and death.
- Each spot records once per assignment. A second SURVEY there says it's already recorded. Anywhere else, the needle wanders (2 s round time).
- The instrument refuses during combat.
- A report can't be paid twice. The axe is given once, and the XP uses one-time `AWARD` ids, so it can't be claimed twice even if a GM resets the stage.
- Lost instrument: Mara sells a replacement for 5 gold crowns, at any point from assignment 1 until all four are done. Progress isn't lost.

## GM aids

Per-character variables (`@peek`/`@setp <player> INTNUMnnnn <value>`):

| INTNUM | Meaning |
|--------|---------|
| 1040 | stage: 0 none · 1, 3, 5, 7 = assignment 1–4 under way · 2, 4, 6 = assignment 1–3 paid · 8 = all done, axe given |
| 1041–1045 | readings taken this assignment, one per spot in the order above (1 = taken) |
| 1046 | 1 = Mara has compiled the report; Edric pays on the next mention |
| 1049 | scratch (last reading finished the assignment) |

- **Skip ahead:** set 1040 to the even number before the assignment you want (e.g. `4` to offer the Matriarch next), and clear 1041–1046.
- **Re-run a finished assignment:** set 1040 back to its odd number and clear 1041–1046. XP isn't paid twice (one-time `AWARD` 1040–1043), but coins are.
- **Lost axe:** `@additem 9401 adj1=9401 adj2=9402`, then give it to the player. **Instrument:** `@additem 9400 adj1=32 adj2=9400`.
- **Moving a spot:** each spot is an `IFVAR RNUM = <room>` line in MACRO 9400 of `HAVARTH.SCR`. Update the matching journal line in MACRO 9404 and Mara's directions in MACRO 9406 too.

## Reserved IDs (94xx block)

| Type | IDs |
|------|-----|
| NOUNDEF | 9400 (cleft) |
| ADJDEF | 9400 surveyor's · 9401 masterwork · 9402 mining |
| INUMBER | 9400 instrument · 9401 axe |
| NUMBER (rooms) | 9400–9402 offices · 9410–9413 Shartan prospect |
| MACRO | 9400–9409 |
| CEVENT | 9400 (a Havarth line every ~2 min in an occupied company room) |
| AWARD | 1040–1043 |
| INTNUM | 1040–1049 |

Overrides: room 394 (copied from `FAYDINDR.SCR`) and newsboard item 304 (copied from `NEWS.SCR`, the poster is page 8 in Fayd). If either original changes, update the copy in `HAVARTH.SCR`.
