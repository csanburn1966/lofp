# Fall Festival & "The Last Lantern of Fayd" — GM Runbook

Three script files:

| File | Purpose | Reuse |
|------|---------|-------|
| `scripts/FESTITEMS.SCR` | **Permanent** (in `LEGENDS.CFG`): every keepsake players can buy or earn — festival goods, clue items, tokens, necklace, charm, ring — plus their vocabulary, so they still work after the event files are removed | Always loaded |
| `scripts/FALLFEST.SCR` | Hot-loaded. Festival stock in shops, the festival lot at the cemetery gate, treasure balloons, crystal-ball fortunes, ambient echoes, 9 Halloween monsters + opt-in spawner | Every Fayd festival |
| `scripts/HALLOWEEN2026.SCR` | Hot-loaded. Al's story event: Marren Vey's ghost, the scavenger hunt, the 12 funeral tokens, the chest and rewards | 2026 only (copy + rename for future stories) |

Story source: `original/halloween.txt`.

## Before the event

1. **Deploy the engine, the GM panel and the permanent items.** The draft support needs `engine/internal/api/api.go`, `engine/internal/engine/engine.go`, `gm_scripts.go`, `monsters.go`, `movement.go` and `frontend/src/components/GMPanel.tsx`. `@season` and `@reboot` need `gametime.go`, `gm_commands.go`, `movement.go` and the new `reboot.go`. Existing GM scripts stay live after the upgrade. Copy `scripts/FESTITEMS.SCR` and the updated `scripts/LEGENDS.CFG` to production, then restart once — this only adds item definitions, so nothing changes for players. The scripts also rely on engine changes made alongside them (deploy these files): `engine/internal/engine/containers.go`, `engine.go`, `gm_scripts.go`, `inventory_commands.go`, `look.go`, `scripts.go`. What they add:
   - `PUT <item> ON <target>` works (previously only `IN`).
   - New verbs `POP`, `RIP`, `LIFT`, `PRY` that reach item scripts (`pry slab with prybar` works).
   - `UNLOCK` runs room-item `IFPREVERB UNLOCK` scripts.
   - A LOOK script that uses `CLEARVERB` replaces the item's default description instead of being printed after it.
   - Re-uploading a GM script replaces its CEVENTs instead of doubling them.
   - `NEWITEM -1` rewards save immediately.
   - `REMOVEITEM -1` removes the exact item (a popped treasure balloon can't consume a plain one).
2. **Hold the season.** A season change used to replace the cemetery and lot room descriptions mid-event. Rooms owned by a live GM script are now left alone, but the rest of Fayd would still change season under the festival. Before the festival opens, type **`@season`** to see the current season and when it next changes. Then either:
   - type **`@season set autumn`** to switch to autumn and hold it, or
   - type **`@season hold`** to freeze whatever season it is now.

   The hold survives restarts. Run **`@season release`** after the event.
3. **Upload both as drafts for review.** In the GM panel → GM Scripts, upload each file and click **Save as Draft**. A draft is stored and readable by every GM, but it isn't applied to the game and isn't loaded at restart; it shows a grey **DRAFT** badge. Give `FALLFEST.SCR` a higher priority than `HALLOWEEN2026.SCR`. FALLFEST has to apply first, because the story file redefines room 287 and needs FALLFEST's vocabulary.
4. **Go live when the festival opens.** Select `FALLFEST.SCR` and click **Activate**, then do the same for `HALLOWEEN2026.SCR`. Each turns into a green **LIVE** badge. Shops change immediately; the story stays dormant until armed.
5. **If something is wrong during the event,** click **Disable** on the script. It goes back to being a draft, and its rooms, items, monsters, macros and timed events are taken out of the running game immediately, with no restart:
   - Anything it replaced reverts to the original, or to the other live script's version. For example, disabling only the story file puts room 287 back to FALLFEST's festival lot.
   - Creatures that only it defined are despawned.
   - Edit the draft, **Save as Draft**, and **Activate** again when it's fixed.
   - **Save & Apply** on a live script applies edits directly, and removes anything the new version no longer defines.
   - Disabling or re-activating replaces the affected rooms, so anything players dropped in those rooms is lost.

## Festival (FALLFEST.SCR)

| Where | What |
|-------|------|
| 287 Poor Quarter, festival lot (outside the cemetery gate) | Campfire, free `GET BREAD` / `GET CIDER`, balloon cart (`READ CART`, `BUY ORANGE BALLOON`, `BUY TREASURE BALLOON`), crystal ball (`TOUCH BALL` → fortune, also sent to GMs) |
| 359 Tailor's Shop | Festival shirts (orange/black/purple/white, 3 crowns), festival masks (pumpkin/skull/ghostly/black, 2 crowns) |
| 199 Randolphs | Spider rings (black/orange/purple, 5 shillings), spiderweb earcuffs (silver 5 cr, gold 12 cr). `READ SIGN` |
| 398 Aunt Emma's Bakery | Sugar skulls, caramel apples, pumpkin pie, pumpkin bread, spiced cider |

Treasure balloons: `POP`, `POKE` or `RIP` gives a random prize: black plastic spider ring, glowing enamelled skull pin, or silver spiderweb earcuff. A GM can also make any balloon a treasure balloon by setting its VAL5 > 0.

The shirt's tagline is in `FALLFEST.SCR` (INUMBER 9350). Edit it each festival year if you like.

**Switches:**

| Command | Effect |
|---------|--------|
| `@set HWQUIET 1` | Silence the festival / cemetery ambient echoes (0 = on, the default) |
| `@set HWHAUNT 1` | Halloween monsters spawn in the cemetery (≈ every 2 min, max one per room). **Keep 0 during the story.** |

**Monsters** (`@genmon <#>`):

| # | Monster | Notes |
|---|---------|-------|
| 9350 | cackling witch | caster: flame bolt, fear, slow |
| 9351 | wailing ghost | needs a magic weapon, cold touch |
| 9352 | stitched monster | boss: big, slow, immune to electricity, burns |
| 9353 | gill monster | place near water; skinnable scales/claws |
| 9354 | headless horseman | boss: scythe, no head |
| 9355 | grinning pumpkinhead | weak, burns |
| 9356 | black cat | harmless, flees |
| 9357 | pale vampire | fatigue drain, charm/slow |
| 9358 | shrieking banshee | needs a magic weapon, scream |
| 9300 | restless spirit | story only (non-hostile, unkillable) |

## The Last Lantern of Fayd (HALLOWEEN2026.SCR)

The cemetery's regular monster spawning is switched off in rooms 299–310 while this file is loaded. `@zap` anything already wandering there before you start. Soulcatcher must be **visible** (not `@invis`/`@hide`) whenever their words or actions trigger a scene, or the echoes are suppressed.

| Step | GM does | Players do |
|------|---------|------------|
| Stories (0–15 min) | Gather at the fire (room 287). When ready: **`@set HWLANTERN 1`** | Tell ghost stories |
| Manifestation | Soulcatcher says: *"Marren Vey, your vigil is done."* The script runs about 25 seconds of scene: blue fire, shadows, ghost, his words, the fire dies. An **ancient obelisk token** lands in the ashes | `GET TOKEN`, `CLIMB GATE` |
| Clue 1 — room 300 | | `SEARCH OBELISK`, `SEARCH SKULL` → copper strip, `READ STRIP` |
| Clue 2 — room 309 | | `SEARCH ROOTS` (or tree) → tarnished brass key; Marren appears briefly |
| Tools — room 307 | | `SEARCH` outside the house → old shovel + rusty prybar |
| Clue 3 — room 308 (`GO WALL`) | | `SEARCH FLOOR` / `SEARCH BED` (needs the key) → blackened parchment, `READ PARCHMENT` |
| Cache — room 310 | | `SEARCH MAUSOLEUM` → slab; `PRY SLAB` / `LIFT SLAB` (needs the prybar or any shovel) → chest + bag of 12 tokens |
| | Forcing the chest (`KICK`/`BREAK`/`PRY CHEST`) raises 3 restless spirits that surround the party | `LOOK IN BAG`, `GET <symbol> TOKEN FROM BAG` |
| Marren speaks | Anyone saying "Marren" in room 310 brings him out (about 15 seconds) | |
| Tokens | Hold back the **lantern** token (Elira Vey); it only works after the other 11 | `SEARCH GRAVESTONES`, then `PUT SUNRISE TOKEN ON SUNRISE GRAVESTONE`, etc. |
| Finale | `PUT LANTERN TOKEN ON LANTERN GRAVESTONE`. About 45 seconds of scene; the chest opens and the spirits fade | |
| Rewards | | `SEARCH CHEST`, once per character |
| Closing (55–60 min) | Return to the lot; the fire rekindles when the first person walks in. Soulcatcher's closing lines are delivered live | |

Each player's share is: 50 crowns, an antique silver necklace, and Marren Vey's lantern charm (`RUB CHARM` → Light, 50 charges, rechargeable with Charge Wand).

**Symbols:** sunrise, teardrop, wheatsheaf, hammer, birdcage, crescent, star, hound, quill, leafless, handprint, lantern. `LOOK <symbol> GRAVESTONE` shows the name and epitaph.

**GM aids**

- `[Last Lantern] …` GM messages announce each stage and the **first finder** of each leg (strip, key, parchment, slab). Give those players a funeral ring: `@additem 9346 adj1=115` (an engraved funeral ring), then `@give ring to <player>`.
- Exceptional ghost stories: hand out gems with `@additem` (find gem archetypes with `@find`).
- Necklace enchantment: Enchant an Item (244) isn't implemented yet, so a GM imprints a spell with `@editem` — VAL3 = spell #, VAL2 = charges. After that, `RUB` while worn casts it, exactly like other rub-to-cast jewelry. You can copy the values from the player's old necklace.
- **Hint:** `@set HWHINT 1` makes Marren's lantern drift toward the next location once (seen from every cemetery room within about 30 seconds).
- **Story stage:** `@peek HWLANTERN`. To skip or repair, `@set HWLANTERN <n>`:

| HWLANTERN | Meaning |
|-----------|---------|
| 0 | dormant |
| 1 | armed |
| 2 | manifested — clue 1 live |
| 3 | strip found — clue 2 live |
| 4 | key found — house live |
| 5 | parchment found — mausoleum live |
| 6 | chest found |
| 7 | name spoken — tokens |
| 8 | chest open |

  Other variables: `HWTOKENS` (0–11), `HWSPIRITS`, `HWFIRE`.
- Per character, `INTNUM1031 = 2026` marks a claimed share. `@setp <player> INTNUM1031 0` lets them claim again.
- Don't `@dupe` funeral tokens; extra copies would count toward the 11.

## After the event

1. `@set HWQUIET 1` and `@set HWHAUNT 0`, then `@zap` any leftover Halloween monsters. `@season release` to let the calendar run again.
2. GM panel → GM Scripts: click **Disable** on `HALLOWEEN2026.SCR`, then on `FALLFEST.SCR`. The game returns to normal immediately, without a restart. Keep them as drafts to reuse `FALLFEST.SCR` next Fayd year, or delete them; deleting a live script also removes it from the game.
3. A few harmless leftovers stay in memory until the next restart: the named variables (`HW…`) and the Halloween monster adjectives. A GM can restart the server with `@reboot 300`, which gives a 5-minute warning; `@reboot cancel` stops it. Don't re-upload `FAYDINDR.SCR` / `FAYD*.SCR` as GM scripts to "undo" anything: they'd stay in MongoDB and override the seasonal files on every later restart.

Everything players bought or earned keeps working afterwards, because those definitions live in the permanent `FESTITEMS.SCR`. Leave that file in `LEGENDS.CFG`.

## Reserved IDs

| Type | FALLFEST | HALLOWEEN2026 |
|------|----------|---------------|
| NOUNDEF | 9350–9351 *(in FESTITEMS)* | 9300–9303 *(in FESTITEMS)* |
| ADJDEF | 9350–9356 *(in FESTITEMS)* | 9300–9311 *(in FESTITEMS)* |
| MADJDEF | 9350–9357 | 9300 |
| INUMBER — keepsakes, in FESTITEMS | 9350–9357 | 9302–9306, 9320–9331, 9344–9346 |
| INUMBER — fixtures/scenery, in the event file | 9360–9365, plus override of 739 (balloon) | 9300–9301, 9307–9314, 9332–9343 |
| MNUMBER | 9350–9358 | 9300 |
| MACRO | 9350–9353 | 9300–9306 |
| CEVENT | 9350–9352 | 9300 |

Rooms overridden: FALLFEST 287, 199, 359, 398; HALLOWEEN2026 287, 299, 300, 301, 302, 307, 308, 309, 310. The 93xx block is reserved for Halloween / Fall Festival content.
