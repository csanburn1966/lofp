# Changelog

## v15.4.0 — 2026-10-24

Your spells no longer wear off while you’re logged out. Log off with 40 minutes of Haste left and you’ll still have 40 minutes when you return. The weather makes more sense too: no more heavy snow at 39 degrees in spring, and it rains far less often. Miners can `SURVEY` an area to learn its ore grade, and Yshtarin’s Confounding Translocation can now send other players to your mark.

### Changed: Spells Pause While You’re Offline
- Every timed spell on you stops counting down when you log off and picks up where it left off when you log back in. This includes Mindlink, Haste, Fly, Mystic Armor, the strength and agility spells, Heat and Cold Shield, Camouflage, Light, Night Vision, Divine Blessing, Resist Weather, Claw Growth, the Repel Plants spells, and Wizard’s Armor and the other defense spells
- Spells on your own items pause too, including items in containers and in the bank: Storm, Inferno and Winter Blade, and an hour of Tempering. Items left on the ground keep counting down
- Harmful effects pause as well. Logging off won’t wait out a Silence, Imprison, Slow, Paranoia or an entangle
- Spells keep their time through a server restart

### Changed: Weather and Temperature
- Each region now has its own temperatures for each season, taken from the original world files. They vary from day to day and are cooler at night
- Snow falls only when it’s at or below freezing, and rain only when it’s above it. Snow can no longer come in summer or a warm spring
- Some regions are wetter or drier than others, as the original world files set out
- Storms clear up more readily. It used to rain or snow about two thirds of the time. It’s now closer to a quarter in an average region

### Fixed: SURVEY
- With the Mining skill, `SURVEY` tells you whether the area can be mined, and if so, whether its ore is grade A, B or C. Each survey takes 2 seconds
- Places with special survey messages in their scripts still show them

### Fixed: Targeting Players and Translocation
- `TARGET` works on players in the room again, as it did in the original game. Players and creatures share the limit of 9 targets, and `TARGET` on its own lists both
- Yshtarin’s Confounding Translocation (246) sends up to four of your targets, creatures or players, to your marked location. Creatures go first
- Chain Lightning, Flaming Arrows and Siryx’s Terrible Tentacles never harm players, since there is no player-versus-player combat. You’ll see a notice that the spell passed them by

### Fixed: Thesnia Leaf
- Eating or drinking more thesnia while its telepathy is still running now adds 20 minutes, up to 4 hours, as recasting Mindlink does. It tells you how many minutes are left

## v15.3.0 — 2026-10-23

Three commands that only said “coming soon” now work: `BREAK` smashes things open, `STALK` lets the stealthy shadow someone unseen, and `SELFTRAIN` lets you push a skill past what any teacher offers. Treasure chests are dangerous again. Their traps go off whether the chest is on the floor or in your hands, and `DISARM` finds and removes them. Picking locks now looks and plays as it did in the original game, and `UNLEARN` returns every build point you spent.

### Added: BREAK \<object\> WITH \<object\>
- Smash an item with something made of stronger stuff (synonym `DESTROY`). Wood breaks against stone, stone against metal, and so on. Nothing breaks against a weaker material
- It can take many tries. A hammer or an axe helps, and so does a strong arm. Each attempt takes 5 seconds, halved when you’re hasted
- A broken container spills what’s inside onto the floor. Potions, vials, flasks and bottles inside are destroyed, and each gem has an even chance of coming out chipped or cracked if it wasn’t already
- Breaking a trapped chest sets off its trap

### Added: STALK
- `STALK <player or creature>` follows them like `FOLLOW`, but quietly. You don’t join their group, and no one sees you come and go
- Needs the Stealth skill. Each time your quarry moves, you make a Stealth roll against their Perception. Fail it and you lose track of them
- `STALK` on its own, `LEAVE`, `FOLLOW` or `HOLD` ends it, and so does walking off on your own

### Added: SELFTRAIN
- `SELFTRAIN <skill>` raises a skill you already know without a teacher or a fee. It costs 1 build point more than `TRAIN` would
- Your build point totals keep track of which ranks you trained yourself, so the extra point is always counted correctly

### Fixed: Trapped Chests
- Chests and other containers from treasure have always come with traps, but a trap only worked while the chest sat on the floor. Since most adventurers pick a chest up first, traps never went off. Now they go off wherever the chest is, when it is opened, its lock is picked or it is broken
- An acid trap eats the lockpick working on it, and an explosive trap destroys the chest and everything inside, as in the original game
- `DISARM <object>` works as the original manual describes. With Trap & Poison Lore, the first try finds the trap and the next tries remove it. Your skill rank and Perception improve your chances, and harder traps lower them. Each attempt takes 5 seconds, and a very bad fumble sets the trap off
- It also works on chests you are carrying (`DISARM MY CHEST`)

### Changed: PICK and UNLEARN
- `PICK` now shows your chance and roll, with the original game’s messages. Each attempt takes 5 seconds, halved by the Haste spell or the psionic Speed ability. Some locks really are beyond you now: your chance can drop to 0%
- `PICK CHEST WITH LOCK` and other shortened lockpick names now work
- `UNLEARN` returns every build point you spent on that rank, including the extra point for a self-trained one. As the manual says, it can be used once per level. It also accepts a skill number
- Unlearning a rank of Endurance now takes back the 4 body points it gave you

### GM & Scripting Notes
- When `@edsk` lowers a skill, self-trained ranks come off first and their extra build point is returned. It now reports the change to the player’s build points, and it won’t set a skill below 0
- A trap in an item’s `VAL4` now works on any container that opens or any door, wherever it is, and `DISARM` can find and remove it. Potion containers and vessels in the middle of a `BREW` keep their own `VAL4` meanings and are never treated as trapped

## v15.2.0 — 2026-10-21

Enchanters gain a new level 28 spell, Tempering, which hardens a weapon or shield against breaking, for an hour or, with a rare polished diamond, for good. Aelfen can finally `CALL` a woodland creature to serve them. Mass Invisibility now hides your whole group, Mindlink can be extended by casting it again, and CyberTech now serves humans only. GMs can work through player reports from the GM page, and look up adjective and noun numbers there.

### Added: Tempering (223)
- A new level 28 Enchantment spell. `PREPARE 223`, then `CAST <weapon or shield>`, makes the item much harder to break in a weapon clash. It works on anything you wield, hold, wear or carry
- Prepared on its own, the tempering lasts one hour. Casting it again on the same item resets the hour; the time doesn’t add up
- Prepared with a large or huge diamond that a Highlander has polished with `MOLD` (`PREPARE 223 WITH DIAMOND`), the tempering is permanent. The diamond is used up only if the spell succeeds
- Tempering never stacks. An item that is permanently tempered can’t be tempered again

### Added: Aelfen CALL
- Aelfen can now `CALL` a woodland creature to be their familiar, as described in the original player manual. It answers the same way the druid spell Call Animal does, with the same animals
- It only works in the wilderness (forest, mountains, plains, swamp, jungle or wasteland), and you can have only one called creature at a time

### Changed: Spells and Cyberware
- Mass Invisibility (212) now turns every member of your group in the room invisible along with you. Any group member can cast it, not just the leader
- Casting Mindlink on someone who already has it adds 20 minutes, up to the usual maximum, instead of doing nothing. Eating or drinking thesnia no longer cuts a longer telepathy short
- CyberTech’s devices were built for humans, and the surgery now turns other races away before taking any money
- Summoned creatures’ messages now give the right command for watching a room: `COMMAND <creature> WATCH`

### Changed: World
- All three Fayd restaurants that list grilled shrimp on their menu now sell it, in every season
- The Crafters Guild’s weavers’ cabinet, cooler and rack no longer turn away dyes and materials you found by searching or by foraging in the Caldis Forest as “already purchased”

### GM & Scripting Notes
- The GM page has a new **Reports** tab listing players’ `REPORT`s, with a count of unresolved ones on the tab. GMs can mark reports done, reopen them and reply, as on the Admin page. Replies are credited to the GM character
- Number Ranges now covers **Adjectives**, **Monster Adjectives** and **Nouns** (`ADJDEF`, `MADJDEF`, `NOUNDEF`), with free gaps, reservations and a searchable list of every number, its word and the script that defines it
- `ITEMWORTH` now gives the same full value SELL starts from, so items found by SEARCH or made by `NEWITEM` have a worth too. A negative crafted value (`CRAFTEDVALUE=-1` on `NEWITEM`/`NEWPUT`) marks an item as having no resale value, and its `ITEMWORTH` reads 0

## v15.1.0 — 2026-10-19

Guarding is reworked: `GUARD` no longer switches itself off, the new `UNGUARD` command stops it, and a guarding elemental now steps in front of monsters and fights them for you. The way into the sewers beneath Fayd is open again, scripted exits written with short directions (such as `D` or `NE`) now work across the world, and the Night Shades’ hideout is part of the game. GMs can export a character from the live server and import it into their own.

### Changed: Guarding
- `GUARD <target>` only starts guarding. Using it on someone or something you already guard tells you so, instead of quietly stopping
- New `UNGUARD <target>` stops guarding a person, doorway or item. `UNGUARD` on its own stops guarding everything, and `GUARD` on its own lists what you’re guarding
- Summoned creatures work the same way: `COMMAND <creature> GUARD ME` starts, and `COMMAND <creature> UNGUARD ME` (or `UNGUARD <name>`) stops
- A guarding elemental now blocks monsters that come for you: “The ogre advances towards you, but the earth elemental blocks its way!” The monster fights the elemental, not you, so you’re free to move, cast or retreat
- A monster already fighting you turns to face your elemental when it steps in, and if the elemental falls, the shock still stuns you

### Changed: World
- The Mullock Heaps are back on the map, southeast of Broadway. Sit by the chute between the bins and slide down into the sewers beneath Fayd
- Scripted exits written with a short direction now work, as they did in the original game. Among them: the knife gang in the alleys, the Thief’s shop back room, the crawlway, the mountain trail, the PeeWee ladder and some deadly traps. Explore carefully
- The Night Shades’ hideout is now part of the world. Knocking on the hidden panel in their safe house leads the way in
- Robotic cleaners no longer wander out of the mall into the Caldis Forest and Blackthorn

### Changed: Combat, Crafting and Cyberware
- Bare-handed punches now do crushing damage and no longer cause bleeding. A Wolfling in wolf form still claws
- The loom refuses mismatched materials: leather armor needs a hide (leather, suede, doeskin, fur and so on), and cloth armor needs a woven fabric. Ordinary garments still take any material
- Any cyberware implant now stops you casting spells, and Coriam Seeds can no longer restore Empathy while an implant is installed
- A Wolfling in wolf form who `HOLD`s someone takes them by the scruff of the neck, not by the hand
- Skinning messages show a disguised player’s disguise, not their name

### GM & Scripting Notes
- Admin → Players has an **Export** button that downloads the selected character as a `.lofp-player.json` file, and **Import Character…** loads one into an account you choose, optionally under a new name. Imported characters get a fresh ID, and anyone standing in a room that doesn’t exist on your server lands in room 201. Both actions are logged
- @GM channel lines are shown in bold magenta in the web client, telnet and SSH, and read “@GM Name SAYS: …”
- `@cyberware <player> remove all` removes every implant and restores the player’s spellcasting
- Room scripts match directions in either form: `IFPREVERB D -1` and `IFPREVERB DOWN -1` now behave the same
- `NSHADES.SCR` is now loaded from `LEGENDS.CFG`, ahead of `FAYDINDR.SCR`
- The mall’s spawn groups 35 and 36 were renumbered to 124 and 125. They had been sharing group numbers with the Caldis Forest and Blackthorn

## v15.0.0 — 2026-10-18

A major update: CyberTech is open for business. Implants bought in Fayd are now grafted onto your body as they were in the original game, trading Empathy for strength, speed, sharper senses, telepathy or full-body plating. The calendar gains the seven Feast Days of the Feast of the Great Moon that open each year, and a new `DATECVT` command converts between real and in-game dates. Fighting alongside your group now helps you hit, with Wolflings fighting as a pack. Sorcerous Summons finally calls its demons, and a number of bugs are fixed, including a shop that disconnected players.

### Added: Cyberware
- Buying an implant at CyberTech now installs it. You’re sedated and wake on the operating table in the surgery, sore and woozy. `INFO` lists your implants under “Cyberware:”
- Bionic Arms raise Strength, Bionic Legs raise Quickness, and Reflexors raise Agility and Quickness. Each comes in Alpha, Beta, Gamma and Omega grades, and an Omega Arm alone adds 24 Strength
- Bionic Eyes raise Perception, and from Beta up they let you see in the dark. Bionic Ears sharpen your hearing, making it easier to find someone hiding and to notice someone sneaking in
- The Telepathor gives you permanent telepathy. `TELEPATHY` turns it on or off, and `THINK` sends your thoughts. It doesn’t grant `CONTACT`
- Cyberplate Armor protects your body, arms, back and legs: 70% at Alpha, up to 85% at Omega. It weighs nothing and can’t be lost or damaged
- When Cyberplate is installed, any body armor, shoulder or thigh guards and bracers you’re wearing go into your pack, and you can’t wear them again. Helms, gloves, gauntlets and boots still fit, and so do cloaks
- Every implant drains Empathy, and your mana shrinks with it. Heavy cyberware can end your spellcasting altogether
- Buying a higher grade replaces the implant you have. You can’t buy the same grade twice or a lower one. Ephemerals can’t receive implants
- Implants bought before this update, which sat uselessly in your pack, are installed automatically the next time you log in

### Added: Feast of the Great Moon
- Each year now opens with the seven Feast Days of the Feast of the Great Moon, the New Year’s celebration, before Abra. The year is 343 days long, as in the original lore
- `TIME` reads, for example, “It is the third day of the Feast of the Great Moon, 1047”. The Feast is not a month, and it counts as winter
- Today’s date and time stay exactly where they were. Only future years gain the Feast Days

### Added: DATECVT
- `DATECVT 10-25-2026` tells you which in-game days a real date covers. Add a time, such as `DATECVT 10-25-2026 3:30pm`, for the exact in-game moment
- `DATECVT GAME Manta 14 1047` (or `GAME 3-14-1047`, or `GAME FEAST 3 1047`) gives the real date and time span. `DATECVT NOW` and `DATECVT TODAY` also work
- Replies give the in-game day of the week and use the server’s time zone, which is named in the reply. Dates far from today are estimates, because the game clock pauses while the server is down

### Added: Group Fighting Bonus
- Fighting with your group gives +2% to hit for every group member in the room, counting yourself. It applies to weapons, claws, bites and backstabs
- Wolflings fight as a pack: each Wolfling in the group is worth +5% to a Wolfling. In a group of 3 Wolflings and 2 Humans, the Humans get +10% and the Wolflings +19%
- Your attack modifier in `INFO` includes the bonus

### Added: Sorcerous Summons
- Sorcerous Summons I (139) tears open the air and calls a lesser demon. Sorcerous Summons II (140) calls a greater demon. Neither demon can be controlled, and each attacks everyone, including whoever summoned it
- A badly fumbled cast simply fails instead of backfiring
- Creatures that know these spells, such as the vith drone and the winged panther, now really summon a demon, and won’t summon another while theirs is still alive

### Changed: Hiding, Surfaces and Repairs
- Darkness and night now help you `HIDE`, not just make you harder to find: +20% in an unlit dark room, and +10% outdoors between 7pm and 5am
- `GET` now picks items up off tables, workbenches and other surfaces in the room when there’s no match on the floor. `GET 2 LOCKPICK` counts floor items first, then those on surfaces
- `LOOK ON` a surface lists what’s on it, and `LOOK IN` a surface shows the same. Surfaces no longer show as “(open)” or “(closed)”
- `REPAIR` accepts numbers and ordinals: `REPAIR 4 BROADSWORD` or `REPAIR FOURTH BROADSWORD`
- The Reconstruction spell can now target your undead pet instead of always targeting you. It only works on undead creatures

### Fixed
- Pushing coins at the Key’n Anchor locksmith disconnected you and lost your money. The purchase now works and leaves your item on the workbench. Similar crashes are prevented when other scripted items are flipped or read
- Mindlink’s telepathy (from the spell, herb or potion) now wears off after an hour as intended. It used to last forever
- Dispel no longer strips an Ephemeral’s natural telepathy or a Telepathor’s
- The day of the month was off by one: the last day of each month showed as the 1st. Dates now count correctly

### GM & Scripting Notes
- New `@cyberware <player>` lists a player’s implants. `@cyberware <player> add <kind> <grade>` installs one, and `remove <kind>` takes one out, undoing its stat changes. The kinds are arm, leg, ear, armor, telepathor, reflexor and eye
- The script variable `MONTH` is 13 during the Feast Days, as GMSCRIPT.DOC describes. `DATE` is now the day of the month (1–28, or 1–7 during the Feast), as documented, instead of the day of the year
- The `INTANGIBLE` item flag is now read from scripts

## v14.0.0 — 2026-10-16

A major update: darkness is real. Rooms the original game marked as dark can no longer be seen in without a light, as it worked in 1996. Torches, lanterns, the Light and Night Vision spells all matter again, and races that see in the dark finally have that edge. Two races also gain the abilities the manual always promised. Ephemerals are untouchable by ordinary weapons until they go on the attack, and Mechanoids can switch their emotions off for sharper skills. GMs can now set a character’s race, including undead, and have a new view of free room, item and creature numbers. `@dupe` now makes a true copy.

### Added: Darkness
- In a dark room, `LOOK` shows only the room’s name, “It is too dark to see anything.”, and the exits. You can’t see the description, items or people there
- Looking in a direction, into a container, or at anything, as well as `READ`, `SEARCH` and `SKIN`, gets “It is too dark to do that.” Peering into a dark room next door shows only its name
- A lit torch, lantern or flashlight lights the room for everyone there, whether it’s held, worn, carried or lying on the floor. So does anyone with the Light spell active
- Murgs, Highlanders and Ephemerals see in the dark, as does anyone under the Night Vision spell
- Only rooms the game marks as dark are affected. Night-time doesn’t darken other rooms

### Changed: Light and Night Vision Spells, LIGHT
- The Light spell (113) now surrounds you with a glow that lights dark rooms for 20 minutes. Night Vision (520) lets you see in the dark for 20 minutes. Casting either again adds more time, up to 4 hours
- You’re told when each one wears off
- `LIGHT` now finds a torch you’re holding in either hand or wearing, not just one in your pack. It tells you if the item is already lit, and the room sees you light it
- You can’t light anything underwater: “You can’t do that underwater.”

### Added: Ephemeral Abilities
- Ordinary weapons, claws and bites pass harmlessly through an Ephemeral. Spells, psionics, elemental attacks and magical creatures still harm them
- Going on the offense draws an Ephemeral onto the physical plane for 30 seconds: preparing a spell, choosing a psionic discipline, or advancing on or attacking a target. While on the physical plane they can be hurt like anyone else. Each further attack resets the 30 seconds. You and the room see “becomes more substantial” and, later, “becomes more ethereal”
- Ephemerals don’t bleed
- Ephemerals can’t pick up, wield, wear, accept or buy anything over 2 pounds. Gear an Ephemeral already has is left alone
- Darkness gives someone hiding no advantage against an Ephemeral’s `SEARCH`

### Added: Mechanoid EMOTE and UNEMOTE
- `UNEMOTE` switches a Mechanoid to the analytical half of its brain. Mining, smelting, forging, repairing, foraging, gem molding, hiding, sneaking, skinning, picking locks, disarming traps, legerdemain and scripted skill checks all get +10% to their chance of success. Combat, spells and psionics are unchanged
- While unemoted, a Mechanoid intones everything it says (“Alpe intones, …”) and can’t use emotional emotes such as `LAUGH`, `SMILE` or `CRY`. Gestures like `NOD`, `BOW` and `WAVE` still work
- `EMOTE` switches back to normal. Every Mechanoid starts in this mode
- New Mechanoid emotes: `BEEP`, `CLICK` and `BUZZ`

### Changed: Mindlink Messages
- Casting Mindlink (403) on yourself now reads “You gesture.” then “You feel the power of your mind expanding.”, and the room sees that you seem changed
- Casting it on someone else shows you placing your fingers upon their temples, and they feel the power of their mind expanding

### Fixed: Items Losing Details When Moved
- Picking up or dropping an item could lose some of its details: writing on a note, sigils, a blade imbue, or other stored values. `GET ALL` also lost an item’s tail. Every way of moving an item between the floor and your hands now keeps all of them

### GM & Scripting Notes
- `@dupe` now makes an exact copy, including the item’s tail, written text, sigils, blade imbue, crafted value and item bits, and the contents of a container
- `@set race` now works after `@edpl`. Give the race as a number or a name. An unknown name lists the valid races
- GMs can set races players can’t choose at creation: Vampire (9) and the racial groups 21–31 (Extraplanar, Undead, Construct and others)
- An undead or vampire character is immune to poison, disease and bleeding, and any they have is cured when the race is set. Healing harms them, and undead-only spells work for them
- Making a character Ephemeral turns on its innate telepathy, and changing away from Ephemeral turns it off. Changing a Wolfling to another race ends wolf form
- The GM panel has a new Number Ranges tab. It shows which room, item and creature numbers are in use, the free ranges between them, and any numbers used twice. Room, item and creature numbers have no upper limit
- The Number Ranges tab includes a reservation list, so a block of numbers can be noted as set aside for a GM. Reservations are only a reference: nothing stops GMs working in the same numbers when they collaborate

## v13.9.0 — 2026-10-14

You can now help a fallen friend back to their feet with `PULL`. Speaking in cant no longer gives you away to people who can’t understand it, and keeps your capitalization. A mug of ale poured for you by script now actually has ale in it. Scripts can read and set an item’s crafted value, so every place that buys goods can use the same price. The API docs link and the MUD client connection details on the main page are fixed.

### Added: Help Someone to Their Feet
- `PULL <player>` on someone lying down now stands them up. They see “\<you\> helps you to your feet.” and everyone else sees “\<you\> helps \<them\> to his/her feet.”
- It doesn’t work on someone who is unconscious, asleep, dead, or held fast by webs, plants, tentacles and the like. In those cases `PULL` works as before

### Changed: Cant Is Truly Secret
- When you use `CANT`, people who can’t understand it no longer see “\<name\> mutters something under their breath.” They now see nothing at all
- Those who can understand cant still hear what you said, as before
- Your cant message keeps the capitalization you typed instead of being shown in all capitals

### Fixed: Empty Mug of Ale
- Asking Torm for a taste in the Foresters crafts area gave you a pint mug that looked empty. `LOOK IN` now shows the ale, whether the mug is on the floor or in your hands
- You can `DRINK` the ale from the mug once you’re holding it. The same applies to any drink a script pours into a cup or mug

### Fixed: Main Page Links
- The API docs link on the main page returned a “not found” error. It now opens the docs
- The MUD client connection details, and the addresses in the API docs and examples, now show this server’s address instead of the old one

### GM & Scripting Notes
- New variable `ITEMCRAFTEDVALUE` reads an item’s crafted value, and can be set with `EQUAL`, `ADD` and `SUB` like `ITEMVAL1`
- New variable `ITEMWORTH` gives an item’s sale value: its crafted value if it has one, otherwise `ITEMVAL1`. Scripts that buy items, such as the weavers’ storage cabinet, should check `ITEMWORTH` instead of `ITEMVAL1` so crafted and foraged goods are priced the same as at merchants
- `NEWITEM` and `NEWPUT` accept `CRAFTEDVALUE=` alongside `VAL1=` and the others

## v13.8.0 — 2026-10-12

Defeated creatures can now leave more than one treasure item behind, as they could in the original game. Foraged herbs and woods are worth real money at merchants instead of a single copper. Training in a region with its own money, such as Kragenholm, is now paid in the local currency. Items stored by script in a room’s containers, such as the Weavers Storage Alcove cabinet and cooler, now survive a server restart.

### Changed: More Than One Treasure Item
- When a creature drops a treasure item, there is now a chance of a second, and each further item gets the same chance again. A creature that drops an item 60% of the time has about a 1-in-4 chance of a second one
- Every creature that carries treasure can drop up to 2 items. Creatures with more treasure can drop up to 3 or 4
- Each item is chosen separately, so one kill can leave, for example, a piece of jewelry and a locked container. The chance of the first item is unchanged

### Fixed: Foraged Items Sold for 1 Copper
- Everything found with `FORAGE` sold for a single copper. Foraged items now have a value
- Items that a shop also sells are worth that shop’s price, so merchants pay half of it: about 25 gold crowns for mandrake root and nearly 13 for the rare jungle herbs. Items no shop sells are valued by how rare they are
- Foraged items you’re already carrying, and herbs bought from shops, sell at the new value too

### Fixed: Training Fees in Local Currency
- Trainers in regions with their own money charged gold crowns even though the region’s shops take local money. The Kragenholm Weaponmaster’s yard now charges kragenmarks, at the usual rate of 12 to the crown
- The fee itself is unchanged: 5 gold crowns per rank, or the local equivalent, for characters above level 4
- Where training is paid in crowns, you no longer need gold coins on hand. The fee can be paid from silver and copper, with change made, as in shops

### Fixed: Stored Items Lost on Restart
- Some room containers, such as the cabinet and cooler in the Weavers Storage Alcove, take your item and store a copy of it by script. Those copies were lost whenever the server restarted, while the rack next to them kept its pelts. They are now saved like anything put in by hand
- Items stored before this update are still lost at the next restart. Anything stored from now on is kept

### GM & Scripting Notes
- `NEWPUT` into a container now stores the item the same way a player’s `PUT` does, so a container marked with `ITEMBIT19` keeps it through a restart. `LOOK IN`, `GET` and `IFIN` behave as before
- `@rdata` listed items stored inside a container by script as if they were on the floor, with the container’s item number, and showed the container as empty. They now appear in the container’s Contents line, and the Items count leaves them out

## v13.7.0 — 2026-10-10

Game Masters get new tools for running the world. Uploaded scripts can be kept as drafts until they’re ready, and switched on or off at any time without a restart. The season can be held or set by hand. A GM can now restart the server from inside the game, with a countdown warning for everyone online. Seasons also change on schedule again. For players: `PUT … ON` works, balloons can be popped, and a few scripted objects behave better.

### Added: Draft Scripts in the GM Panel
- Scripts uploaded in the GM panel can be saved with **Save as Draft**. A draft is stored and readable by every GM, but it isn’t applied to the game and isn’t loaded when the server starts. Each script shows a **LIVE** or **DRAFT** badge
- **Activate** puts a draft into the game. **Disable** turns a live script back into a draft and takes it out of the game straight away, with no restart: the rooms, items, creatures and timed events it changed return to their original versions, and creatures that only it defined are removed
- Saving changes to a live script now also removes anything the new version no longer contains, and deleting a live script removes it from the game first. Previously both needed a server restart
- Re-uploading a script no longer makes its timed events run twice as often
- Scripts that were already uploaded stay live

### Added: Season Control
- `@season` shows the current season, the game date and roughly how long until the next season change
- `@season hold` freezes the current season, `@season set autumn` (or spring, summer, winter) switches season immediately and holds it, and `@season release` returns to the calendar. A held season survives server restarts. Time of day and dates keep moving while it’s held

### Added: Restart From Inside the Game
- `@reboot <seconds>` (10 to 1800) or `@reboot now` saves every online character, the game clock and global values, then restarts the server, so a GM without access to the host can do it. Everyone online is warned when it’s scheduled and again at 5 minutes, 2 minutes, 1 minute, 30 seconds and 10 seconds. `@reboot cancel` stops it
- It must be typed in full, so an abbreviation can’t trigger it by accident. The server only comes back on its own if the host is set to restart it automatically

### Fixed: Seasons
- The world only changed season while someone had the admin event monitor open, so seasons could go months late. They now change on schedule
- Rooms a live GM script has replaced keep the script’s version when the season changes, instead of being overwritten with the seasonal description

### Fixed: Commands
- `PUT <item> ON <thing>` now works; before, only `PUT … IN` was understood
- `POP`, `RIP`, `LIFT` and `PRY` are now commands, so objects that react to them work. Ripping a balloon to pop it had never worked, because `RIP` wasn’t a command. `PRY <thing> WITH <tool>` is accepted
- Looking at a scripted object now shows only its scripted description. Previously the default description came first; for example, the Cafe de la Paix menu said “It is nondescript.” before “You’ll have to read that.”
- Items handed to you by a scripted event are saved right away, so a crash or restart can’t lose them

### GM & Scripting Notes
- `UNLOCK` now runs a room item’s `IFPREVERB UNLOCK` script before checking whether the item is lockable, the same way `OPEN` already did
- An `IFPREVERB LOOK` script that uses `CLEARVERB` and echoes text now replaces the item’s default description instead of being added after it
- `REMOVEITEM -1` removes the exact item the script is running on (matching adjectives and values) before falling back to any item of the same type. Popping a special balloon can no longer use up a plain one you’re also carrying
- `NEWITEM -1` saves the character immediately
- A GM script’s `CEVENT` replaces an already-loaded one with the same number and room instead of running alongside it
- Disabling a script does not undo vocabulary, named variables or monster adjectives it added. They’re harmless and clear at the next restart. Disabling also replaces the affected rooms, so items dropped in them are lost

## v13.6.0 — 2026-10-08

Fatigue, mana and psi now all recover at the same, faster rate. Searching for hidden players has been reworked so that darkness and night help the hider and no one is ever impossible (or certain) to find. Walking through a door or into an alley while hidden now gives you away; use `SNEAK` instead. Several ways of getting stuck are gone: you can no longer be left permanently stunned, and `DEPART` and `REPORT` always work. Also: looking through windows and fences shows who’s on the other side, and other players now see what happens when you eat or drink something special.

### Changed: Faster, Even Regeneration
- Each minute, fatigue, mana and psi now recover one tenth of the matching stat, rounded up: Constitution for fatigue, Empathy for mana, Willpower for psi. A stat of 75 recovers 8 points a minute while standing
- Mana used to recover noticeably slower than fatigue and psi. All three now use the same rate
- Resting still helps as before: kneeling recovers 1.5 times as fast, sitting twice as fast and lying down three times as fast

### Changed: Searching for Hidden Players
- Only the searcher’s Perception counts now. Your own Stealth skill no longer helps you find others
- A search always has at least a 5% chance and never more than a 95% chance of finding someone, the same limits combat uses
- Hiders get a bonus in a dark room with no light source, and another bonus outdoors between 7pm and 5am. The two add together

### Changed: Moving While Hidden
- Using `GO` to walk through a door, gate or alley now reveals you, just as walking in a direction always has. If the way is blocked, you stay hidden
- `SNEAK` now works on alleys and similar places, as in `SNEAK ALLEY`, not just on directions. `STEAL ALLEY` still works too

### Fixed: Getting Stuck
- Logging out or a server restart while stunned could leave you stunned forever, unable to do almost anything. Stuns no longer survive logging out, and anyone stuck this way is freed the next time they log in
- Dying now clears any stun
- `DEPART` and `REPORT` now always work, whether you’re dead, unconscious, stunned or in round time

### Fixed: Round Time
- `RETREAT`, `ADVANCE` and `FLEE` could be used while still in round time. They now wait, like walking does, and tell you how long is left

### Fixed: Looking Into Other Rooms
- Looking through a window, fence or peephole into another room only showed that room’s description. You now see everything a `LOOK` there would show: the people and creatures present, items, weather and exits. For example, `LOOK FENCE` at the Cafe de la Paix in Fayd shows who’s out on the street

### Fixed: What Others See
- When you ate or drank something with a special effect, only you saw what happened. Biting into a chocolate and spitting out a gem now shows to everyone in the room
- The off-hand swing of a two-weapon attack now reads like the main hand, damage included: “Trieste backstabs at a bog ooze. Hit! Massive damage.”
- A monster’s dropped weapon read “An a greatsword clatters to the ground.” It now reads “A greatsword clatters to the ground.”
- “You see a dagger lies among the remains.” now reads “You see a dagger lying among the remains.”

### Balance: Crescent Isle
- The lower muldragun spawn rate in v13.5.0 never actually took effect, because a second copy of the old values was overriding it. Muldraguns and bhulondags now really do appear less often, about a third as often as originally

### GM & Scripting Notes
- `SHOWROOM` now shows a full `LOOK` of the target room (occupants, monsters, items, weather, exits), as `GMSCRIPT.DOC` describes, instead of only the name and description. This affects all of its roughly 520 uses in the scripts. Hidden and invisible players stay hidden
- Monster group 107 is defined in both `MONSTERS.SCR` and `HAVEN.SCR`. `HAVEN.SCR` loads later, so its copy is the one used. Both are now 30/15/5; keep them identical. Before changing any spawn group, check every script file for other copies of it
- `ECHO OTHERS`, `ECHO ALL` and `GMMSG` in an item’s `EAT` or `DRINK` script now reach the room and the GMs
- `SNEAK <item>` runs the item’s `IFPREVERB STEAL` script, the verb the original alley items use for sneaking in. `GO` through a portal or scripted item un-hides the player before its scripts run, so their `ECHO OTHERS` lines are seen
- Search tuning: a dark room with no light source gives the hider +20, outdoors from 7pm to 5am gives +10
- The admin Monsters page showed item adjectives instead of monster adjectives (monster 95 appeared as “amazing beetle” instead of “fire beetle”). Fixed
- The `*HONK*` alert for a queued treasure drop no longer reads “got an a …”

## v13.5.0 — 2026-10-06

When you `REPORT` a problem, the Game Masters can now write back, and you’ll see their reply the next time you log in. Hiding now works as it did in the original game: it takes 5 seconds of round time and can’t be tried while you’re still in round time, and neither can preparing a psionic discipline. Creatures that aren’t made of flesh and bone no longer get gory death messages. Also: winter worgs are more vulnerable to fire, snowy caterpillars can be skinned, and Crescent Isle is a little less crowded.

### New: Replies to Your Reports
- The Game Masters can now reply to anything you send with `REPORT`. The next time you enter the game, you’ll see what you reported and their reply, just after the room description
- If you’re already playing when they reply, it shows up right away
- Each reply is shown once. If a GM changes a reply, you’ll see the new version on your next login
- Works the same whether you play in the browser or through telnet or SSH

### Changed: Hiding Takes Round Time
- `HIDE` now costs 5 seconds of round time, whether or not you find a hiding place. Haste from a spell or psionic power halves it
- You can’t try to hide while you’re still in round time. You’ll see `[Wait 4 seconds...]` instead
- When you fail to hide, others in the room now see “Rhys gropes around, trying to hide.” instead of “Rhys looks around nervously.”

### Fixed: Psionics During Round Time
- You could prepare a discipline with `PSI` while still in round time. It now waits until your round time is over and tells you how long is left, like `[Wait 4 seconds...]`. Typing `PSI` by itself still lists your disciplines at any time

### Fixed: Gory Deaths for Creatures Without Bodies
- Killing blows on slimes, elementals, constructs, and other creatures without a normal humanoid or animal body could describe ruptured organs, a lacerated liver, and similar deaths. These creatures now get the normal damage line instead
- This applies to weapon kills (including off-hand weapons), elemental weapon criticals, and spell kills

### Balance: Winter Creatures and Crescent Isle
- Winter worgs now take extra damage from heat and only half damage from cold
- Snowy caterpillars can now be skinned for a snow-white pelt. Their spine attack also no longer reads “flings its spines spines”
- Muldraguns and bhulondags on Crescent Isle appear about half as often as before

### GM & Scripting Notes
- Player reports on the admin Logs page have a new **Reply** button. Under each report you can see the reply and whether the player has seen it yet. **Edit Reply** changes it (and the player sees it again), **Remove** deletes it. Replying doesn’t mark the report Done; that’s still separate. Replies are signed “The Game Masters”, so your account name isn’t shown to players
- Replies go to the character with the reporter’s full name, and disappear with the report after 90 days if never seen
- `IMMUNITY 9` (susceptibility to death criticals) is now enforced. Any value below the default of 2 turns off the fancy death lines for that creature; most creatures that use it are set to `IMMUNITY 9 0`. The creature’s own `TEXD` death text is still shown
- Winter worg (`WINMON.SCR`, monster 44): added `IMMUNITY 3 3` and `IMMUNITY 5 1`. Snowy caterpillar (monster 92): added `SKINITEM 509` with `SKINADJ 1108`
- Monster group 107 (Crescent Isle) in `MONSTERS.SCR`: 100/50/10 → 50/25/5

## v13.4.0 — 2026-10-04

Backstab is now a real skill: you need training in it, a puncture weapon, and a target with a back, and every successful backstab strikes that back. Potions can be poured together again to top them off or gamble on a mix, as in the original game. Locked chests turn up a little more often, a damaged weapon warns you before it breaks, and experienced skinners rarely ruin a corpse. Also: winter spawns in the Dark Keep are back to normal, and you can take out items that were placed in a room’s containers.

### Changed: Backstab Requires Training
- `BACKSTAB` used to work for anyone hidden with a dagger or rapier. You now need training in the Backstab skill, and Stealth is required before you can train it
- You still need to be hidden and holding a puncture weapon. If you fight with two weapons and have Two Weapons training, a dagger or rapier in your off hand works too, and your main weapon follows up with a normal swing
- Your chance to hit comes from your weapon skill, as with any attack. Every hit lands on the target’s back for full damage, and each level of Backstab adds to that damage
- You close in on the target as part of the strike, so you don’t need to `ADVANCE` first
- Creatures with no back, such as snakes, slimes, mists, and some insects, can’t be backstabbed
- A backstab that can’t go ahead (wrong target, no back, still in round time) now leaves you hidden. Before, it revealed you anyway
- The Weaponmaster’s yard in Kragenholm trains both Stealth and Backstab, and you don’t have to join a guild to use it

### New: Pouring Potions Together
- `POUR` into a container that already holds liquid now mixes the two instead of refusing
- The same kind of potion combines into more sips, though the mixture changes color. Water or ale simply tops off water or ale
- Two different potions also combine into more sips, but what you end up with is a gamble: usually one of the two original potions, sometimes a completely different potion, a Body Destruction potion, or a useless liquid. Occasionally the mixture explodes and destroys the container
- Mixing water or ale into a potion dilutes it, and the magic is lost
- A container holds 12 sips. Pour in more than that and the extra spills away

### Improved: Damaged Weapon Warning
- Wielding a damaged weapon now warns you: “Your steel broadsword is damaged, you risk it being destroyed if you continue using it!” Another weapon clash will break it

### Improved: Enchanted Item Names
- The Enchantment I, II, and III spells now put their word first, so you get “an ensorcelled randar broadsword” rather than “a randar ensorcelled broadsword”. Items enchanted before this release keep their current names

### Fixed: Advancing on Something Already Advancing on You
- After you `ADVANCE`d on a creature, it could still announce “advances towards you” a moment later. Once you’ve closed in, it now just fights you
- `ADVANCE` on a creature or player that has already closed in on you now tells you “You are already engaging” them instead of advancing a second time

### Fixed: Taking Items Out of Room Containers
- Items placed in a room’s cabinet, chest, or other container by the room itself showed up when you looked inside, but `GET` claimed “You don’t see that in there.” Now you can take them, one at a time or all at once

### Balance: Skinning, Treasure, and the Dark Keep
- Expert skinners ruin far fewer corpses. With enough Sagecraft and Wood Lore, a skin now fails about 1 time in 100 rather than 1 in 20. Beginners’ chances are unchanged
- When a creature drops an item, a locked chest turns up a little more often (about 1 in 5 item drops instead of about 1 in 7). Armor drops are a little rarer to make room
- The last release cut the Dark Keep’s winter spawns too far, so it was much quieter than in other seasons. Winter now spawns at the same rate as the rest of the year

### GM & Scripting Notes
- New `ADJCVT <digit> <value>` stores an adjective’s word in `%0`–`%9`, where `STRCVT` would store the number. The value can be a number or a variable (`ITEMADJ3`, `ITEMPARAM1`...). It treats any value as an adjective number, so use it only on values that really are adjectives. 0 or an unknown number gives an empty string
- `%0`–`%9` are one shared set: `STRCVT`, `ADJCVT`, `STRCPY`, and `STRCAT` all write the same ten slots, so reusing a digit replaces what was there. The slots start empty each time a script runs and after a `PLREVENT` delay
- New read-only variables `ITEMPARAM1`–`ITEMPARAM3` give the current item’s PARAMETER1–3 values. These belong to the item definition, so `EQUAL ITEMPARAM1 ...` does nothing
- `NEWPUT` now accepts variables for the item number and every `ADJ`/`VAL` value, so a container’s PUT script can store a copy of what was put in: `NEWPUT 0 ARCHNUM ADJ1=ITEMADJ1 ADJ2=ITEMADJ2 ADJ3=ITEMADJ3 VAL1=0`. The first number is the container’s item tag in the room, not `-1`
- `IFIN <tag> <item # or variable>` checks both what players have put in a container and what scripts placed there. It doesn’t yet set the “current item” as GMSCRIPT.DOC describes
- Backstab (skill 2) now requires Stealth (33) to train. Creatures defined only with `PART` lines need `PART 2` (back) to be backstabbed; `HUMAN`, `ANIMAL`, and `AVINE` creatures always have one
- Monster group 47 (Dark Keep) in `WINOUT.SCR`: 15/15/11 → 40/40/30, matching the other three seasons

## v13.3.0 — 2026-09-30

Ming K’Tuk’s elkyri door works on every visit now, not just the first: a door the dragon had shut again no longer answers your knock with “The door is open, fool!” while staying firmly closed. The Dark Keep no longer sends a horde at you as soon as winter arrives, while the undead of the Fayd cemetery rise more often. Walking out of the Golden Chalice saloon no longer claims you can’t go that way, and GMs get a friendlier `@editem`.

### Fixed: Ming K’Tuk’s Door Stuck Shut After Its First Opening
- Once the heavy elkyri door in the Deep Realms had swung shut again at noon or midnight, knocking on it made the dragon face insist “The door is open, fool! Stop knocking and go in!” — but the door stayed closed, no riddle was given, and the way was blocked until the server restarted
- A door that has been shut and locked now counts as closed, so knocking gives you a riddle again and the right answer opens it, however many times it has closed before
- The same fix applies to any door or container in the world whose script checks whether it’s closed, locked, or unlocked. A few that could never respond correctly before now do

### Balance: The Dark Keep in Winter
- In winter, every creature type inside the Dark Keep appeared on every spawn check, so a fresh wave arrived before you could finish the last one — a problem dating back to the original winter data, not something new. Winter spawns in the keep are now in line with the rest of the game

### Balance: More Undead in the Fayd Cemetery
- You could walk every row of the cemetery for minutes and meet only a skeleton or two. Skeletons, zombies, ghouls, lesser wights, and phantom caretakers now rise several times as often, in every season. The cemetery holds no more of them at once than before

### Fixed: “You can’t go that way” Leaving the Saloon
- Going `OUT` through the swinging doors of the Golden Chalice saloon took you to the bazaar and then also told you “You can’t go that way.” You now simply step out into the bazaar
- People left in the saloon now see you go out through the swinging doors; before, that message appeared out in the bazaar instead
- This applies to any exit whose room script moves the whole group through it, including `ASCEND` and `DESCEND`

### GM & Scripting Notes
- `@editem` no longer refuses when several items match. It edits the **last** matching item in inventory (normally the one most recently picked up) and lists every match with a `*` beside the one it changed. If nothing in inventory matches, it uses the last wielded, worn, or container match
- `@editem` and `@iexamine` accept adjectives in any order and by prefix, so `@edit steel rapier tail polished to a high sheen` or `@iexamine pol rap` finds “a polished steel rapier”
- `IFITEM <ref> CLOSED` is now true for a locked item, and `IFITEM <ref> UNLOCKED` is true for anything that isn’t locked, so open/closed and locked/unlocked can be tested independently as the original scripts do (e.g. `NEWCITY.SCR` nests `CLOSED` around `LOCKED`)
- Script `CLOSE` no longer unlocks an item that’s already locked, and script `UNLOCK` leaves the item closed, as the player `UNLOCK` command does. Previously a script-unlocked door could be walked through without being opened
- A `MOVEGROUP` in an `IFPREVERB` on a direction or `ASCEND`/`DESCEND` now always ends the command, with pre-move echoes sent to the room being left
- Monster group 47 (Dark Keep) in `WINOUT.SCR`: 400/400/300 → 15/15/11. Monster group 4 (Fayd cemetery, all five Fayd files): skeleton/zombie 10, ghoul 6, wight 4, caretaker 6

## v13.2.0 — 2026-09-29

First and most important: the server no longer throws everyone out of the game every minute or so. A crash introduced with the last release took the whole server down whenever a monster wandered, and the server itself now shrugs off that kind of failure instead of disconnecting every player at once. Beyond that, fighting with two weapons finally lands proper critical hits on a stunned or fallen foe, the sharkhors along the Inner Sea turn up more often, and wolf spiders now make their home on the forested hilltop of Nyraine — which means the spider fangs one of the fold’s quests asks for can finally be found.

### Fixed: Everyone Disconnected Every 15 to 60 Seconds
- Shortly after the last update, every player was being dropped from the game at the same moment, over and over. The cause was the new “a monster has appeared” announcement: under some server setups it could crash the whole game the moment any creature anywhere wandered from one room to another. That can no longer happen
- More broadly, if something goes wrong while monsters are acting or a timed room event is firing, the server now logs the problem and carries on instead of shutting down and disconnecting everyone

### Fixed: Off-Hand Attacks Never Scored Critical Hits
- With two weapons, your off-hand swing always reported a plain “Hit!” and dealt ordinary damage — even on a natural 96 or better, and even against a creature your main hand had just knocked off its feet or stunned
- The off hand now follows the same rule as the main hand: a roll of 96+, or any hit on a stunned or knocked-down target, is an **Excellent Hit** (or **Devastating Critical**) with the bonus damage to match, and onlookers see it too

### New: Wolf Spiders in Nyraine
- Wolf spiders now hunt the hilltop firs, poplars, and meadow at the edge of the Nyraine forest, in every season
- They’re the only creature whose fangs can be skinned as **spider fangs** — until now wolf spiders never appeared in the world on their own, so the fold’s quest that asks for spider fangs couldn’t be completed without a GM’s help

### Balance: More Sharkhors Along the Inner Sea
- Sharkhors and sharkhor priests on the slippery rocks of the Inner Sea were appearing too rarely to be worth the trip. They now turn up about half again as often, in every season

### Fixed: Nyraine Pond Was Empty in Spring
- The pond in the Nyraine forest had its usual creatures in summer, autumn, and winter, but none at all in spring. It now matches the other seasons

### GM & Scripting Notes
- Items can now listen for speech: an `IFSAY` block on an item a player is wielding, wearing, or carrying fires when they say the phrase, with that item as the script’s item (so `ITEMVAL`, `ITEMBIT`, and `IFITEM -1 WORN` all work). Previously only rooms could respond to speech, so original items with spoken triggers never reacted
- When something a player says or pays for moves them (`MOVE` from an `IFSAY` or `PAY` script), the script’s own echoes now go to the right place: lines echoed before the `MOVE` are seen in the room they left, and lines after an `AFFECT` to the destination are seen where they arrive. The generic “leaves”/“arrives” lines are used only when the script doesn’t supply its own
- `MOVE` now ignores a room number that doesn’t exist instead of stranding the player in it — useful when the destination comes from an item value a GM sets
- Nyraine’s hilltop rooms 1953–1955 use the new monster group 122 (`MLIST 122 94 10 3`)

## v13.1.0 — 2026-09-26

This release is about knowing what’s in the room with you and getting a fair chance to react. Hostile creatures now wait a moment before closing to melee when you walk in, so a traveler moving at a normal pace can pass through without being pinned down by every creature that happens to be there. Players with `BATTLEBRIEF` on were never told when a monster appeared, wandered in, or wandered off — that’s fixed, and a creature summoned by a script is now announced to everyone, including whoever set it off. Monsters also no longer notice or go after invisible players. And for anyone who has stood at the elkyri door in the Deep Realms giving the right answer to no effect: the door opens now.

### Changed: Monsters Give You a Moment Before Advancing
- A hostile creature used to advance on you the instant you entered its room — often before you’d even seen the room description or the exits. It now waits 3 seconds after you arrive, long enough to read the room and keep moving
- The same 3-second wait applies when a creature already in the room first notices you
- Leaving a room and coming back no longer triggers an instant advance left over from your earlier visit

### Fixed: Monsters Sometimes Appeared Without Any Message
- With `BATTLEBRIEF` on, you never saw a monster appear, wander in, or wander off. `BATTLEBRIEF` now does only what the manual describes — it hides creatures’ flavor actions and trims combat scroll — and monsters arriving or leaving are always shown
- When a room or item script summons a creature, the player whose action caused it now sees it appear too — before, only everyone else in the room did
- A creature summoned while the player who triggered it is hidden or invisible is still announced to the whole room

### Fixed: Monsters Could Go After Invisible Players
- Hostile creatures could notice and advance on a player who was invisible or in Phantom Form. Monsters now never pick a hidden, invisible, or Phantom Form player as a target, and a creature whose target vanishes that way gives up on them

### Fixed: The Elkyri Door Riddle
- Answering the draconic door’s riddle correctly announced that the doors swing open, but the door stayed shut. It now actually unlocks and opens

### Fixed: Stray “%b” in Item Messages
- Putting things into certain items — the mortar, a few garments, the hole in the clouds — showed a literal “%b” where the item’s name belonged (“You place a dried herb into %b”). The name now appears

### GM & Scripting Notes
- Script text now supports `%b`/`%B`: in an `IFPREVERB2 PUT` block, `%a` is the item being put in and `%b` is the item that owns the script (the container)
- `IFVAR ITEMFLAG<name>` works with any item flag, not just DYE and SKIN — e.g. `ITEMFLAGCRAFTABLE`, `ITEMFLAGDYEABLE`, `ITEMFLAGLIGHTABLE`
- Newly forged weapons, armor, and shields now store their sell value in `CraftedValue` (shown by `@iex`) instead of Val1, like every other crafted item. Items forged earlier keep their Val1 value and sell for exactly what they did before
- Script macros no longer drop steps like `UNLOCK` and `OPEN` — the cause of the stuck elkyri door

## v13.0.0 — 2026-09-24

A major addition: regional currencies. The original game defined its own money for the far-flung regions — kragenmarks in Kragenholm and the Deep Realms, shards in Idemmu Ag and Salicae, darktars in Ilzar — but until now nothing actually used them, so every shop, bank, and healer in those regions charged and paid gold, silver, and copper. Monsters there now carry the local currency, and the local merchants, banks, and physicians deal in it. Alongside that, the Legerdemain skill finally does what it says: appraising items, filching, picking pockets, and shoplifting are all live. Skinning now depends on Sagecraft and Wood Lore the way the manual always described, being stunned actually stops you from doing things, and running out of flight while over open sky now drops you to the ground instead of leaving you stranded in mid-air.

### New: Regional Currencies
- Kragenholm and the Deep Realms use **kragenmarks** (12 to the gold crown), Idemmu Ag and Salicae use **shards** (3 to the crown), and Ilzar uses **darktars** (100 to the crown)
- Monsters in those regions carry the local money — always in Ilzar, most of the time in the Deep Realms and Idemmu Ag, exactly as the original region definitions set it. A creature decides once what it’s carrying, so its dropped coins, the coins in any chest it leaves, a `SEARCH` of its corpse, and a successful `ROB` all pay in the same currency
- Shops in those regions sell at their posted local prices and pay local money for anything you `SELL`; `APPRAISE` quotes in it too. Banks there keep a separate account in their own currency (`DEPOSIT 50`, `WITHDRAW 50`, `BALANCE`), and physicians charge in it
- `GET`, `DROP`, `GIVE`, and `SPLIT` all work with regional coins — e.g. `GIVE 20 KRAGENMARKS TO ARIS` or `SPLIT 90 SHARDS`
- `WEALTH` now lists only what you’re actually carrying: one line for your crowns, shillings, and pennies (leaving out any you have none of), then one line for each regional currency you hold
- There’s no money changer — the original game never had one. Money earned in a region is spent in that region

### New: Legerdemain — Appraise, Filch, Pickpocket, and Shoplift
- `APPRAISE <item>` away from a merchant now gives your own estimate of what a merchant would pay. It’s within 40% at level 1, tightening by 10% per level, and exact from level 5. Appraising the same item again gives the same answer
- `FILCH` (or `STEAL`) `<item>` picks something up off the ground without anyone seeing you do it. If the attempt fails you still get the item, but everyone sees you pick it up — or sees “something” pick it up, if you’re invisible
- `ROB` or `PICKPOCKET` (shortest form `PICKP` — `PICK` is still lockpicking) `<player or creature>` lifts coins from their purse. A player who notices only feels a tug on their coin purse, never who did it. `ROB` or `STEAL` with no target, in a shop, lifts coins from the till — up to one gold crown’s worth per skill level
- Your odds improve with Agility, Quickness, and skill, at night, in an unlit dark room, and while invisible; a player target’s Perception works against you. Every attempt shows your chance and roll, is capped at 95%, and costs 5 seconds of round time (halved by Haste)
- Each player, creature, or shop can only be robbed once per game day, and a successful rob earns 50 experience per Legerdemain level

### Changed: Skinning Now Uses Sagecraft and Wood Lore
- `SKIN` never checked any skill at all — anyone could do it and it never failed. It now requires Sagecraft or Wood Lore, as the manual describes, and can fail; a failed attempt ruins the corpse
- Both skills raise your odds of success, and higher Sagecraft can salvage a second or even third specimen from the same creature

### Changed: Being Stunned Actually Stops You
- A stunned player could still `ADVANCE`, `RETREAT`, `FLEE`, and even walk out of the room. Now no physical action works until the stun wears off — you can still look around, check your status, talk, and use telepathy
- The psychic shock from losing a summoned creature is now a real stun rather than just a round-time delay

### Balance: Bows and Crossbows Hold Up Better in a Weapon Clash
- Light missile weapons were damaged by almost every weapon clash, often before their owners even reached a hunting ground. Bows, crossbows, slings, and blowguns now get a large bonus to clash resistance
- They get an extra bonus if made from mirmach, oak, rosewood, or uquart, and +10/+20/+30 when enchanted, ensorcelled, or eldritch — including crafted bows whose enchantment never showed in their name

### Changed: Combat Tweaks
- A creature that’s already on the ground — knocked down, asleep, or still getting up — can’t be knocked down again until it’s back on its feet; the critical hit stuns it instead. The same goes for players, who must be standing to be knocked down
- `ADVANCE` on a creature that another creature is guarding now runs into the guard, and you advance on the guard instead — matching how `ATTACK` already worked
- Earth Wave’s knockdown now lasts as long as its stun, instead of the creature being back up within a few seconds

### Fixed: Stranded in the Sky When Flight Ran Out
- If a Fly spell ran out — including while you were logged out — or you let go of psionic Flight, or had Fly dispelled while over open sky, you were left “standing” in mid-air, unable to `DESCEND` and with no way down. You now fall to the ground below and take some crushing damage (never enough to kill you)
- Drakin, who fly on their own wings, no longer get grounded when a Fly spell on them runs out, and anyone who can still fly is put back in the air if they log in over open sky

### Fixed: Mushroom Wine and Stew at the Drunken Murg
- The tavern in Kragenholm advertised mushroom wine and mushroom stew, but `BUY MUSHROOM WINE` said it wasn’t for sale — the items were stocked under a different name. They now sell as advertised

### New: GM Tools for Regional Money
- `@set` / `@setp` accept any currency name for a player’s purse (`@set kragenmarks 50`) and `bank` + the name for their bank account (`@set bankshards 20`), plus the new `bankgold`/`banksilver`/`bankcopper`. `@peek` reads them, and `@qstat`, `@edpl`, and the admin panel show them
- `@rdata` now shows a room’s region and which currency it uses

## v12.14.0 — 2026-09-23

Today’s headline is a real progression bug: several non-combat sources of experience (crafting, lockpicking, first aid) added XP directly without ever checking for a level-up themselves, relying on some later action to notice a level threshold had been crossed. But the two most common such actions — `STATUS` and `EXP` — quietly advanced a character’s level to match without granting any of the level-up bonus, and once that happened it could never self-correct, leaving Body Points/Fatigue/Mana/Psi stuck at their starting values no matter how many levels followed. A related formula bug meant Mana was also being padded by Willpower on top of Empathy at every level and regen tick; it’s now Empathy alone, matching the manual. Also fixed: a stunned player could stay stunned for up to a full minute instead of the intended few seconds, `Divine Blessing` couldn’t be cast on a summoned or controlled creature (the same gap `Haste`/`Strength`/`Agility`/`Mystic Armor` had before last release), and turquoise was quietly worthless to sell no matter its quality.

### Fixed: Body Points, Fatigue, Mana, and Psi Could Stop Growing With Level
- Several non-combat experience sources (crafting, lockpicking, healing/first aid) added XP straight to a character’s total without ever triggering a level-up check, leaving it to whatever happened next to notice. But checking `STATUS` or `EXP` before your next kill silently synced your level to match your experience with zero stat bonus granted — and because the next real check saw your level already “caught up,” the missed bonus could never be recovered on its own. A character who leveled this way could be stuck at their starting Body Points/Fatigue/Mana/Psi indefinitely
- Fixed at the source: the level-up bonus is now applied for every level actually gained, no matter which code path is the one that notices the change — including a big multi-level jump from a single large experience award, which now grants every level’s worth of bonus instead of just one

### Fixed: Mana Wrongly Scaled With Willpower
- Both the per-level Mana bonus and passive Mana regeneration were factoring in Willpower alongside Empathy. Willpower governs Psi, not Mana — Mana growth and regen are now based on Empathy alone

### Fixed: Player Stuns Could Drag On for Up to a Minute
- A melee crit’s stun is only meant to last 3–6 seconds, and always has for monsters — but a stunned player’s flag was only ever cleared by the once-a-minute background regen tick, so it could visibly pin a player for up to a full minute (averaging around 30 seconds) after it should have worn off. It now clears the instant your next command is processed, as soon as it’s actually expired

### Fixed: Divine Blessing Couldn’t Target a Summoned or Controlled Creature
- `Divine Blessing` still used old player-only targeting logic, so casting it on your own caretaker or another player’s pet failed with “you don’t see it here” — the same gap `Haste`/`Strength`/`Agility`/`Mystic Armor` had before last release’s fix. It now recognizes a summoned/controlled creature by name and grants it the same combat bonus for a minute

### Fixed: Turquoise Always Sold for 1 Copper
- Turquoise was missing entirely from the gem valuation table, so selling or appraising it to a merchant fell through to a generic near-zero fallback regardless of quality or size. It now has a proper base value like every other gem

## v12.13.0 — 2026-09-21

A big batch driven almost entirely by player reports about summoned and controlled creatures. Commanding one to `ATTACK` used to permanently strip its guard/follow duty instead of just pausing it, its kills paid the summoner zero experience, and none of `Haste`, `Strength`, `Agility`, `Mystic Armor`, or the Globe of Protection family actually took effect on it despite the spell reporting success — all four traced back to the same root cause, a buff being written to a throwaway copy of the creature instead of the real one. Monster pacing got two genuine fixes too: a hostile monster now waits a beat (1–2 seconds) after first spotting you before it advances, instead of jumping a player who’s just passing through a room, and a knocked-down creature now stays down long enough for a follow-up hit to actually land the automatic critical it’s supposed to grant, instead of scrambling back up almost instantly. Also fixed: `RUB`/`TAP`/`CONCENTRATE` could target an unworn spare copy of an item instead of the one you’re actually wearing, silently no-opping any script gated on it being equipped; `REPAIR` only ever recognized Weaponsmithing at a forge, so a Wood Lore crafter could never repair the bows they themselves made; and a scripting-engine bug was quietly leaking Fayd’s “Day of Rebellion” region’s ambient war echoes into unrelated rooms clear across the map. On the lighter side: commands can now be chained on one line with `;` like a MUD client, sheep can be sheared for wool without killing them, and yes — `CURSE ME` does what you remember.

### Fixed: Summoned & Controlled Creatures — Guard Duty, Kill XP, and Buffs
- `COMMAND ATTACK` deliberately stops a creature from following you while it fights (so a room move can’t yank it off a live target), but nothing ever resumed that duty afterward — a creature commanded to attack even once was permanently detached from guarding and following you. It now automatically resumes guarding/following as soon as its fight ends
- A kill by your summoned or controlled creature only ever sent you a flavor-text notice (“Your skeleton has slain a goblin!”) with zero actual reward. It now runs the same experience/loot/treasure/level-up pipeline a normal kill does
- `Haste`, `Strength I/II/III`, `Agility I/II/III`, `Mystic Armor`, and the Globe of Protection/Mass Protection/Spectral Shield/Ride the Lightning family all reported a successful cast on a summoned/controlled creature but never actually applied — the write was silently landing on a temporary snapshot instead of the creature the game actually simulates. All five now take effect for real

### Fixed: Monsters Jumped Players Instantly, and Knockdowns Ended Too Fast
- A hostile monster used to lock onto and advance on a player the very first moment it noticed them — often within a second or two of just walking through a room. It now waits a randomized 1–2 seconds after first spotting someone before actually closing in, giving players passing through a beat to keep moving
- A monster guarded by another creature or a player using Combat Maneuvering’s `GUARD` already couldn’t be approached — hostile monsters now respect that too, announcing that they tried to advance but were blocked, instead of aggroing the guarded player anyway
- A knocked-down creature used to recover on its own very next action tick, which could be faster than a player’s follow-up swing — so the “any hit against a downed target is an automatic critical” bonus (added last release) rarely had a real target to apply to. Knockdown now lasts a proper 3–6 second window, the same way Stun already does

### Fixed: RUB/TAP/CONCENTRATE Could Grab the Wrong Copy of an Item
- If you were carrying a second, unworn copy of an item (e.g. a spare amulet next to the one you’re wearing), `RUB`/`TAP`/`CONCENTRATE` and similar verbs always matched the unworn one first. Since most such effects are scripted to only work while the item is actually equipped, this silently did nothing. Worn and wielded items are now checked before plain inventory

### Fixed: REPAIR Excluded Wood Lore Crafters
- `REPAIR` only ever checked for a forge and Weaponsmithing skill, so a Wood Lore crafter could never repair a damaged bow or other wooden weapon — not even one they crafted themselves. It now recognizes Wood Lore at a fletcher’s workshop for wood-substance weapons, the same split `CRAFT`/`WORK` already use

### Fixed: A Scripting Bug Leaked Ambient Echoes Into the Wrong Rooms
- The script interpreter’s `RANDOM` command silently dropped its optional offset argument, which several original scripts rely on to shift a roll into an absolute range (e.g. picking a specific room number). Fayd’s “Day of Rebellion” region script used this to pick one of its own war-torn rooms to echo ambient text into — with the offset dropped, the roll landed as a tiny 1–24 value instead, and any unrelated room that happened to share that number (a report traced it to the Foundry) received Fayd’s rebellion-era ambient echoes out of nowhere. Fixed at the source, so every script depending on this pattern is corrected

### New: Chain Commands on One Line with “;”
- Like most MUD clients, multiple commands can now be stacked in a single line separated by `;` — `retreat; go east` or just `ret;e`. Quoted speech and `SAY` are exempt, since a semicolon there is just punctuation, not a separator

### New: Sheep Can Be Sheared Without Killing Them
- `GET` on a live sheep attempts to catch and clip it for wool instead of just refusing — a Dyeing/Weaving skill check against an awake sheep (miss and it dodges, badly miss and it bites back), or a guaranteed clip against one put to sleep with the Slumber spell. A dead sheep’s wool is ruined; skin it for a pelt instead

### Also Today
- `CURSE ME` is back — a private “Having a bad day?” only you ever see
- Wolflings must unwield any weapon and shield before `TRANSFORM`-ing into wolf form, instead of transforming with them still equipped
- GM command `@lwho` now shows each online character’s account/login name alongside their character name
- A container script’s `IFIN` check now sees items a player has actually put into a room container at runtime, not just a script’s own starting contents

## v12.12.0 — 2026-09-19

A big one today. The headline is a real combat-mechanics addition: a stunned or knocked-down target — player or monster — is now defenseless, and any successful hit against one is an automatic critical, not just a lucky roll. Player stuns and knockdowns are a genuinely new mechanic as part of this; monsters have always been able to stagger you, but nothing ever actually happened when they did. `FLEE` also got a serious overhaul: it never learned about the newer `ADVANCE`/`RETREAT` system, so it could leave other attackers still locked onto you even after a “successful” escape, and it wrongly refused to work at all unless you’d already landed a hit. Three new player conveniences: `SAYTO` for addressing one person publicly, `WRITE`/`READ` for leaving notes on writable items, and `INVENTORY` now accepts a search term. On the bug side: missile ammo bundles are finally tracked as real stacks instead of behaving like a single unit no matter how many you’re carrying, a subtle bug could silently drop the message and effects of certain scripted room arrivals, and several item messages had an awkward doubled word like “retrieve your some steel.”

### New: Automatic Critical Hits Against Stunned or Knocked-Down Targets
- A stunned or knocked-down opponent can’t defend itself, so any successful hit against one now lands as an excellent/critical hit automatically — not just when the attack roll itself happens to be a natural 96 or better. This applies both ways: players benefit from it against staggered monsters, and monsters benefit from it against a staggered player
- Player stuns and knockdowns are new: an excellent hit against a healthy, standing player now has a chance to stun them (briefly unable to act) or knock them down (forced to stand back up before doing anything else), mirroring the stagger mechanic monsters have always had. A flying player can only be stunned, never knocked down, since there’s no ground to knock them onto

### Fixed: FLEE Didn’t Know About ADVANCE/RETREAT
- `FLEE` predates the `ADVANCE`/`RETREAT` melee system and was never updated for it. If more than one monster had closed in on you, fleeing only broke off the most recent one — the rest stayed locked on as if you’d never left. It now fully disengages from everything, the same way `RETREAT` does
- `FLEE` also refused to work at all (“You are not in combat”) unless you’d already landed an attack — so a player who’d only been advanced on or noticed a monster closing in, but hadn’t swung yet, had no way to run. It now works whenever you’re actually engaged, attack or not
- A panicked flight now only ever picks a compass direction (N/S/E/W/NE/NW/SE/SW) out of the room, never up, down, or through a portal

### New: SAYTO — Speak Publicly to One Person
- `SAYTO <player> <message>` is a public statement addressed to someone specific — everyone in the room hears it (unlike `WHISPER`, which is private), but it’s clear who you’re speaking to: others see “*Name* says to *Player*, ‘message’” and you see “You say to *Player*, ‘message’”

### New: WRITE and READ — Leave a Note on an Item
- Items scripted as writeable (like a blank note) can now actually be written on and read back — `WRITE <item> WITH <text>` followed by `READ <item>`. Previously the writeable flag was silently ignored, `WRITE` was just flavor text with no real effect, and nothing was ever saved to read back

### New: INVENTORY Accepts a Search Term
- `INV vial` now lists every vial you’re carrying — glass, onyx, crystal, whatever — instead of making you scroll your whole pack looking for it. Narrow further with an adjective (`INV glass vial`) the same way `GET`/`DROP` already let you target items. Plain `INVENTORY` still lists everything as before

### Fixed: Ammo Bundles Behaved as a Single Unit
- Loading a bow or crossbow from “some arrows” consumed the entire stack in one shot, no matter how many were actually in it — a bundle of 24 behaved exactly like a single arrow. Bundles now track their real count and deplete one unit at a time
- Bundling machines like the Bowyer & Fletcher’s in Faydindria couldn’t distinguish between two same-named deposit slots (arrows vs. bolts) — the first one always won even when it explicitly rejected what you were putting in it. Ordinal targeting (`PUT bolts IN 2nd slot`) now actually reaches the second one

### Fixed: Some Scripted Room Arrivals Silently Dropped Their Message
- A script that moves a player into a room with its own arrival reaction (a warning, damage, or being sent right back) could compute that reaction correctly but never actually deliver it — the player was just silently teleported with no explanation. This affected every script using that mechanism, including the Dark Cathedral’s inner chancel, whose guardian wards deal a punishing shock and send you straight back out if you’re not carrying a properly bonded ritual dagger

### Fixed: Doubled Article in Item Messages
- Several messages read awkwardly with a repeated word, like “You hand over your money and retrieve your some steel” when buying raw material from a merchant. Buying, loading a weapon with ammo, repairing a weapon at the forge, and molding/polishing a gem are all corrected

## v12.11.0 — 2026-09-18

The headline is a follow-up correction to yesterday’s monster-resistance fix: the penalty math had a hidden bug that doubled up a baseline penalty, making most monsters far more resistant to magic than intended, and a resisted cast still announced “Success!” right before saying the target resisted — which read exactly like the two-chance bug it was supposed to fix, even though the roll itself really was a single one. Both are corrected, and GMs now properly bypass resistance checks entirely, as their guaranteed casting success was always meant to imply. Also today: weapons damaged in combat are now sheathed automatically instead of staying in your hand where the next clash could destroy them outright, only creatures with legs — and not the truly massive — can be knocked off their feet in melee, dyeing an item can no longer erase an existing adjective like a metal’s type, `WORK` now accepts a metal’s full descriptive name and never drops part of it when matching by a shorter word, and a short, pluralized guess like `SELL BAS` can no longer grab a completely unrelated item. Three smaller content fixes round out the day.

### Fixed: Monster Resistance Penalty Was Miscalibrated, and Falsely Announced “Success” Before Failing
- Yesterday’s fix made resistance a modifier on your single casting roll instead of a second roll, which was the right idea — but the formula used to size that modifier effectively applied an unconditional extra penalty even when your skill and the monster’s resistance were evenly matched, making most monsters resist far more often than intended, sometimes almost always. That baseline is removed: an even matchup now costs no penalty at all, and only a real skill deficit or a genuinely tough monster reduces your odds
- Separately, a cast that ended up resisted still displayed “[Success: X%, Roll Y] Success!” before the message explaining the target resisted — two contradictory lines for what is mechanically one outcome. A resisted cast now correctly reads “Failure” from the start
- GMs’ casting success is hardcoded to 100% so their spells never fail for testing purposes, but a resisted cast could still slip through that guarantee. Resistance checks are now fully bypassed for GMs, matching how their cast-success roll already worked

### New: Damaged Weapons Are Sheathed Automatically
- A weapon that takes damage in a clash is now put away the instant it happens, with a warning that you sheathed it before it could break completely. Previously it stayed in hand, and a player who didn’t notice the damage message could keep swinging until the next clash destroyed the weapon outright

### Fixed: Melee Knockdown Applied to Every Creature Equally
- An excellent melee hit had a chance to knock any target off its feet, including creatures with no feet to knock out from under them — a rattlesnake could be “knocked off its feet” the same as a troll. Knockdown now checks the creature’s actual anatomy (legs or no legs) and size; giants, dragons, and other truly massive creatures are also too big to topple. A creature that can’t be knocked down still gets stunned instead on the same roll, so an excellent hit still has some effect on it

### Fixed: Dyeing an Item Could Erase Its Existing Adjectives
- `DYE` used to write a dye’s color words into fixed slots on the target item, overwriting whatever adjective already happened to live there. Dyeing a piece of elkyri metal, for example, could erase “elkyri” from its name entirely. Dye colors are now added in front of whatever’s already there instead of overwriting it; if there truly isn’t room for both, dyeing is refused with a message rather than silently destroying something

### Fixed: WORK Could Drop a Metal’s Other Adjectives, and Now Accepts Its Full Name
- Working a decorated metal like “sparkling green steel” by its base name alone (`WORK STEEL`) kept only “steel” and silently dropped “sparkling” and “green” from the finished item — a weaponsmith crafting a kite shield from that metal ended up with a plain steel shield. All of a metal’s adjectives are now preserved no matter which word you worked it by. `WORK` also now accepts a metal’s complete descriptive name (`WORK SPARKLING GREEN STEEL`), not just a single word from it — this also benefits `REINFORCE`, which shares the same lookup

### Fixed: Short, Pluralized Item Names Could Match the Wrong Item
- Typing a short item name ending in “s” (treated as a possible plural) could match a completely unrelated item if stripping the “s” left a very short leftover that happened to prefix another item’s name — `SELL BAS`, intended for a basinet, could instead sell a backpack, since “bas” minus its “s” is “ba,” which prefixes “backpack” too. This affected every item-targeting command, not just `SELL`. A minimum length now applies to that shortened form, so a coincidental two-letter overlap like this can no longer happen

### Fixed: Subway Control Panel’s Lights Never Updating
- Turning the knob on the panel in the Subway station-master’s office correctly cycled its internal state and said the lights changed, but looking at the panel always showed the same description no matter what you’d turned it to. Examining a room item now correctly runs the scripts that report on its current state, the same way examining a carried item already did

### Fixed: Subway Token Machine Taking Your Coins and Giving Nothing
- `PAY` at the Subway terminal’s token machine deducted your coins immediately but never followed through on the delayed “clunk” step that actually hands you the token

### Fixed: Deeplander Battleaxe Not For Sale
- The Blacksmith shop’s sign advertised a Deeplander Battleaxe for 500 Kragenmarks, but it referenced an item number nothing in the game actually defines, so it could never be bought under any name. It’s now a real, purchasable weapon

## v12.10.0 — 2026-09-17

A big batch today. Two psionic disciplines that never worked — `Strengthen Steel` and `Transform Steel` — are real now, and `Yshtarin’s Confounding Translocation` is implemented for the first time. Two new GM commands, `@freenum` and `@spawning`. The headline fix, though, is how monster magic and psionic resistance actually works: monsters were quietly getting two independent chances to shrug off a spell or psionic attack instead of one, which is not how the original game worked — that’s corrected everywhere it applied, across every offensive spell and psionic discipline in the game. Also: `Earth Wave` finally stuns and knocks down its target the way it was always documented to, snared creatures (webbed, tentacled, entangled) can’t recover from being knocked down or asleep until the snare itself wears off, and eight smaller bugs are fixed, including a room players could become stuck in forever.

### Changed: Monster Magic and Psionic Resistance
- Casting a spell or projecting a psionic discipline at a resistant monster used to roll twice: once for your own casting skill, and then — only if that already succeeded — a completely separate roll for whether the monster resisted. That gave resistant monsters an unintended second chance to avoid being affected. Resistance now works as a modifier on the single roll you already made, lowering your effective chance to affect that specific target, instead of being a second independent check stacked on top of it. This touches every direct-damage spell, Chain Lightning and Flaming Arrows’ per-target rolls, Siryx’s Terrible Tentacles, Slumber, Web, Sunray, Fear, Charm, Silence, Imprison, and offensive psionic disciplines

### New: Strengthen Steel Actually Does Something
- Mind over Matter discipline 7 has been a pure flavor no-op since it was added — activating it printed a message and nothing else happened. While maintained, it now grants a wielded steel weapon +15 to attack rating and +15 to its break-resistance in a weapon clash, matching its documented “more sturdy and sharper” effect. Both bonuses disappear the instant you deactivate it or stop wielding a steel weapon

### New: Transform Steel
- Mind over Matter discipline 13 (level 15) permanently transforms a wielded steel weapon into elkyri, randar, or truesteel and imprints a random +3 to +17 magical bonus — “the results are unpredictable,” per the discipline’s own description. Refuses anything that isn’t a wielded steel weapon, and refuses a weapon that’s already magically enchanted rather than overwriting it

### New: Yshtarin’s Confounding Translocation
- Spell 246 sends up to the first four creatures in your `TARGET` list to a marked room — the mirror image of Bend Space I: the caster stays put and the targets go instead. A moved creature loses track of what it was attacking. `CAST` at an unmarked or invalid mark number reports the target location as invalid rather than silently failing

### New GM Commands: @freenum and @spawning
- `@freenum <item|room|monster|noun|adj> [count]` lists unused numbers in a numbering space — gaps left by deleted content plus a safe unused block above the highest number in use — for finding a number to practice scripting with, instead of guessing
- `@spawning room off/on` or `@spawning world off/on` toggles natural monster spawning for the room you’re standing in, or the entire game, without affecting manually staged `@spawn`/`@genmon` monsters

### New: Earth Wave Stuns and Knocks Down Its Target
- Earth Wave was a plain damage spell with no secondary effect. A successful hit that doesn’t kill its target now also stuns it briefly and knocks it down, matching its documented “stuns, knocks down” description

### New: Snared Creatures Stay Down Until the Snare Clears
- A knocked-down or sleeping creature that’s also currently webbed, tentacled, or (for a player) entangled in roots no longer recovers on schedule — it stays down until the snare itself wears off, then recovers normally. Also fixed a related bug where a summoned or controlled creature specifically could scramble back to its feet while still snared, when an ordinary monster in the same situation already correctly couldn’t

### Fixed: TARGET Command’s Nine-Creature Cap
- `TARGET` capped out at 6 creatures; it’s meant to allow 9 (Chain Lightning, Flaming Arrows, and Siryx’s Terrible Tentacles all benefit), with only Yshtarin’s Confounding Translocation limited to the first four of whatever’s targeted

### Fixed: Bend Space I Bringing Along a Following Pet
- Bend Space I is documented to teleport only the caster, but it was also dragging along any summoned creature (elemental, familiar, guard) currently set to follow. Bend Space II, which is meant to bring your whole group and their pets, is unaffected

### Fixed: Bend Space II Leaving Group Members Behind
- A group member who was sitting, laying down, or still in round time got silently dropped from the group by Bend Space II instead of being swept along — that “can’t keep up” rule belongs to walking through an exit, not to a spell that teleports the whole group magically. Everyone in the group now arrives together regardless of position or round time

### Fixed: Speed (Psi 42) Surviving Logout
- Deactivating the Speed discipline by logging out instead of typing `PSI 42` again left its round-time-halving effect permanently active on the character, even though `PSI`’s own listing correctly showed it as inactive. Logging back in now correctly clears it, the same way Flight already does

### Fixed: GM Login/Logout Announcements
- A GM connecting or disconnecting normally — not hidden, not invisible — still triggered the global “has just entered/left the Realms” announcement every player sees. GMs log in and out frequently for admin work; that announcement (and the matching room-arrival line) is suppressed for them now, the same as it already was for a hidden or invisible GM

### Fixed: Overly Fast Spawning at the City Wall
- Room 1070’s spawn probabilities summed to 204% per check — a spectral lord and a rakshasha alone were each rolling at 94% and 68% — producing near-constant spawns. Rescaled down to roughly half that, matching how the same over-spawning bug was fixed elsewhere before

### Fixed: A Room Players Could Get Stuck in Forever
- The cloud room high above the ground — meant to toss you around briefly and then drop you back to earth — never actually triggered its own drop. Two separate script bugs conspired to keep it from ever firing; both are fixed, so the timed fall out of the clouds works again

### Fixed: SKIN Command Double Space
- `SKIN` reported skinning “a  ghoul” with a double space before the creature’s name, both to the skinner and to the room

## v12.9.0 — 2026-09-15

`Divine Blessing` is a real spell for the first time — it used to just print a message and do nothing. Also fixed four bugs: `DESCEND` refusing to land you when you’re already as low as you can fly, `SEARCH` letting you skip its own round-time delay entirely, Body Restoration claiming to heal someone who was already at full body points, and every bow, instrument, and piece of jewelry or woven cloth crafted outside a forge selling for a single copper no matter how it was made.

### New: Divine Blessing Actually Does Something
- `Divine Blessing` (spell 347) has sat in the spellbook as a pure no-op — casting it printed a flavor message and nothing else happened. The original documentation was mislabeled under a typo (“Diving Blessing”), which is likely why the effect was lost: it grants +3 ranks of weapon skill for one minute, on yourself or another player. Like True Aim, Haste, and Slow, recasting it doesn’t stack — it just resets the timer back to a full minute

### Fixed: DESCEND Refusing to Land at Your Lowest Altitude
- Flying and typing `DESCEND` somewhere with no lower point to drop to gave a flat “nowhere to descend” error. It now lands you instead, matching the documented behavior

### Fixed: SEARCH Ignoring Its Own Round Time
- `SEARCH` — scanning a room, searching a corpse, or searching an item — never actually checked whether you were still in round time before acting, letting it be spammed with no delay at all. It’s gated like every other timed action now

### Fixed: Body Restoration Claiming to Heal Someone Already at Full Health
- Casting Body Restoration I, II, or III on a target already at full body points still reported “healing X body points” — the roll happened, it just silently had nowhere to go. It now correctly says the target is already at full health instead of claiming a heal that never occurred

### Fixed: Crafted Bows, Jewelry, and Woven Goods Selling for a Single Copper
- Anything finished at a fletcher’s bench, jeweler’s workbench, or loom — bows, musical instruments, rings, cloth and leather goods — never had a real sell value set, so it always sold to a merchant for the shop’s absolute minimum regardless of how demanding it was to craft. These items are now priced the same way forged weapons and armor already were: scaled by the crafting skill required and capped at half the raw material’s own cost

## v12.8.0 — 2026-09-12

`Slow` now works as an offensive spell against monsters instead of only the caster, and two script commands that were parsed but never actually executed — one-time quest experience awards and a monster permanently changing the world on death — are wired up for the first time. Also fixed six bugs: skinned trophies always selling for a single copper piece regardless of their real value, a whole class of commands silently failing whenever you typed an item’s own displayed name back at it, `READ` ignoring the more common form of item script, `FLY` refusing to work in rooms explicitly built around flying, undead still vulnerable to poison and disease, and a long-unreachable dragon-tribute puzzle in the Deep Realms.

### New: Slow Can Target Monsters
- `Slow` (spell 211) could previously only be cast on yourself or another player — aiming it at a monster returned “You don’t see ‘X’ here” even when it was standing right there. It now works on any monster in the room, doubling its action interval the same way it already doubles a hasted player’s cast time

### New: AWARD and Monster-Death World-Flag Script Commands
- `AWARD`, used in 14 places across 11 script files for one-time quest experience (including 2,500 experience for first crossing the natural bridge deep in the Realms), was parsed but never actually executed. It now grants each reward exactly once per character, no matter how many times the triggering room is revisited
- A monster dying can now permanently flip a world flag its own script defines — used by Ming-K’Tuk, the great wyrm guarding the Deep Realms’ treasure hoard, so that killing him instead of appeasing him no longer leaves his lair’s script convinced he’s still alive forever after

### Fixed: Dragon Tribute Puzzle in the Deep Realms
- Giving Ming-K’Tuk a valuable tribute always accepted it, but his delayed follow-up — the roared question you’re supposed to flatter him about before he’ll let you pass — never actually fired, because handing an item to a monster wasn’t scheduling the script’s delayed continuation. The full tribute-then-flattery-then-passage sequence works end to end now

### Fixed: Skinned Trophies Always Selling for a Single Copper
- Skinning a dead monster never carried over the skin’s own script-defined value, so every trophy — a muldragun claw, a dragon scale, anything skinned from anything — fell back to a near-zero generic price when sold to a merchant, regardless of how valuable the source creature’s hide was meant to be

### Fixed: Commands Failing When You Type an Item’s Own Displayed Name
- Many items display with a built-in article — “some water,” for instance — but typing that back verbatim (`DRINK SOME WATER`) failed to match, since leading articles were only ever stripped before matching a monster or summoned creature’s name, not a general item’s. `GET`, `DRINK`, `EAT`, and every other item-targeting command strip them now too

### Fixed: READ Ignoring an Item’s Own Script
- `READ` only ever checked the less common of an item’s two script forms, silently skipping the other — used by 20 different readable items across 11 script files, including the ragged map in Menelian’s Cottage that’s supposed to reveal a hidden crevice’s location. Both forms fire correctly now, whether the item is still on the ground or already in your pack

### Fixed: FLY Blocked in Rooms Explicitly Built Around Flying
- `FLY` refused to work in any cave or indoor-tagged room, even ones built entirely around flying — like a Deep Realms waterfall chamber whose only way down is a pair of exits that can only ever be used while airborne. Rooms that define one of those exits now allow flight regardless of terrain

### Fixed: Undead Immune to Poison and Disease
- Undead already couldn’t bleed from wounds; they now can’t be poisoned or diseased either, matching the same “no living body to affect” logic, whether the source is a monster’s bite, weapon venom, or a trap

## v12.7.0 — 2026-09-10

New enchantment spell `Disenchant` and an optional reagent for `Siryx’s Terrible Tentacles`, plus a throttle on how fast Body Restoration and Body Destruction shift your alignment. Also fixed a bug that could strand a flying player in mid-air unable to descend, an experience exploit in weaponsmithing tin weapons, targeting commands for controlled/summoned creatures that only worked with a full name, and several spells that silently consumed their reagent with no “turns to dust” message to anyone.

### New: Disenchant
- Level 27 Enchantment spell that undoes whatever Enchantment I, II, or III put on a weapon, armor piece, or shield — removing the enchanted/ensorcelled/eldritch quality and its magic bonus
- Requires a spriggan finger as a reagent (skinned from a spriggan), consumed on cast
- Falls back to checking the item’s magic bonus directly on an item enchanted before it had a free adjective slot to record which quality it received

### New: Optional Worm Scale Reagent for Siryx’s Terrible Tentacles
- `PREPARE 134 WITH worm scale` is now an optional catalyst (skinned from the giant worm) that grants +25 to the resist-roll rating against every target the tentacles try to grab — casting without it works exactly as it always has

### Balance: Alignment Shifts From Healing/Harming Spells Throttled
- Body Restoration I/II/III nudging your alignment toward good was firing on every single cast, letting a caster reach a white aura in a tiny fraction of the effort it took in the original game. Now only a 1-in-100 chance per cast actually shifts it (and shows the “you sense the pleasure of the gods” message)
- Body Destruction I/II/III now has the same dark mirror — a 1-in-100 chance per successful cast to nudge alignment toward evil, with its own “you sense the displeasure of the gods” message
- Breath of Life is unaffected — reviving the dead is rare enough on its own to stay an unconditional +2

### Fixed: Flying Players Could Get Stranded Mid-Air
- `SIT`, `STAND`, `KNEEL`, and `LAY` while flying used to silently ground you with no warning — over open sky, with nothing to actually sit or lie on, leaving you unable to `DESCEND` and no clue why. They now refuse and tell you to `LAND` first
- Fleeing combat while flying no longer drops you out of the sky either — every other position still snaps back to standing on a successful flee, same as before

### Fixed: Tin Weapons Paying Wildly Inflated Experience
- Tin had no entry of its own in the weaponsmithing quality table, so it fell through to the tier meant for rare exotic metals like randar and elkyri — a tin cutlass could out-earn one forged from steel or even truesteel. Tin now sits appropriately below copper, both for experience reward and for how easy it is to work (it’s a very soft, low-melting-point metal, just a poor choice for an actual weapon)

### Fixed: COMMAND GUARD/FOLLOW/ATTACK Needing a Full Name
- Directing a controlled or summoned creature with `COMMAND GUARD`, `FOLLOW`, or `ATTACK` required typing a player’s entire name — every other targeting command in the game accepts an abbreviation. All three now do too

### Fixed: Spell Reagents Vanishing With No Message
- Casting a spell that consumes a physical reagent (Enchantment I/II/III, Disenchant, Charge Wand, the elemental summons, Breath of Life, Summon Spectral Warrior) removed the item with no acknowledgment at all in several cases, and even where a message existed for the caster, it never reached anyone else in the room. Both the caster and onlookers now see the reagent “turn to dust” as it’s absorbed into the spell, matching the original game’s wording

## v12.6.0 — 2026-09-09

The `HELP` command works like it originally did: typing `HELP` alone shows the real introductory text new players saw in the 1990s, and `HELP <topic>` now actually answers for all eight of the topics it always claimed to cover.

### New: HELP \<topic\> Actually Works
- `HELP` with no argument now shows the original 1996 introduction — how to talk to the world, how to speak out loud with `'`, the handful of commands every new player should know, and the list of deeper help topics — recovered from an actual session capture rather than a generic modern summary
- `HELP <topic>` now dispatches to one of eight topics — `combat`, `commanding`, `death`, `movement`, `psionics`, `spells`, `training`, `verbs` — each listing the commands that actually apply, instead of the topic name going nowhere
- `SHAREDXP` is listed under `HELP commanding` alongside the rest of the group commands, so it’s discoverable without already knowing it exists

## v12.5.0 — 2026-09-09

New: `SHAREDXP`, a group leader’s toggle that splits combat experience across the whole party instead of handing it all to whoever landed the kill. Also fixed several scripted-item bugs that silently dropped a script’s logic in favor of generic default behavior — an openable item with its own reaction to being opened (or closed), a container with its own “look in” description, and an item that rolls a value before checking it — plus a crafting XP bug that could pay out thousands of experience for a low-skill craft.

### New: SHAREDXP — Group Experience Sharing
- Only a group’s leader can toggle it, with `SHAREDXP`, and the whole group is told when it’s switched on or off
- While active, any kill by a group member splits that monster’s experience across everyone in the group who’s online, alive, and in the room when it dies — instead of the killer taking it all
- Grouping up grows the total pool of experience rather than just slicing the same amount thinner: each additional member adds 25% more (a 3-person group creates a 150% pool), so a party that kills faster together also earns more together
- Shares aren’t flat — each member’s cut is nudged up or down based on their level relative to the group’s average, modestly and within a cap, so a much lower-level member’s share is reduced but never negligible, and a much higher-level member’s is boosted but never doubled
- An extreme level gap across the group (say, a level 50 alongside a level 2) shrinks the pool itself substantially, to blunt power-leveling through grouping without making the low-level member’s own cut worthless
- A monster that dies from bleeding out (or any other damage-over-time death with no single decisive blow) now shares the same way, credited to whoever landed the last hit and their group — if `SHAREDXP` is off, that last hit alone gets full credit, same as before

### Fixed: OPEN and CLOSE Ignoring a Container’s Own Script
- An item that’s both a normal openable container *and* has its own scripted reaction to being opened — like the Grimoire of Chaos, which is supposed to flip to a random page and trigger one of thirteen different effects — always fell through to the plain “you open it” message instead, because the mechanical container logic took over before the script ever got a chance to run. Fixed for both `OPEN` and `CLOSE`
- Scripts that pause partway through (like the grimoire’s “the pages stop turning” beat before the effect resolves) weren’t being scheduled to actually finish, so the pause never resolved into anything

### Fixed: LOOK IN Ignoring a Container’s Custom Description
- A room item flagged as a container that also has its own custom “look in” text — a well with murky water below, a chest long since plundered down to a few centipedes — always showed the generic “it is empty” message instead of that text

### Fixed: Item Scripts Silently Dropping Setup Actions
- An item script that rolls or sets a value before checking it further down (like a pair of dice, which roll two values before `ROLL` reports them) had that setup step silently discarded during parsing on any item, so the check downstream always saw the value at zero. `ROLL DICE` now actually reports what came up

### Fixed: Crafting Experience Reward Reading the Wrong Item Field
- Crafting a sling (a skill-2 Wood Lore item) was paying out 8,660 experience — the reward formula was reading a missile weapon’s ammunition-item reference as if it were a difficulty rating. It now uses the same skill-level field the game already uses to decide whether you’re allowed to craft the item in the first place, so the reward actually matches the difficulty

## v12.4.0 — 2026-09-07

A first real alignment system: life-affirming spells nudge your alignment, `Aura Sense` reveals anyone’s as a color, and GMs can view and set it directly. Mass Protection now shares its defense bonus with your whole group instead of just the caster. Also fixed a cluster of Island pool/volcano-vent bugs (a broken cave exit, an unreliable breath-holding mechanic, a ritual that echoed its message three times, and a crank mistaken for a book), the Cult of Dahkahn’s temple pointing at the wrong building, and a gem-molding bug that could destroy a flawless gem’s quality instead of just polishing it.

### New: Alignment System
- Casting Body Restoration I, II, or III on a living target now nudges your alignment toward good, with the flavor line “You sense the pleasure of the gods as you call upon the forces of life.” Breath of Life does the same at double strength, since pulling someone back from death is the greater act
- `Aura Sense` (spell 404), previously a complete no-op, now reveals a target’s (or your own) alignment as a color, using the original scale: 50+ White, 10–49 Soft Yellow, 1–9 Green, -5–0 Blue, -99–-6 Deep Purple, -100 Pure Darkness
- GMs can now see a player’s alignment via `@edpl` and change it directly with `@set align <value>`

### Changed: Mass Protection Now Protects the Whole Group
- Mass Protection’s +25 defense used to apply only to the caster — it now extends to every other member of the caster’s group, leader or follower, whoever happens to be holding the spell
- It never stacks: if two group members each have their own copy active, the group still only gets +25 total
- The bonus is tracked live rather than copied onto each member, so leaving the group (either the beneficiary’s or the original caster’s) drops it immediately

### Fixed: Island Pool and Volcano Vent Scripts
- `GO SURFACE` always failed in the underwater pools — the surfacing script was checking a hidden placeholder item instead of the actual “surface” the command resolves to
- The “you can’t hold your breath” auto-surface only ever worked if you happened to say something out loud — it now fires reliably from any action that raises your held-breath count, regardless of whether you speak
- Speaking the volcano vent’s sacrifice phrase printed “You feel a rumbling.” three times instead of once, since the room script defines several near-duplicate triggers (for punctuation and a misspelling) that our engine was firing all at once
- Turning the volcano vent’s crank got swallowed as “You carefully turn the page” once the sacrifice ritual had touched it — the engine was mistaking the crank’s internal state value for a book’s page count

### Fixed: Cult of Dahkahn’s Temple Pointed at the Wrong Room
- Two conflicting registrations meant the org’s temple was either a nonexistent room or literally Shin-jah’s altar — training worked there but never counted toward Cult of Dahkahn rank. It now points at the Dark Cathedral chancel, where the training list and altar actually are

### Fixed: MOLD Could Destroy a Flawless Gem’s Quality
- Molding a flawless or perfect gem overwrote that adjective with “polished” instead of adding polish alongside it — only an actual flaw (chipped or cracked) should ever get replaced

## v12.3.0 — 2026-09-06

A big batch of fixes to verbs that silently failed against scripted items — BREAK, WIELD, and GO all now actually reach the item scripts written for them, fixing a breakable window, a huggable doll, and a bedroll/coffin you could get stuck inside with no way out. Also: GMs get two new broadcast tools (`@monitor`, `@sndroom`), `@snd` finally reaches the room instead of just the GM, and the admin Logs page can mark a player report as done.

### Fixed: BREAK, WIELD, and GO Ignoring Scripted Items
- `BREAK` was entirely stubbed out (“Object destruction coming soon”) and never checked an item’s own script — a wasteland window that was supposed to shatter open under a solid hit now actually breaks
- `WIELD` only ever ran a weapon’s or shield’s normal equip logic — a decorative item with its own scripted reaction to being wielded (like a child’s doll you can only brandish “like a shield”) now gets a chance to show that reaction instead of a flat “you can’t wield that”
- `GO` never checked your inventory at all — only the room floor — so a portable bedroll or coffin you carry and crawl inside of had no way to trigger its own “go” script. Fixed, along with a deeper parsing gap that made a script nested inside an item-state check (“while the bedroll is open, going in does X”) unreachable no matter what
- The bedroll and coffin scripts themselves also had a copy-pasted ordering bug: they recorded which room to return to *after* already moving you into the pocket room, so “go opening” from inside just sent you right back into the same tiny room. Both now save your original room first
- The `APPEAR` script command (used by items to change how you look to others, e.g. pants pulled down or a face stained with jelly) was broadcasting its text as a one-time room announcement instead of actually setting your appearance line — it now sets the persistent line, visible to anyone who examines you until it’s set again

### New: @monitor — Relay a Room’s Activity to a GM
- `@monitor <room#>` or `@monitor <player>` silently relays everything happening in that room back to the GM as `**`-prefixed lines, the same way a familiar’s `COMMAND WATCH WILL` already does — `@monitor off` stops it, and `@monitor` alone shows what you’re currently watching

### Fixed/New: @snd, @sndroom, and @announce
- `@snd <text>` used to only show the message to the GM who typed it — it now actually echoes to everyone else in the room too
- New `@sndroom <room#> <text>` echoes a line to every player in a room by number, even when the GM isn’t standing there
- `@announce 1 <msg>` no longer prefixes the message with “[Global Announcement]” — it sends the plain text as-is; mode 2 (mindlink) was also fixed to match its own documented usage, since the code had been checking for mode 0 instead

### Changed: Treasure Queue Drop Rate, Plus a GM Heads-Up
- The chance a GM-staged `@queue` item drops from a dying monster is now 10% (down from 20%), and every online GM now gets a `*HONK*` notice naming the player, the monster, and the item whenever one actually drops

### New: Admin Logs — Mark a Player Report Done
- The admin panel’s Logs page can now mark a Player Report as done, so resolved reports drop out of the default view instead of piling up next to ones that still need attention — a filter switches between Unresolved (default), Resolved, and All

### Fixed: Necromancy Backfire Hurt Living Casters
- Fumbling Turn Undead I/II or Destroy Undead I/II/III used to backfire and damage the caster even though these spells are only supposed to affect the undead in the first place — a living caster’s own fumble now just fizzles harmlessly, the same way the spell already fizzles against a living target

### Fixed: Disguised GMs Outed by Their Own Teleports
- `@go`, `@goplr`, and `@answer` used to announce a GM’s real name arriving/vanishing even while disguised — they now show the disguise (or wolf/mist/slime form) like every other movement in the game already does

### Fixed: Smelting and Mining Small Bugs
- `SMELT` only ever matched ore by its bare name — `smelt tin ore` or `smelt 2 ore` failed unless you dropped every adjective and ordinal and just typed `smelt ore`; adjectives and ordinals now work like they do everywhere else
- Mining kept letting you swing away at zero fatigue instead of telling you to rest

## v12.2.0 — 2026-09-04

GMs can now stage specific items to drop from monster kills. `@queue` and `@unqueue`, previously no-ops, now manage a real FIFO of items that can drop from any dying monster, and `@give` can target a monster in the room instead of only a player, silently loading it up to drop when it dies.

### New: @queue / @unqueue — a Real Treasure Drop Queue
- `@queue` by itself lists the treasure drop queue top to bottom; `@queue <item>` (with ordinal/adjective matching) pulls an item from the GM’s inventory, or failing that the ground, onto the bottom of the queue — `@queue top <item>` puts it on top instead
- Any monster that dies now has a chance to drop the item at the front of the queue, independent of its own treasure level — on a hit, that item is removed from the queue and appears among its remains
- `@unqueue <position>` pulls a specific queued item back into the GM’s inventory by the position shown in `@queue`’s listing; `@unqueue all` empties the whole queue back into the GM’s inventory

### New: @give Can Target a Monster
- `@give <item> to <monster>` (e.g. `@give steel dagger to goblin`) silently loads an item from the GM’s inventory into a monster in the room, instead of only working on players — the item drops onto the ground when that monster dies

## v12.1.0 — 2026-09-03

Weaponsmiths can now size up a weapon just by looking at it: EXAMINE reveals its non-magical crafting bonus, magical enchantment bonus, and Weapon Clash hardness rating to anyone with Weaponsmithing skill. Also fixed: the Keep’s Skull House was spawning monsters far too quickly.

### New: Weaponsmiths Can Appraise a Weapon’s True Bonuses
- Any character with Weaponsmithing skill now sees extra detail when they `EXAMINE` a weapon — its non-magical bonus, magical bonus, and hardness rating (the same figures that decide how well it fights and how likely it is to break in a Weapon Clash) — previously invisible to inspection no matter how skilled the smith

### Fixed: Skull House Spawning Monsters Too Quickly
- The lords, rakshashas, and inseviax haunting the Keep’s Skull House had a combined spawn chance of 110% per check — on average, more than one new monster was arriving every time the area was checked for spawns. Rescaled to a combined 44%, bringing its pace in line with the rest of the Keep grounds

## v12.0.0 — 2026-09-01

A major addition: craftable shields and armor. Five shield shapes in wood or metal, a full head-to-toe armor progression across four slots in metal, leather, and cloth, and a new `REINFORCE` command that lets a weaponsmith add metal plating to a shield a Wood Lore crafter already built — two different players can each do their part on the same item. Combat damage reduction now depends on which body part your armor actually covers, instead of any armor you wore reducing all damage everywhere. A real bug in crafting difficulty that affected every piece of armor in the game, not just the new ones, is fixed along the way.

### New: Shields (5 Shapes, Wood or Metal)
- Buckler, round, large round, kite, and tower shields, each craftable as a wood shield through Wood Lore at a fletcher’s workshop, or as a full-metal shield through Weaponsmithing at a forge — the same `CRAFT` then `WORK <metal>` (heat, hammer, quench, buff, finish) cycle every other forged item already uses, not a shortcut
- A new off-hand parrying weapon, the main-gauche, doubles as both a real weapon and a shield — its defense bonus scales with your Two-Weapon Style skill

### New Command: REINFORCE — Two-Crafter Shields
- `REINFORCE <shield> WITH <metal>`, usable at a forge with Weaponsmithing skill, adds metal plating to an existing wood shield — upgrading it in place rather than replacing it, so a Wood Lore crafter and a Weaponsmith can each contribute to the same shield

### New: Armor (Head, Torso, Hands, and Feet)
- A full metal armor progression: great helm, basinet, barbute, kettle hat, skull cap, diadem, and circlet for the head; full plate, brigandine, scale mail, and chain mail for the torso; matching gauntlets and boots for each, all craftable through Weaponsmithing
- A parallel leather and cloth progression — leather helmet and cap, boiled leather, layered hide, cloth gambeson, supple leather, and matching gloves and boots — craftable through Dyeing/Weaving instead
- Thigh guards, a new leg armor piece alongside the game’s existing shoulder guards
- All four armor slots can be worn together, and alongside a shield, stacking their protection

### Changed: Damage Reduction Now Depends on Where You’re Hit
- Previously, any armor you wore reduced damage from a hit anywhere on your body — a helmet was quietly reducing damage to your legs too. Now only armor that actually covers the body part struck reduces that hit: a helmet protects your head, gauntlets protect your hands, and your torso piece (which covers the body, back, arms, and legs as one combined piece, per how it’s actually built) protects everywhere else, with boots and thigh guards adding extra protection to your legs and shoulder guards adding extra protection to your back

### Fixed: Crafting Skill Requirements for Armor
- The skill level needed to craft any piece of armor was being calculated from the item’s weight, which was never actually correct — every armor item in the game, including ones that existed long before this update, has always carried its real required skill level in a separate field with no relationship to weight. Crafting requirements for armor across the game now use the correct value instead of a guess

## v11.37.0 — 2026-08-29

Gem sale prices have been reconstructed from real 1996 session data and are now locked in per gem with a bit of natural variation, instead of being recalculated from scratch every time. Also: Highlanders can mold any ungraded gem now (not just chipped/cracked ones), a botched mold properly devalues the gem, DYE recognizes adjectives and ordinals like every other item command, and a stealthed player reconnecting no longer announces their return.

### New: Gem Sale Values Reconstructed & Locked In
- Per-gem sale prices are rebuilt from roughly 50 actual “sell \<gem\>” transactions found across several 1996 session captures, where a plain diamond sold for 156 gold. The placeholder prices in place before this were squeezed into a narrow 40–800 copper range that badly undervalued precious stones — diamond and emerald in particular are now roughly 30–85 times their old value, with common stones like agate and citrine adjusted more modestly
- A gem’s sale value is now rolled once, when it’s created, and locked in from then on, instead of being recalculated fresh every time it’s appraised or sold — a diamond found today will still be worth the same amount whenever you eventually sell it
- Two gems of the same type, quality, and size no longer sell for identical amounts — each gem’s price now includes its own small random variance, matching how the same plain gem sold for visibly different prices from one sale to the next in the original captures

### Fixed: MOLD Limited to Chipped/Cracked Gems, and Not Devaluing a Botched Gem
- Highlanders using `MOLD` can now work any gem that isn’t already polished or damaged — plain gems and ones already flawless or perfect can be worked now too, not just chipped or cracked ones as before
- A botched mold attempt that leaves a gem “damaged” now actually halves its sale value as intended — previously the gem was marked damaged but kept selling for its full, undamaged price

### Fixed: DYE Couldn’t Match Adjectives or Ordinals
- `DYE <item> WITH <dye>` only ever matched a bare noun, so a foraged dye reagent like “a wart stem” couldn’t be referenced by its adjective — `dye cotton with wart stem` failed, and only `dye cotton with stem` worked. Both the item being dyed and the dye reagent now support adjectives and ordinals (“second”, “2”, etc.), matching how every other item-targeting command in the game already works

### Fixed: Hidden/Invisible Players Announcing Their Return on Reconnect
- A player who logs back in while Hidden, under the Invisibility or Phantom Form spells, or GM-invisible no longer broadcasts “X arrives.” to the room or “** X has just entered the Realms” globally — reconnecting only checked the GM invisibility/hidden flags before, not stealth or those spells

## v11.36.0 — 2026-08-28

A new invasive spell, Truename, is implemented for the first time. Also two real fixes: a melee critical hit that’s supposed to stun or knock down a monster wasn’t actually disabling it, and group-following now handles separation and incapacitated followers consistently across every way the group can travel.

### New Spell: Truename
- Every player now has a hidden, made-up truename — never their real first/last name — randomly rolled and guaranteed unique the first time anything actually needs it: casting `CAST TRUENAME` on them, or them summoning/controlling a creature or inscribing a sigil/glyph
- Cast on a summoned or controlled creature, Truename reveals its summoner’s truename; cast on an item you inscribed with a sigil/glyph spell (Imprison Rune, the three elemental Glyphs, Death Scythe), it reveals whoever inscribed it — neither can put up any resistance
- Cast on another player it’s an aggressive act: their normal mental defenses apply and can make the spell fail, unless they’ve `SUBMIT`ted first, which drops those defenses to almost nothing
- Casting with no target given (or explicitly on `me`/`self`) reveals your own truename, the same default-to-self behavior every other targeted spell already has

### Fixed: Stun/Knockdown From a Critical Hit Not Actually Applying
- An “Excellent Hit” that rolls to stun or knock a monster down showed “It is stunned!” correctly, but wrote the flag onto a disposable copy of the monster instead of the real one tracked by the world — so the monster was free to attack again on its very next turn as if nothing had happened, sometimes less than a second later

### Fixed: Group Following on Separation and Incapacitated Members
- Walking away from your leader through a door, gate, or other portal now breaks your membership in the group the same way walking through a plain compass exit already did — previously only directional movement cleared the link, so a portal could leave you still “following” someone you’d actually left behind
- When a group leader moves, a follower who’s in round time, sitting, laying down, or stunned is no longer silently dragged along with the rest of the group — they’re left behind and dropped from the group instead. If that was the leader’s last follower, the leader stops being a group leader entirely

## v11.35.1 — 2026-08-27

Three bug fixes: repeatable lockpicking on the second+ matching lock in a room, the Teeth of Shartan stream-jump only working once per direction, and several precious gems selling for a single copper instead of their real value.

### Fixed: Lockpicking a Second Matching Lock
- `PICK <ordinal> <item> WITH LOCKPICK` (e.g. `pick 2 door` or `pick second door`) now actually disambiguates when a room has more than one item with the same name — it previously ignored the ordinal entirely and reported “You don’t see anything locked here,” matching how `OPEN`/`CLOSE`/`UNLOCK` already number matching items in the room

### Fixed: “Jump Boulder” Only Working Once Per Direction
- The stream-crossing boulders at the Teeth of Shartan (all four seasonal variants) let you jump across once each way, then refused with “You can’t do that” on a repeat attempt — a one-line flag reset in the room’s own script was being silently dropped by the script loader instead of running before every jump attempt, so the flag set by a successful jump never cleared. Bare reset lines like this one, sitting outside any `IF` block in a room’s script, are now loaded and re-run on every interaction the way the original scripts intended

### Fixed: Several Gems Selling for 1 Copper
- Sapphire, diamond, garnet, sardonyx, agate, coral, and topaz previously sold to shops for a single copper regardless of quality, since the code used to price gems only recognized ones flagged as spell reagents — and several genuine gem types were never flagged that way in the original data. Gem pricing is now keyed to the same item-number range (99–121) the Wrecked Ship’s offering bowl already uses to recognize a valid gem, so every gem type — including quartz/crystal stones, which were also underpriced — now sells for its proper value

## v11.35.0 — 2026-08-25

A pass focused on summoned and controlled creatures: buff spells like Strength, Haste, and Globe of Protection can now actually be cast on a familiar, elemental, or dominated undead instead of failing outright, and two real bugs affecting controlled undead are fixed along the way.

### New Feature: Buff Spells on Summoned & Controlled Creatures
- Strength I/II/III, Agility I/II/III, Haste, Mystic Armor, and the whole Globe of Protection family (Globe of Protection I/II, Mass Protection, Spectral Shield, Ride the Lightning) previously only ever looked for a named player as a target, so casting one on a familiar, elemental, animated skeleton, or dominated undead failed with “You don’t see ‘X’ here.” Any caster — not just the pet’s own summoner — can now target it by name
- These aren’t cosmetic: Strength and Agility scale a creature’s effective attack and defense rating the same way they scale a player’s, Mystic Armor and the Globe of Protection spells add real stacking defense, and Haste actually halves the creature’s action-tick interval so a hasted pet acts and attacks twice as often — mirroring how Haste halves a player’s round time
- Buffing someone else’s pet now sends them a heads-up (“X casts Strength on your skeleton”), the same way `COMMAND GUARD` already notifies a guarded player

### Fixed: Controlled Undead Defending Hostile Monsters Instead of Fighting Them
- Undead dominated with Control Undead I/II kept their species’ original wild instinct to guard other monster types it’s scripted to protect — a phantom caretaker, for example, guards zombies and zovembies in the wild. Bound to a player, it kept doing that instead of obeying its new master: bringing a controlled caretaker near a hostile zovembie made it “guard” the zovembie and intercept attacks meant for it, instead of fighting alongside its controller. A summoned or controlled creature no longer acts as an independent guardian of its own kind

### Fixed: Item-Cast Reconstruction Healing the Living
- Reconstruction (337) correctly fizzles when `CAST` on a living target — it’s an undead-only heal — but an item imprinted with it (like a rub-activated idol) used a separate item-triggered spell path that never had the same check, so the idol healed any player who rubbed it. Body Restoration I/II/III had the matching gap in the other direction: cast from an item on an undead target, it healed them instead of searing their undead flesh as damage. Both now match how the spells behave when cast normally

## v11.34.0 — 2026-08-14

A security hardening pass to protect from bots — account registration is now guarded by a CAPTCHA challenge and a much tighter rate limit. Also, a mistyped or missing spell target no longer wastes the spell.

### New: Account Registration Security
- Creating an account now requires solving a CAPTCHA challenge first — invisible to almost everyone, only escalating to an interactive checkbox or puzzle when the traffic pattern looks automated, so real players notice nothing while scripted mass-account creation gets blocked outright
- The registration rate limit dropped from 5 attempts/minute to 3/hour per IP address — the old limit still allowed hundreds of accounts (and hundreds of verification emails) to be created in well under an hour from a single source

### Fixed: Spells No Longer Wasted on a Bad Target
- Casting a spell at a target that doesn’t exist — a typo, or nothing specified at all — used to permanently consume the spell’s mana and clear it from being `PREPARE`d, forcing a full re-prepare even though nothing was ever actually cast. Every spell that resolves a named target (damage spells, buffs, item-enchantment spells, summon/command spells, and more) now leaves the prepared spell and its mana untouched when the target can’t be found — a spell that genuinely gets cast but fails for some other reason (resisted, already in effect, wrong creature type) still costs the attempt as before

## v11.33.0 — 2026-08-12

The big one: room containers like a player’s home chest can now be marked persistent, so their contents actually survive a server restart instead of resetting every time. Also a new GM registry command, and a wide sweep of fixes to summoned creatures, item commands, and combat/social text.

### New Feature: Persistent Containers (Home Chests)
- Room containers are normally rebuilt from scratch every restart, along with the rest of the world — anything a player `PUT` in a chest, bin, or other container was always gone by the next reboot. A specific room-item instance can now be flagged durable with `ITEMBIT19=1` on its `ITEM` line (or `EQUAL ITEMBIT19 1` from a script with that item in context), and its contents are written through to the database on every `PUT`/`GET` and restored on boot
- `@rdata` now shows a room item’s `ItemBits` and, for containers, their current contents — neither was visible anywhere before, unlike `@iexamine` for inventory items

### New GM Command: @intnum3
- INTNUM3 is a per-player identifier scripts lean on constantly — theft prevention, guildmaster/packleader checks in WOLFHOME.SCR, item ownership stamps — that GM staff used to track by hand on an external document. `@intnum3` alone lists every assignment, `@intnum3 <plr>` checks one player’s value, and `@intnum3 <plr> <val>` assigns one, refusing if another player already has it (the shared GM sentinel value of 1 is exempt)

### Fixed: Summoned Creatures
- `COMMAND GUARD ME` was silently cancelling the creature’s follow order the moment guard mode turned on, so it got left behind the next time you moved. Guarding is now purely additive — a creature can guard several people at once — and never touches who it’s following; only `COMMAND FOLLOW` changes that
- Bend Space I and II teleported the caster (and group) but left any summoned or guarding creature stranded in the room they left, since neither spell went through the normal movement code that brings summons along
- Bend Space I and II also broke Hidden and Invisibility on cast — per the original session logs, a hidden or invisible caster stays unseen straight through the teleport

### Fixed: Item Commands
- `PUT` could never find a wielded weapon, so any scripted interaction that requires one to be actively held (like a weaponsmith’s whetstone) always failed with “You aren’t carrying that,” even right after wielding it
- `LEARN` counted scrolls differently than `EXAMINE` does, so `learn 12 scroll` could target a different scroll than `exam 12 scroll` just showed, or fail outright if anything non-scroll shared the name
- `@give` and `@take` silently deleted the item entirely when the target player wasn’t currently online, instead of failing — both now refuse with a clear “not online” error

### Fixes
- `WHISTLE` always said “You whistle innocently,” whether you meant it that way or not, and had no way to whistle at someone — it’s now a plain “You whistle,” with target support (`whistle mortif`)
- `PULL` now supports targeting another player (`pull mortif`) the same way other physical emotes do, respecting `AVOID`
- Invisibility told the caster “You fade from sight” — which they obviously can’t see happen to themselves — instead of “You feel a tingling sensation,” per the original session logs
- `LOOK`/`EXAMINE` on a player listed health and injuries before worn equipment; the original always lists equipment (and any active magical effects) first
- The `H>` prompt indicator only ever showed for a manual `HIDE`, not for Invisibility or Phantom Form, even though both make you just as hard to see

## v11.32.0 — 2026-08-11

Wizard’s Armor, Spell Shield, and Cloak Mind were all secretly the same generic physical defense buff instead of what MAGIC.TXT actually describes them as — fixing Cloak Mind meant building an entirely new monster psionic attack system from scratch, since nothing existed for “psi resistance” to defend against. Also: a working Charge Wand, five new “sigil” item-warding trap spells, and two small player-command fixes.

### New Feature: Monster Psionic Attacks
- Monster psi disciplines (`DISCIPLINE`/`PSIUSE`/`PSISKILL`/`PSI` in the script data) were parsed and stored but only ever consumed for a monster’s own passive defense bonus — no monster had ever actually attacked a player with one. Monsters with an offensive discipline (Kinetic Thrust, Pyrokinetics, Cryokinetics, Electrify, Psychic Blast/Crush/Terror/Pain) now windup and strike with it, mirroring the existing monster-spellcasting system (windup announcement, then a hit-flavor line and damage, dodgeable, and disruptable by taking a hit mid-charge, same as an interrupted spell)
- Per GMSCRIPT.DOC (“powerful psionic creatures should have well over 100 PSISKILL so as to defeat a player’s natural resistance to psionics”), every player now has a baseline psi-resistance rating even without Cloak Mind — a weak-to-moderate psionic monster can be shrugged off some of the time; a strong one usually still connects

### Fixed Spells: Wizard’s Armor, Spell Shield, Cloak Mind
- Wizard’s Armor (229), Spell Shield (234), and Cloak Mind (235) were all routed through the same generic timed-defense-buff spell handler as Mystic Armor or Globe of Protection, quietly granting flat physical `+N defense` — none of which matches what MAGIC.TXT documents for any of the three
- Wizard’s Armor now does what it’s actually supposed to: “no spell disruption” — while active, taking a hit no longer breaks a prepared spell. Casting it surrounds the target in “a yellow curtain of light” instead of granting any defense bonus
- Spell Shield now grants “+25 magic resistance” — a real, live mechanic: a 25% chance to deflect a spell a monster casts at the target outright, on top of the monster’s own cast-chance and the target’s dodge. Casting it wraps the target in “an antimagical field”
- Cloak Mind now grants “+25 psi resistance” against the new monster psionic attacks above, and deliberately shows no line at all in a `LOOK`’s active-effects list — the point of this one is that the target doesn’t look warded. Casting it on someone else just says “X seems changed.”
- All three still stack duration the same way every other timed defense spell does — 20 minutes per cast, recasting before it expires extends it, capped at 4 hours

### New Spell: Charge Wand
- Charge Wand (243, “Recharge Wand” per MAGIC.TXT) requires mandrake root to prepare, then `CAST <item>` tops up the charges on a wand, rod, or trinket you’re holding or wearing that already has a spell imprinted on it — it can’t create a new magic item from scratch. Drains every point of your remaining mana and converts it into charges (remaining mana ÷ the imprinted spell’s own mana cost, rounded up), added on top of whatever charges are already there

### New Spells: Sigil Traps
- Imprison Rune (227), Thunder Glyph (125), Inferno Glyph (124), Ice Glyph (126), and Death Scythe (322) all share the same mechanic per MAGIC.TXT: “cast upon an item, and attuned to the next person who touches them.” Cast on an item you’re holding, wearing, or one lying on the ground — including a door or gate. The first person to `TOUCH` it (or try to `OPEN` it while closed or locked) silently claims it and can go on handling it freely; anyone else springs the trap once and it disarms
- Imprison Rune traps the intruder in a 5-minute force bubble (spell 231). Thunder/Inferno/Ice Glyph hit them with their own damage type at double the roll. Death Scythe conjures a Spectral Sword (345) at 5× damage

### Fixes
- `SIT` from a laying position said “You sit down.” like sitting from standing; it now says “You sit up.”
- `ACT` always used your real name even while in Slime Form, Mist Form, or a Wolfling’s wolf form — `ACT sloshes around` in Slime Form now shows “A slime sloshes around” instead of your name, matching how every other identity-masking system in the game already works

## v11.31.0 — 2026-08-10

Two new spells — a stealthy Paranoia curse and a wild-card Dispel Lesser Magic — a spellcraft rebalance to help new casters actually land their first spells, and fixes to potion drinking, herb-based spell effects, and a room-routing bug that sent departure messages to the wrong room when climbing or slipping through a portal.

### New Spells: Paranoia and Dispel Lesser Magic
- Paranoia (226) settles a 20-minute creeping unease on its target: each real minute there’s a 50% chance of one of thirteen unsettling echoes (“Something taps you on the shoulder.”, “You hear a scream behind you!”, and eleven others). Unlike every other spell, casting it at someone else produces no message at all to the caster, the target, or the room — the target has no way of knowing they’ve been cursed — and it doesn’t reveal a Hidden or Invisible caster the way casting normally does
- Dispel Lesser Magic (401) strips one random active timed magical effect from its target — a buff, debuff, entangle, or defense spell — regardless of how much duration it had left. “A deep red light twinkles around” the target is visible to everyone in the room, but which effect actually faded, if any, is only reported privately to the target

### Balance: Easier Early Spellcasting
- The base spellcraft success chance for casting any spell rose from 25% to 50% (still Empathy/10 + spellcraft skill×5% on top, capped at 95%) — a fresh level-1 caster with one rank of Spellcraft was landing spells far less often than a level-1 weapon user landed hits, since casters needed a favorable roll just to get the spell off at all, on top of a second resist roll most offensive spells still require against the target

### Fixes
- Drinking a bound potion showed backwards, over-verbose messages (“You drink the potion from a reeking bottle engraved with the writing, ‘Night Vision’.”) — now reads “You drink the reeking potion from a bottle.”, naming the potion itself rather than misapplying its color to the container, and no longer spoils which spell is bound to it
- `EAT` on a spell-bearing herb or mushroom only worked for a couple of hardcoded cases (Cure Poison, Body Restoration I) — anything else printed “[Spell #318 effect coming soon.]” even though the spell was fully implemented, e.g. a riyong mushroom’s Body Restoration III. `EAT` now falls back to the same general spell-effect handling `DRINK` already used for potions
- `CLIMB` (e.g. the ladder in room 615) and slipping through a scripted portal via `STEAL` were sending the “X leaves.” departure message to the destination room instead of the room being left, because the room being left was read after the room-changing script had already moved the player. Both now capture the departing room first, matching how ordinary `GO` movement already handled it

## v11.30.0 — 2026-08-07

Guild rank now advances automatically as you train, matching the original game’s formula, and Menelian’s traveling cottage in Grymwood no longer leaves a stray duplicate of itself behind as it moves.

### New Feature: Automatic Guild Rank Advancement
- Training a skill at your guild’s own trainer now raises your rank in that guild by the build points spent — reconstructed from the original documented formula (“ORGRANK = build points spent on training within guild”), so rank reflects how much you’ve actually trained there since joining, not your overall character level or experience
- Automatic advancement caps at rank 199 (just below High Master/High Priest) — crossing into the 200+ tier still requires a GM’s `@rank`
- Training now prints “You are now rank X in the guild.” whenever your displayed rank ticks up, matching the original flavor text

### Fixes
- Menelian’s Cottage (Grymwood, room 128) periodically appeared with an extra duplicate of itself and a phantom door, and could leave the cottage stranded behind with its door missing after moving on — caused by the cottage’s placement script reusing an item-ref slot the room already used for its own ground-snow decoration. Placing an item at a ref that’s already occupied now replaces what was there instead of stacking a duplicate, which also protects any other script that reuses a room’s item refs the same way

## v11.29.0 — 2026-08-06

A fully working Scry spell — remote room visions and a two-step eye-of-scrying ritual — a brand new guild area, the Order of the Skull, and a wide sweep of commands that were leaking a disguised player’s real name instead of their persona.

### New Spell: Scry
- Scry (215) has two uses. `CAST <mark#>` shows a brief vision of a marked room — full occupants, items, and exits, respecting `BRIEF` — without moving you there or revealing who’s watching. Anyone actually standing in the scried room gets “You have a brief yet distinct feeling that you are being watched.”
- The second use targets a carried eye (termite eye, sharkhor eye, newt eye, werewolf eye, etc.): prepare and cast Rite of Preparation (412, a newly added spell — previously mislabeled “Bloodsight” in the spell list and otherwise unimplemented) at the eye to turn it cloudy, then Scry at the cloudy eye to make it translucent. `LOOK IN` a translucent eye reveals the room where the last player death occurred, with a 1-in-10 chance per look that the eye crumbles to dust and is destroyed
- Fixed a message-ordering bug shared with Bend Space (213/222): the success-roll line was being inserted before “You gesture into the air.” instead of after it

### New Guild: Order of the Skull
- A new necromancy-focused organization (28) with a six-room tower reachable via a climbable path from the cemetery trail near Fayd. Entry to the tower is restricted to Order members; inside, a Great Hall connects to a reagent shop (Anatomicatory), a skill-training room (Chamber of Necrology), and a hidden passage behind the throne — concentrate, pull, tap, or push the onyx skull revealed by looking behind the throne — leading up to the Skull Lord’s private chamber

### Disguise Command
- `DISGUISE` with no field value now always lists the basic personas (trader, merchant, lawkeeper, priest, commoner, beggar) available regardless of skill rank, instead of only mentioning them below rank 10
- `DISGUISE LIST <#>` now shows Hairstyle and Haircolor as separate lines instead of blending them into one “Hair” entry
- New `DISGUISE CLEAR <#>` command resets a save slot back to empty

### Fixes: Disguise Identity Leaks
- A disguised player’s real name was leaking through instead of their persona in a large number of places: tending wounds (self, another player, or a monster/corpse); wielding or unwielding a weapon or shield, wearing or removing an item; eating, drinking, picking up, and dropping items (including coins and `GET ALL` / `GET ALL FROM` a container); following, holding, leaving, and disbanding a group, including the equivalent messages when a group leader or member disconnects; picking a lock; opening, closing, locking, and unlocking a door, gate, or chest; giving an item or money to another player; every step of crafting (mining, smelting/forging, weaving, dyeing, foraging, alchemy, jewelcraft, engraving); buying from a shop; dumping a container; and logging out (“X fades from the Realms”)
- Leveling up no longer announces a disguised player’s advancement to the room at all, rather than exposing their real name

## v11.28.0 — 2026-08-05

Nine new Enchantment spells — temporary creature domination and a full suite of concealment/transformation magic — plus a rebuilt `HEALTH` command, a new `FATIGUE` command, an optional reagent bonus reconstructed from a 1990s session log, and a handful of smaller fixes.

### New Spells: Temporary Domination
- Command (205), Domination I (206), and Domination II (214) let a caster temporarily seize control of a living (non-undead) creature — the same `COMMAND`-verb control (`FOLLOW`, `STAY`, `GUARD`, `ATTACK`, `LOOK`, `SAY`, `BEGONE`) already granted by a summoned elemental, but only for the duration, and only against a creature at or below a body-point ceiling based on its *current* body points, not its max: Command 50 BP for 2 minutes, Domination I 100 BP for 20 minutes, Domination II 200 BP for 20 minutes. Recasting on the same creature resets the duration instead of stacking it

### New Spells: Concealment & Transformation
- Invisibility (225), Mass Invisibility (212), and Phantom Form (248) have no duration — the target stays concealed until they smile, cast a spell (preparing doesn’t count, only the actual cast), or attack. All three can be cast on yourself or a named player in the room
- See Hidden (405) reveals every hidden, invisible, or phantom-formed player in the room: a hidden player shows as “You see something.”, an invisible one shows their real effective name (accounting for wolf form, disguise, mist, or slime), and Phantom Form always shows as “You see a shimmering grey form.” no matter who it actually is
- Mist Form (232) and Slime Form (245) — self-only transforms lasting until `TRANSFORM` reverts them (15-second roundtime, “You congeal back into your body.”). Both block attacking, casting, wearing/removing items, all speech, and every emote (even a smile), and lock your inventory entirely (`INVENTORY`, `GET`, `DROP`, `PUT`, `WIELD`, `UNWIELD`, `GIVE`) so there’s no way to fall back on a magic item instead. Mist Form grants full immunity to physical and magical damage — including bleeding, poison, and disease ticks — plus flight/ascend/descend and passing through closed doors and gates unless they’re flagged `SEALED`; Slime Form only reduces damage taken to 10% (also skipping DOT ticks entirely), with none of the mobility perks. Successfully casting either immediately breaks Plant Snare and Imprison. Both appear in room listings and `EXAMINE` as “some mist” / “a slime” instead of your real name

### Commands
- `HEALTH` rewritten to match the original session-capture format (“You have 256/276 body points.”) and now shows your current bleeding/poison/disease drain rate per minute when any are active, in place of the old vague “You are moderately wounded.” descriptor
- New `FATIGUE` command (abbreviates to `FAT`) — shows current/max Mana, Psi, and Fatigue, split out of `HEALTH` into its own command

### Items & Reagents
- A moonstone can now be used as an optional `PREPARE` reagent for any spell that doesn’t already require one of its own — `PREPARE <spell> WITH <moonstone>` consumes it for a +25% bonus to that cast’s success chance, reconstructed from a 1990s player session log where a moonstone was used exactly this way alongside Freezing Sphere

### Fixes
- A room’s bare-verb `SMELL`/`LISTEN`/`SNIFF` script (no item targeted) was silently dropping everything after a `PLREVENT`/`CONTPLREVENT` pause — e.g. the wolfling cave’s cleansing-smoke effect on `SMELL` never actually applied its second half
- `OPEN` on a container with something inside now says “You open the chest, revealing a rusty key and 12 gold.” instead of stopping at “You open the chest.”
- Call Storm (501) now has its own flavor text — “Energy crackles between \<caster\>’s fingertips and then lances skyward.” — instead of a generic “gestures and casts” line
- New kill-flavor line for heat/fire deaths: “Internal organs stew in their own juices. Throw another on the barbe’, mate.”

## v11.27.1 — 2026-08-04

Three long-dormant Enchantment spells brought to life — Silence, Imprison, and Identify previously did nothing when cast — plus a couple of message-ordering fixes.

### New Enchantment Spells
- Silence (219) now silences its target for a full minute — unable to `SAY`/`'`/`YELL`/`SING`/`RECITE`, and since casting requires speech, unable to cast any spell either. Recasting resets the duration back to a full minute rather than stacking. A monster flagged `SILENCEIGNORE` (one that casts via hand movements or symbols instead of speech, per GMSCRIPT.DOC) keeps casting right through it
- Imprison (231) now traps its target in a blue force bubble for 5 minutes — they can’t attack anyone or cast any spell, including on themselves, so they can’t even attempt to dispel it. The bubble works both ways: nobody else can land a physical or spell attack on the trapped target either. An imprisoned monster now shows “(imprisoned)” in room listings
- Identify (228) now names the exact spell bound to an item and how many charges remain, the precise counterpart to Detect Magic’s vaguer “glows a soft blue” hint

### Fixes
- Fixed a handful of spells whose cast messages named the target (Call Meteor, Chain Lightning, Siryx’s Terrible Tentacles) showing the success roll before the opening gesture line instead of after it, unlike every other targeted spell

## v11.27.0 — 2026-08-03

A full Disguise skill — compose and wear a persona that blends seamlessly into the world’s common NPCs or a custom identity of your own, with matching movement, speech, and appearance across every system that shows a player’s identity — a merged room-look format matching the original game’s style, and a sweep of related bugs including a Rakes guild password that echoed but never actually let anyone in.

### New Feature: Disguise Skill
- `DISGUISE <slot> <field> <value>` composes a saved persona — each Disguise rank unlocks a new field to change (1 name, 2 gender, 3 hair color/style, 4 skin/eye color, 5 age, 6 apparent strength, 7 height, 8 weight, 9 race, 10 a custom name of your own instead of a generic one), and grants a save slot every 2 ranks, up to 5
- `DISGUISE APPLY <slot>` puts one on (`[Round: 30 sec]`), `DISGUISE REMOVE` takes it off (`[Round: 15 sec]`), and `DISGUISE LIST` / `DISGUISE LIST <slot>` review what you’ve composed. Bare `DISGUISE` shows instructions tailored to your current rank
- Below rank 10, your name is restricted to one of the game’s generic town NPC types — commoner, trader, merchant, lawkeeper, beggar — and disguised as one of those, you’re indistinguishable from the real thing: your name gets the same article (“a commoner”), you wander and arrive the way those NPCs do (“A commoner wanders south.” / “has arrived.” instead of “goes/arrives”), and `EXAMINE` skips straight to the description with no “You look at X.” opener, exactly like looking at a genuine NPC of that type
- At rank 10+, a custom persona can have a full first and last name; room lists and broadcasts still show only the first name (matching how real players are always shown), with the full name reserved for `EXAMINE`
- Your disguised identity now carries through everywhere another player would see or address you — `LOOK`, `WHISPER`, `CONTACT`, a summoned creature’s `COMMAND FOLLOW/GUARD/ATTACK`, and the broadcast text for combat, movement, spellcasting, and emotes all resolve and display your disguise instead of your real name while one is worn

### Room Look Overhaul
- Players and monsters/NPCs are now listed together in one sentence (“You see Rion, a lawkeeper and a commoner.”) instead of two separate ones, and players are shown by first name only — matching the original 1990s format — instead of full name plus race
- Items now get their own sentence (“A table is here.”, “A glass and some wine are here.”), matching the original session-capture wording, instead of being folded into a “You see…” sentence
- The five generic town NPC types (commoner, trader, merchant, lawkeeper, beggar) now get a full randomly-rolled appearance on `LOOK` — race, gender, age, build, eyes, skin, hair — generated once per spawn and cached so looking twice shows the same person, instead of a flat “You see a trader.”

### Fixes
- A script using `MOVEGROUP` from an `IFSAY` block — like the Rakes guild hall’s “admit” password at Geoffrey the wine cellar doorman — played out its flavor text but never actually relocated anyone, since `IFSAY` scripts only ever handled the single-player `MOVE` action, not `MOVEGROUP`
- Items placed inside a container via `PUT`/`NEWPUT` (e.g. wine poured into a glass) were showing up as loose items lying on the floor in addition to being in the container — room `LOOK` now correctly hides anything marked as being inside another item
- An `@invis` GM moving through a portal (door, arch, etc.) leaked their real name in the “X goes through Y.”/“X arrives.” messages, even though speech already correctly anonymized them as “Something says…” — portal movement is now silent for concealed players, matching ordinary directional movement
- `@editem` couldn’t set any multi-word field except `tail` — added `examinedesc` (sets an item archetype’s `EXAMINE` text) and generalized the multi-word-value parsing instead of hardcoding it to one field
- `EXAMINE` on an item with a custom examine description showed a redundant “You look at your X.” line before the description; it now prints just the description, matching the original game
- `APPEARANCE` with no argument now clears your custom appearance line instead of just redisplaying it — there was previously no other way to reset it
- `@speech <name>` with no verb phrase now resets that player’s speech pattern to default, instead of failing with a usage error

## v11.26.0 — 2026-08-02

A new Highlander gem-molding ability, casting a spell now actually breaks stealth and invisibility the way preparing one doesn’t, monsters that flee combat properly release everyone who was fighting them, and a wide sweep of command-abbreviation and script-hijack bugs uncovered while chasing down a single quirky item.

### New Highlander Ability
- `MOLD <gem>` — Highlanders can now work a chipped or cracked gem into a polished one (e.g. `MOLD chipped diamond` or `MOLD 3 diamond`), raising its resale value. There’s always some risk of botching it: a fresh Highlander has a 20% chance to ruin the gem instead (leaving it permanently damaged and un-moldable), dropping 1% per level to a permanent floor of 1%. A successful mold raises the gem’s value by 25% at level 1, up to a 5%-per-level bonus capped at a 100% increase. A gem already polished, or already ruined from a prior botch, can’t be molded again

### Spellcasting & Stealth
- Casting a spell (not preparing one) now reveals a hidden or invisible caster — the spoken words and hand gestures give away your position the moment you actually cast, even on a failed or fumbled attempt. `PREPARE` stays silent by comparison: onlookers just see “Something prepares a spell.” while you remain concealed right up until the cast itself
- Invisibility (225) now actually breaks when you cast any other spell, on top of already breaking when you attack
- An invisible player who `SMILE`s now fades back into view — the expression gives away your position even when the rest of you shouldn’t be seen. Psionic powers still don’t reveal you, since they’re worked by thought alone

### Combat
- When a monster flees combat, everyone who was fighting it is now automatically taken out of combat, the same way it already worked on a kill — previously only the monster’s own target reference was cleared, leaving players stuck thinking they were still fighting a monster that had already left the room

### Command & Script Engine Fixes
- Fixed a latent bug in command-abbreviation resolution: nine verbs (`TOUCH`, `THINK`, `TAP`, `RUB`, `RECITE`, `PROMPT`, `FULL`, `DEPART`, `BRIEF`) were accidentally listed twice in the verb registry, which made the ambiguity checker think a verb collided with itself and refuse to resolve any abbreviation of it at all
- `APPRAISE` was missing from the verb registry entirely, so no abbreviation of it could ever resolve (only the fully-typed word worked, by coincidence)
- Several genuinely ambiguous short abbreviations now resolve to their more commonly used meaning instead of failing outright: `mas→MASSAGE`, `pro→PROJECT`, `sel→SELL`, `fli→FLIP`, `poi→POINT`, `sla→SLAP`, `hea→HEADSHAKE`, `lea→LEAN`, `sni→SNIFF`, `cur→CURSE`, and `dro→DROP` — each runner-up verb (e.g. MASTER, POISON, SLAY, HEALTH) stays reachable with one more letter or the full word
- `WEAR`, `OPEN`, `CLOSE`, `LATCH`, `SELL`, `APPRAISE`, and `PROJECT` all shared the same bug: each rejected a target as mechanically ineligible (not wearable, not a container, no shop nearby, no psi discipline prepared, etc.) before ever giving that item’s own scripted reaction a chance to run — so an item designed to hijack one of these verbs unconditionally (regardless of its own type or the caster’s state) could never actually do so. All seven now check for that scripted reaction first

## v11.25.0 — 2026-07-31

Wolfling wolf-form transformation is now a real disguise — a generated animal description, growled/snarled speech, wolf-only emotes, and an anonymized “a wolf” everywhere a transformed wolf acts — plus prepared spells can now actually be interrupted by taking a hit (with Wizard’s Armor as the ward against it), and Nyraine’s long-isolated coastal farmlands are finally reachable.

### Wolf Form (Wolfling)
- `LOOK` at a transformed Wolfling now generates a real wolf description (age, gender, eye color, and fur color mirroring hair color, or black if bald in human form), followed by health and any wounds — described with paw/foreleg/hind-leg wording — instead of showing their real name, race, build, or any GM-set `@line1-3` text. Worn gear and a custom `APPEARANCE` line are hidden too, all reverting to the normal human description (custom or generated) the moment they transform back
- Speech becomes growls (`SAY`) and snarls (`EXCLAIM`) while transformed — asking a question still reads as “asks”. A custom speech adverb is ignored while in wolf form
- A transformed wolf’s identity is now hidden as “a wolf” everywhere they act, not just when looked at or spoken to — `SAY`/`WHISPER`/`YELL`, every emote, combat round messages, arriving/leaving a room (including logging in already transformed), `SIT`/`STAND`/`KNEEL`/`LAY`, and `%N` in room/item scripts. `THINK` and `CONTACT` still use the real name, since telepathy isn’t something anyone could physically observe
- New wolf-only emotes: `WAG` (wags tail, overriding the human finger-wag), `SNIFF`, `LICK` (a sloppy lick, replacing the human kiss/lick-all-over-body variants), `SHAKE`, and `POUNCE` (around, or on a target)
- `LOOK WOLF`, and ordinals like `LOOK 2 WOLF`/`LOOK 3 WOLF`, now correctly disambiguate multiple wolves sharing a room — which also surfaced and fixed a pre-existing bug where the online-player list was built by ranging over a Go map (randomized order on every call), making any “2nd/3rd match” targeting unstable from one command to the next

### Spellcasting
- A fumbled spell cast (backfire) now lands on a random body part with a real recorded wound, same as any other hit, instead of just subtracting body points with nothing to show for it
- Prepared spells can now actually be interrupted — taking a hit while mid-cast (from a monster’s attack, special attack, spell, or an item trap) breaks the prepared spell, mirroring the `NONDISRUPTABLE` flag already documented for disrupting a monster’s own casting. Wizard’s Armor (229) is the ward against it — while active, taking a hit no longer breaks a prepared spell
- Added a missing electrical kill-flavor line: “Spectacular charge electrolyzes water in body.”

### World Connections
- Fixed a duplicate room-numbering bug in Nyraine (all four seasonal script variants) where three separate content passes had silently overwritten each other’s room definitions, leaving the Farmlands area — and a Rocky Trail/Outcropping/Troll Caves connection — completely unreachable, with several nearby exits pointing at the wrong rooms. Renumbered the orphaned Farmlands rooms onto free room numbers and reconnected them via City Gates and the coastal Lower Rocky Trail, so the route from the Inner Sea dock through Nyraine to its farmlands finally works end to end

### GM Tools
- `@set` can now edit a player’s `EYECOLOR`, `HAIRCOLOR`, `HAIRSTYLE`, and `SKINCOLOR` — previously any string value was rejected before the command even checked which field you’d named, so only numeric fields could be set at all. Values are validated against the same fixed choice lists used at character creation, and `@edpl`’s summary now also shows age/height/weight/eyes/skin/hair

## v11.24.0 — 2026-07-30

WEAR and REMOVE now actually run item scripts, three previously-inert Conjuration spells imbue weapons with a temporary elemental crit, and Thunder Call gets its own flavor text plus a missing electrical kill message.

### Script Engine
- `WEAR` and `REMOVE` never ran any item scripts at all — unlike `GET`, `EAT`, `GO`, and `TOUCH`, which already did. Every `IFPREVERB WEAR/REMOVE` and `IFVERB WEAR/REMOVE` block in the world was silently dead code — cursed items that should resist removal, a heartstone that should graft onto your chest, and a stairway in Idemmu Ag that checks whether you’re wearing boots all did nothing. Both commands now run the same script hooks the rest of the verb set already does

### New Spells
- Storm Blade, Inferno Blade, and Winter Blade (135/136/137) previously did nothing at all when cast. They now imbue your wielded weapon with a temporary elemental crit — 20% chance per hit for 1–20 bonus damage, lasting 20 minutes (recasting extends the timer, 4-hour cap) — the same crit mechanic already used by ore-forged elemental weapons. A weapon that already carries a crit of its own (forged-in element or a slayer bonus) can’t accept the spell
- If the weapon has a free adjective slot, it’s marked fiery/icy/electric for the duration and reverts automatically when the spell fades; the imbue lives on the weapon itself, so wielding, unwielding, or giving it to someone else carries the effect along for whatever time is left

### Combat & Spell Messages
- Thunder Call (116) had no unique flavor text of its own; it now beckons a bolt of lightning down from the storm clouds and deals burn damage, matching the original session-capture wording
- Added a missing kill-flavor line for deaths by electrical damage (both a direct spell hit and a weapon elemental crit), and fixed the same gap for fire-crit kills
- Damage spells were broadcasting the exact damage number, wound severity, and body part to everyone in the room, not just the caster — bystanders now see the same vague damage tier (e.g. “Awesome damage.”) melee combat already shows them

## v11.23.0 — 2026-07-29

Karhad’s mountain trail now connects both directions, and the long-isolated coastal city of Nyraine is reachable for the first time — a new sailing ship ferries passengers across the Inner Sea between the two.

### World Connections
- The Teeth of Shartan trail up to Karhad previously only worked in one direction (down to the Inner Sea region, at the dam) — a matching climbable trail back up now exists on the Inner Sea side, so the route works both ways
- Nyraine, a fully-built coastal city that had no connection at all to the rest of the world, is now reachable by sea — Captain Aldous Kestrel and the crew of the *Windrunner* run passage between the Inner Sea docks and Nyraine’s own dock for two silver each way. `PAY` the captain, then board via the gangplank (`GO`) — the ship makes way, cuts across open water, and arrives on the far shore over the course of the crossing

## v11.22.0 — 2026-07-28

A wide sweep of container bugs (carried and worn alike, coins included), a crafting-economy fix so forged items are actually worth something, several scripted verbs that were silently ignoring an item’s own response, and two revived Druidic spells.

### Containers
- `GET <coins> FROM <container>` never recognized a coin pile sitting inside a container (it’s stored differently than a normal item) and reported “You don’t see that in there.” even when `LOOK IN` showed coins present; `GET ALL FROM` was worse — it silently destroyed any coins in the container instead of adding them to your purse. Both fixed
- Worn containers (backpack, waistpack, etc.) were invisible to `OPEN`, `CLOSE`, `PUT … IN`, `GET … FROM`, and `GET ALL FROM` — only a container held in your hands or lying on the ground worked. All five now check worn items too, with the same priority a carried container already had
- Fixed container capacity to match the documented meaning of the two fields: `INTERIOR` caps the total `VOLUME` of everything held inside (it was being used as a raw item-count limit), and a container’s own `VOLUME` governs whether a single item is too bulky to fit through the opening at all (it was being used as the total capacity) — a backpack meant to hold 40 volume of gear was actually being capped by its own small carried-bulk value

### Crafting & Treasure Economy
- A weapon or piece of armor forged via Weaponsmithing never had a resale value set at all — it sold for about a copper no matter what metal it was made from. Now scales with the skill level required to craft it, capped at half of whatever the raw metal actually cost to buy, so cheap material can’t be laundered into profit by forging and reselling it
- Monster-dropped jewelry, spell scrolls, and potions found in treasure containers had the identical problem — no value was ever set on them at all (only gems were handled correctly). All three now scale with treasure level and, for jewelry and potions, the power of any bound spell

### Combat
- `REPAIR` always checked the first inventory item matching the name and gave up immediately if that particular one wasn’t damaged — carrying two items sharing a name (one damaged, one not) meant the damaged one could never be reached and repaired. It now keeps looking until it finds an actually damaged match
- Weapon Clash now factors in Agility (+1 resistance per 5 points) alongside the weapon’s own weight/adjective-based hardness, giving more agile characters better odds of keeping their own weapon undamaged

### Spells
- Sunray (Druidic 519) was dealing heat damage; it’s supposed to stun instead — it now blinds and stuns the target for 4–6 seconds (two 1–2 rolls plus 2)
- Claw Growth (Druidic 518) previously had no effect at all when cast. It now grows natural claws usable whenever you aren’t wielding a weapon — damage and Natural Weapons skill bonus matching a real claw weapon — lasting 20 minutes (re-casting while active adds another 20 minutes rather than resetting the timer), castable on yourself only, and immune to weapon-clash damage since there’s no physical weapon to break

### Script Engine & Commands
- `%A` (the capitalized item name/article, e.g. “The happy fun ball”) was never expanded in scripts and printed literally — only lowercase `%a` worked
- `DRINK` and `EAT` only ever considered items typed FOOD/LIQUID/LIQCONTAINER, so an item with its own scripted response to one of these verbs but no matching type could never trigger it. Both now fall back to the same generic script dispatch `PUSH`/`PUNCH`/`TURN` already use
- `FLIP` only ever checked items lying in the room, never anything you were carrying
- `ACTBRIEF` now reflects each viewer’s own preference — previously the *actor’s* own ACTBRIEF setting controlled whether everyone else in the room saw parentheses around their `ACT` message, instead of each player’s own toggle
- Fixed the “Experience Points until next Build Point” counter reading roughly 1000 XP higher than it should for most of a level, due to a mismatched starting-build-point baseline in the calculation

### New GM Commands
- `@stat <name>` and `@skill <name>` — show a GM the exact same `STATUS` page or `SKILLS` list a player would see themselves, for any player by name (online or offline)

## v11.21.0 — 2026-07-27

Two core script-interpreter bugs fixed that silently no-op’d across dozens of scripts game-wide, a magic/psi resistance overhaul affecting roughly half the bestiary, several dyeing and weaving fixes, and spell fumbles now actually do something instead of just fizzling.

### Script Interpreter
- `IFVAR X == N` (double equals) always evaluated false — only single `=` was recognized as equality. Silently broke every script using the typo’d double-equals form, including a Keep gate password puzzle and an orb-tap stairway reveal that could never trigger no matter how correctly you performed the sequence
- `EQUAL`/`ADD`/`SUB ITEMADJ1-3` were a total no-op — the interpreter had no write case for that variable prefix at all (reads worked, writes silently did nothing). Affects roughly 300 script lines across the world; discovered via a sap reagent that was supposed to change state when lit but never did

### Magic & Psionics
- Monster `RESIST`/`PSIRESIST` values (a rating stat, same family as Attack/Defense, ranging into the thousands) were being compared directly against a 0–99 roll — any monster with a resist rating of 100 or higher (roughly 48% of the bestiary) auto-resisted every spell and psi attack, unconditionally, regardless of caster skill
- Resistance now scales against the caster’s own rating (Spellcraft skill + Empathy for spells, Psionics/school skill + Willpower for psi), using the same rating-vs-rating formula melee ToHit already uses — no monster is unhittable, no resist is a sure thing
- A fumbled spell cast (roll of 100) previously just printed “the spell backfires” and did nothing at all. It now actually backfires: damage spells hurt the caster instead of the intended monster, and heal/defense/buff spells land on the caster instead of whoever else they were aimed at

### Dyeing & Weaving
- `DYE` only ever searched your own inventory for the material — cloth left soaking in a dye cauldron (the normal, intended way to dye raw material) couldn’t be found at all. Now checks inventory, then the floor, then any container in the room
- `DYE` now requires an actual cauldron present in the room, not just a room flagged as a loom — most loom rooms (player housing, guild halls) never had one, so dyeing was usable in places it originally shouldn’t have been
- Fixed the dye-color field mapping — a two-word color like “inky black” was showing the wrong first word (e.g. “amazing black” instead of “inky black”) because the modifier adjective was being read from the wrong item field
- Crafting a garment from cloth dyed with a two-word color (e.g. “olive green cotton”) dropped the second word — the crafted item now keeps both

### Combat & Items
- A monster knocked down while simultaneously fighting a summoned creature or acting as a guardian never announced getting back up — it silently recovered with no message, unlike a knockdown during ordinary player combat
- `GET ALL` could scoop up items that are scripted to be un-gettable (e.g. a shop’s price-list manuscript that’s meant to always refuse `GET`) — it never ran an item’s `GET` script at all, only the single-item `GET` command did. `GET ALL` now respects the same scripted protections
- Doubled the coin yield from the random lootable containers (chests, strongboxes) monsters drop

### World Time
- “It is dawn” used to cover a 3-hour window, but scripted dawn events (like a fountain’s once-a-day effect) trigger on one exact hour in the middle of it — so two-thirds of the time you saw “dawn” displayed, the actual triggering hour hadn’t arrived yet. Narrowed “dawn” to the exact hour scripts check

## v11.20.0 — 2026-07-26

New room-aware `ADVICE` subcommands for newcomers and crafters, HELP/ADVICE pointed out right at character creation, and two real bugs fixed — a parser bug mangling every flush-left price sign in the game, and an enchantment spell that was erasing a store-bought item’s material adjective.

### ADVICE Command
- `ADVICE HINTS` — 21 newcomer tips covering guild locations, the Test Tunnels, banking, healing-by-resting, death penalties, and where to find player-made maps (Facebook group and PDF)
- `ADVICE CRAFTING` is now room-aware — away from a workshop it lists the Foundry, Crafter’s Guild, Bowyer & Fletcher, and New Havarth Mining Company (room 394); standing in one of them, a master crafter steps up with trade-specific tips (smelting/forging/repair at the forge, pelt/hide prep and dyeing at the loom, Wood Lore and carving at the fletcher’s bench, tool/purity tips at the mining shop), including a reminder that a forge or loom doubles as a jeweler’s bench for `ENCRUST`, `INLAY`, `INSET`, and `ENGRAVE`
- New characters are now told “Type HELP for a full list of commands, or ADVICE to get some tips for getting started.” right after character creation, on the web client, telnet, and SSH alike

### Fixes
- Fixed a script-parser bug where a flush-left price list or sign (no leading whitespace or blank lines to mark it as a table) was joined into a single unreadable run-on line instead of keeping its columns — affects every sign of this style across the world, discovered via the hat shop’s price sign in Fayd
- Enchantment I/II/III (202/203/204) were destroying a store-bought item’s material adjective (e.g. casting Enchantment I on leather armor produced “enchanted armor”, losing “leather”) — the spell now fills the first empty adjective slot instead of shifting all three, which was clobbering the variety adjective that purchased items carry in the third slot

## v11.19.0 — 2026-07-24

Five long-dormant Druidic spells brought to life — Camouflage, Call Storm, Disperse Storm, Plant Snare, and Freedom either did nothing or worked incorrectly before today — plus a full `SNEAK` rework and a new hurricane knockdown mechanic.

### Stealth
- Camouflage (Druidic 521) now works — grants +10 effective Stealth skill for 20 minutes, castable on yourself or another player; additional casts extend the duration instead of stacking the bonus, same as other timed buffs
- `SNEAK` rebuilt: ordinary movement now always reveals a hidden player — `SNEAK` is required to attempt staying hidden while moving. The roll is now Stealth + Agility/10 + Quickness/10, made harder by the highest Perception/5 among any players already in the room you’re moving into, and now takes a 2-second round (1 second under Haste) instead of being instant

### Weather & Storms
- Call Storm (Druidic 501) and Disperse Storm (502) previously had no effect at all when cast; Call Storm now intensifies your region’s weather one step toward Hurricane, Disperse Storm calms it one step toward Sunny — both require being outdoors
- Call Lightning (503) now requires being outdoors in at least Heavy Rain to cast, matching the original spell notes, instead of striking regardless of weather
- New Hurricane knockdown — standing, sitting, or kneeling players caught outdoors in Hurricane-force winds now have a chance each minute to be knocked to the ground; heavier and more agile characters are safer
- Resist Weather (506) now actually does something — a 20-minute buff (castable on yourself or another player) that cancels both the new Hurricane knockdown and the weather-based to-hit penalty from severe weather

### Movement-Restricting Spells
- Plant Snare (500) previously had no effect when cast; now entangles another player in grasping roots and vines outdoors, blocking their movement until it wears off or is removed
- Freedom (505) previously didn’t work — the potion/scroll version cleared an unrelated flag, and the spell itself had no live effect at all when cast; it now removes one active movement-restricting spell (e.g. Plant Snare) from the target, chosen at random if more than one is active
- New Repel Plants (509) and Repel Plants and Webs (510) — 20-minute immunity buffs, extending on recast like other timed buffs; the first blocks Plant Snare, the second blocks Plant Snare and Web alike
- Carapace (511) could previously be cast on other players or creatures like every other defense spell; fixed to caster-only, matching the original spell notes

## v11.18.0 — 2026-07-23

Monsters can now actually cast their scripted spells for the first time, special attacks got several real bug fixes (target selection, missing flavor text, damage over-reduction, duplicated messages), and container commands now check what’s in your hands before what’s on the ground.

### Monster Spellcasting
- Monsters configured with a spell (`SPELLUSE`/`SPELL`/`MANA`) can now actually cast it — this was fully parsed but never once invoked, so no monster in the game had ever cast a spell despite plenty being configured to
- Full prepare-then-cast sequence: a windup announcement, then the spell’s own cast time (matching how player spellcasting works), then a gesture line and the spell’s normal hit-flavor text and damage — landing a hit on the monster while it’s mid-cast disrupts the spell (unless the monster is flagged unable to be disrupted), and its mana is spent regardless
- Spell damage to a player now uses the same dice, damage type, and mitigations (armor, elemental vulnerability/resistance, shields) as when a player casts that same spell at a monster — it no longer also stacks the melee hit-location reduction (20–40% for a limb or hand) on top, which was driving it far below what the spell should do

### Monster Special Attacks
- A monster’s special attack (and spell cast) now targets a random player currently fighting it, or a random player in the room if nobody currently is — instead of being limited to whichever single player it happened to be locked onto in melee
- A monster’s limited number of special-attack uses is now actually enforced — previously unlimited regardless of how the monster was configured
- Fixed a parser bug where a special attack’s own scripted flavor text was never loaded due to a stray typo’d token, so every special attack fell back to a generic “uses a special attack” line instead of things like “turns one head to face you and looses a gout of flame!”
- Onlookers in the room now see a simplified outcome (e.g. “… Minor damage.”) when a special attack or spell lands — previously only the target’s own private message showed that anything happened at all
- The player actually hit no longer sees their own special attack or spell hit described twice — once in private detail, once again in the public room recap — the room broadcast now leaves out whoever already got the private version

### Container Commands
- `OPEN`/`CLOSE`/`LOCK`/`UNLOCK`/lockpicking now check what you’re holding before checking the room — previously all of these checked room items first, so holding a locked chest with an identical one lying nearby meant these commands would act on the one on the ground instead of the one in your hands (use `my <item>` to force your own inventory explicitly, same as before)
- `LOCK`/`UNLOCK` previously couldn’t affect anything in your inventory at all, only room items — a keyed chest you were holding could never be locked or unlocked without first setting it down

## v11.17.0 — 2026-07-22

Doors and gates now stay in sync from both sides of the doorway, a fully rebuilt ITEMBIT flag system that was silently broken since launch, two revived-from-the-dead spells, and a new GM item-duplication command.

### Doors & Locks
- `LOCK`/`UNLOCK` now number room items the same way `OPEN`/`CLOSE`/`TAP` already did — previously they only counted the *lockable* items in a room when resolving an ordinal, so in a room with two doors, “`lock 2 door`” could silently miss the door you meant
- Locking, unlocking, opening, or closing a door or gate now mirrors the same change onto its paired door in the room on the other side, and echoes an ambient message there (“You see a door open.”, “You hear a door lock.”, etc.) — previously the two sides could fall out of sync (unlock from one side, walk through, lock from the other — the first side stayed unlocked)
- `LOCK`/`UNLOCK`/`OPEN`/`CLOSE` now also announce to your own room (“Chandra locks a door.”) — previously these actions were completely silent to anyone standing right there
- `LOCK` now refuses to lock an open door (“You must close \<door\> first.”) instead of silently forcing it closed for you
- `@rdata` now shows full per-item detail (adjectives, VAL1–5, state, flags) for every item in a room instead of just its name — the only way to actually verify a door and its key share the same lock code (VAL3)

### ITEMBIT System
- ITEMBIT0–19, the per-item boolean flag system documented in the GM Manual, was never actually wired up — reading a flag collided with VAL4 (which has its own unrelated meaning), and writing one (`EQUAL ITEMBIT#`) was a silent no-op everywhere. Both now work correctly, backed by a real dedicated field
- This fixes any script that was quietly relying on an ITEMBIT check and getting nothing — including the Crimson Band ring’s War Room access check at the tapestry in the Hallway of Warriors
- `@editem` can now set `itembit0` through `itembit19` on an item

### Spells
- Pyrotechnics (Conjuration 141) was incorrectly dealing damage; it’s now the harmless fireworks display it always should have been — casting it launches a volley that, over the following minute, treats every outdoor player in your region to one of 12 random firework displays roughly every 15 seconds
- Mindlink (403) didn’t do anything when cast; it now grants `THINK` (telepathic speech) for an hour, the same effect as eating a thesnia leaf or drinking a thesnia potion — castable on yourself or another player in the room

### New GM Command
- `@dupe [#] <item>` — duplicates an item you’re wielding, wearing, or carrying and places the copy on the ground, including its adjectives, VAL1–5, sharpness, hardness modifier, item bits, and (for an open container) its contents

### Fixes
- `IFCARRY` now also checks worn items, not just loose inventory — scripts gating on a worn ring, amulet, etc. were failing even when you had the item on
- Fixed a data bug across five Fayd script files (the base room plus all four seasonal variants) where a value edit (`VAL3=1234`) glued directly against a trailing comment with no space was silently dropped by the parser — also removed a stray duplicate, unedited copy of the same door item left over in the spring script

## v11.16.0 — 2026-07-21

Unconsciousness replaces instant death at 0 body points, a real monster stun/knockdown system, a game-wide fix for delayed scripted sequences silently dropping, and a new reroll charm for legacy characters.

### Death & Unconsciousness
- Dropping to exactly 0 body points from a weapon hit or spell now knocks you unconscious (laying down, out cold, same presentation as being put to sleep) instead of killing you outright — only a hit that would drive you below 0 is lethal
- Bleeding, poison, and disease never kill outright on the tick that first drops you to 0 — you’re knocked unconscious instead, giving someone else a chance to `TEND` you or cast Body Restoration/Cure Poison/Cure Disease. If the same condition is still active and you’re still unconscious at 0 when it ticks again, that tick is lethal
- Unconscious players are locked out of every command except `LOOK`/`WHO`/`QUIT`/`STATUS`/`HEALTH`/`HELP` — no speech, no actions — until healed or natural regeneration brings them back above 0
- Room listings and `LOOK AT` now correctly show “(unconscious)”/“is unconscious” instead of lumping it in with ordinary “lying down”

### Monster Combat AI
- Stun from an excellent hit is now a real timed status (3–6 seconds) that actually blocks a monster from attacking, moving, or fleeing for its whole duration — previously it only skipped exactly one action before immediately resuming as normal
- New Knocked Down status — an alternative outcome to stun on an excellent hit (50/50 split) — costs a monster one full turn getting back on its feet before it can act again, the same way a monster already had to stand up after being put to sleep
- A monster that bleeds to death (or dies to Siryx’s Terrible Tentacles) now correctly awards kill experience to whoever landed the last hit, split evenly among their group if they’re grouped with others who aren’t hidden or invisible — previously a bleed-out death awarded no experience to anyone at all

### Fixes
- Fixed a game-wide bug where a scripted item’s delayed continuation (`PLREVENT`/`CONTPLREVENT`) was silently dropped after the first pause in `RUB`/`TAP`/`TOUCH`/`PULL`/`PUSH`/`TURN`/`SEARCH`/`DIG`, `WORK`, `EAT`, `DRINK`, `FLIP`, `READ`, `GET`, `LOOK`/`EXAMINE`, `GO`, `STEAL`, and `CLIMB` — any multi-stage scripted sequence using a delay (e.g. a fountain’s dawn-triggered dance) would just stop after its first line
- Fixed a related bug where putting an item into a scripted container/device (e.g. a garment-finishing contraption) discarded the item’s own adjectives/values before its script ran, breaking material checks, and where the device’s success branch never actually ran because it wasn’t recognized as having “handled” the action
- Weather no longer shows Snow Flurries through Blizzard during Summer while reporting a warm temperature — the weather-type simulation previously had no concept of season at all
- Keys no longer falsely open a lock that was never assigned a code — both sides defaulting to an unset value no longer counts as a match

### Drakin
- Drakin now have natural scale armor that strengthens with level (since they can’t wear armor at all), and take 25% additional damage from heat and cold — both documented racial traits that were never actually implemented

### New Item
- Reroll charm — a GM-distributed item for characters created before stat rerolling existed at character creation. `RUB` it to preview a fresh set of stats as many times as you like, then `CONCENTRATE` on it to lock them in; it crumbles to dust once used

### GM Tools
- `@set` can now edit a player’s `AGE`, `HEIGHT`, and `WEIGHT`, and can now write any named script variable (e.g. quest flags) the same way `@peek` could already read them

### New Characters
- Starting gold raised from 5 to 20

## v11.15.0 — 2026-07-20

Character creation overhaul — age, eye/hair/skin appearance, and a reroll-until-you’re-happy stat system, plus `STAT` and `EXAMINE` rebuilt to match the original game’s output, and a scripting bug fix.

### Character Creation
- New characters now roll Age, and choose Eye Color, Skin Color, Hair Style and Hair Color from wide fantasy-appropriate lists (12 hair styles including a Bald option that skips hair color entirely) — available on the web client, telnet, and SSH
- Stats (Strength/Agility/Quickness/Constitution/Perception/Willpower/Empathy) plus Height/Weight/Age can now be rerolled as many times as you like before locking them in, matching the original game’s “reroll until you like what you see” character creation — still respects each race’s stat ranges
- Fixed a telnet/SSH bug where choosing “1) Male” at character creation actually created a Female character internally, and choosing “2) Female” was rejected outright
- Existing characters created before this update are seamlessly backfilled with a rolled Age and appearance the next time they log in

### STAT & EXAMINE
- `STAT` reformatted to match the original layout, verified against a 1996 session capture — guild rank/title shown first (e.g. “You are a high master of the Crimson Band.”), followed by Name/Race/Gender, Level, build points, stats grouped Quickness/Constitution/Strength/Agility and Willpower/Perception/Empathy, an Age/Height/Weight/Load line, and (previously missing entirely) Body Points/Mana/Psi/Fatigue
- `EXAMINE` (on yourself or another player) now builds a real physical description from your stats and appearance choices, e.g. “You see Shirla Rennay, a young female aelfen. She is tall, light weight and robust. She has blue eyes, fair skin and long, flowing golden blond hair.” — height/weight/build descriptors are derived from your actual stats rather than being random

### Fixes
- Fixed item scripts with a verb-gated block nested inside a value-setting `IFVAR` tree (like the lens of worlds’ `SHOWROOM` preview) firing on *any* interaction instead of only the intended verb — e.g. `POINT`ing at the lens no longer shows the hidden room it’s only supposed to reveal on `EXAMINE`
- Other players now see “\<name\> incants a spell.” when someone prepares a spell, instead of “begins preparing a spell”
- Rorin’s Fire now has its own cast flavor text (a wave of red and orange flame that hisses and constricts “like a snake”) instead of the generic fire-bolt message

## v11.14.0 — 2026-07-19

Combat hit messages rebuilt with proper single-word severity descriptors and special elemental death flavor text, plus a batch of crash and quest-blocking fixes and a new weapon hardness modifier for GMs.

### Combat Messages
- Hit messages now show a proper single capitalized severity word (e.g. “Minor burn to right leg. [17 Damage]”) instead of the multi-word, lowercase persistent-wound vocabulary meant for the `HEALTH` command (e.g. “slightly lacerated slash to right arm”)
- Severity is now correctly based on damage as a percentage of the target’s max body points, matching original session-log evidence — the same raw damage number can be “Minor” against a tough monster and “Ghastly” against a weak one
- A killing blow from a cold or heat spell (or a weapon’s elemental crit) now shows a special description of the death itself — e.g. “Chilly body barrage solidifies muscle tissue.” or “Dazzling explosive display carbonizes bones and flesh.” — in place of the normal severity/damage line, matching original wording

### GM Tools
- `@additem <archetype#> [val1=N] … [adj1=N] …` now accepts value/adjective modifiers when spawning an item, so quest items that depend on a specific starting value (e.g. the lens of worlds) can actually be set up correctly for testing
- Weapons now have a GM-editable hardness modifier — `@editem <weapon> hardnessmod <N>` adjusts Weapon Clash break-resistance on top of the weapon’s normal weight/adjective-based value, visible via `@iexamine`, and survives being dropped and picked back up

### Fixes
- Fixed a crash-causing bug where a portal script using `MOVEGROUP` (rather than `MOVE`) — like the hidden hole on Island — never actually moved anyone through, and could misroute its departure message to the wrong room
- Fixed a server crash when stepping through a secret passage revealed by `KNOCK`ing (e.g. the Matriarch Tree basement) — the passage script removing itself from the room after use could invalidate the room’s item list out from under the very next line of code
- Invisible GMs no longer leak a script’s room broadcast (e.g. a portal’s custom “goes through” text) to other players when triggering it — matching the silence already applied to ordinary movement
- Fixed `GET` on an item with a non-blocking script (e.g. the lens of worlds’ first-touch binding ceremony, or a cursed item’s magical backlash) never actually completing the pickup, and fixed a related crash when such a script destroys the item itself as a side effect (e.g. “crumbles into dust”)
- `FLIP` now checks an item’s own script before requiring it be physically flippable — fixes puzzle items like a combination-lock knob that only responds to a room-defined `FLIP` script
- When a monster dies, every player who was fighting it is now taken out of combat automatically — previously only the player who landed the killing blow was disengaged, leaving groupmates (or anyone who died to bleed-out rather than a direct hit) stuck needing to manually `RETREAT`
- Wounds on undead players and monsters no longer bleed — they have no blood to lose

## v11.13.0 — 2026-07-17

New weapon specialization system, and a full hit-location damage & wound-tracking overhaul — real bleeding, per-wound severity descriptions, and TEND reworked to heal wound-by-wound.

### New Commands
- `SPECIALIZE <weapon>` — spend build points to specialize in the weapon you’re wielding (Crushing, Edged, Drakin, Polearms, or Thrown Weapons only; requires 10+ skill in that weapon’s category). Each of the 5 ranks reduces the fatigue cost of attacking with that weapon by 1 (never below 1) and shifts 5% of your hit-location odds toward the head and body. `SPECIALIZE` alone lists your current specializations, and they now appear in the `SKILLS` table

### Combat & Wounds
- Hits now land on a specific body part with real consequences — strikes to the head, body, or back deal full damage; arms and legs take 40%; hands, paws, and tails take 20% — derived from an original session-log analysis
- Wounds are now individually tracked on both players and monsters, each described with vocabulary matching the weapon that caused it — slash (nicked → cut → lacerated → gashed), puncture (pricked → stabbed → punctured → gored), crush (scuffed → bruised → battered → crushed → ruptured), and burn (singed → scorched → burned → charred) for heat/cold/electric damage — 12 severity levels each
- Slash and puncture wounds beyond the mildest two levels now cause real bleeding — body points drain away once a minute until the wound is treated, and can kill if ignored long enough, for both players and monsters
- `HEALTH` and `EXAMINE` now describe accumulated wounds by location, e.g. “You have a nicked head, a nicked body, a nicked and slightly punctured back…”

### Healing
- `TEND` no longer restores a flat chunk of body points — it removes one wound at a time, always the least severe first, healing exactly as many body points as that wound’s severity level (same-race targets still get the existing +50% bonus)
- Higher Healing skill is required to treat more severe wounds — skill 20 can treat any severity, and a healer who isn’t skilled enough for the least severe wound present is turned away rather than skipping ahead to an easier one
- The target being tended must now be sitting or lying down, but no longer needs to be alive — `TEND` now also works on a fresh corpse, player or monster
- Tending a wound now awards the healer experience scaled to its severity

### GM Tools
- New `@specialize <player> [<weapon> <level>]` — list or directly set a player’s weapon specialization rank (0–5), spending or refunding build points as the level changes, the same way `@mastery` already works for spells

### Fixes
- Psionic Teleportation (`PROJECT <mark>`) always sent you to mark 1 no matter which mark number you specified — it now correctly honors marks 1–10, matching Bend Space I
- Thrown weapons (javelins, spears, throwing daggers, etc.) were training your Missile Weapons skill instead of Thrown Weapons for to-hit purposes — fixed to use the correct skill, matching the original game’s distinct bow vs. thrown-weapon skills

## v11.12.0 — 2026-07-16

New WEATHER command and live temperature system, matching GM weather override, a weapon-value economy fix, and original-style spell messages for the defense buffs.

### New Commands
- `WEATHER` shows the current conditions and temperature for your region — the condition name plus a descriptive line (e.g. “A light rain falls steadily, pattering softly on the ground.”) when outdoors, or just an ambient temperature reading when indoors
- Temperature is now a live value derived from season, time of day, and the region’s current weather state, rather than a fixed or absent reading

### GM Tools
- New `@weather [value]` — shows the numeric weather state and temperature for your region with no argument, or sets the region’s weather (0–14) and broadcasts the transition to outdoor players there, same as a natural weather change

### Fixes
- Fixed a long-standing bug where the `WEA` script variable (used by `IFVAR WEA` checks in world scripts) always read region 0’s weather no matter where a script was actually running — weather-gated content in every other region (e.g. rain-triggered foraging on Island) could never correctly detect rain
- The non-magical sharpness bonus a forged weapon gets from Weaponsmithing was being stored in the same field read as the item’s copper sell value — freshly forged weapons were selling for only a few coppers instead of a real price; sharpness now has its own dedicated field, and existing forged weapons in the database were corrected with a one-time migration

### Commands
- `SKILLS` (abbr. `SKILL`) now renders as a fixed-width table (`#`, Skill, Level columns) matching the original game’s output, instead of a simple bulleted list

### Spellcasting
- `PREPARE` now shows “You prepare the \<spell\> spell.”, matching original wording
- Mystic Armor, Globe of Protection, Mass Protection, and Spectral Shield now show original-style self-cast messages — “You gesture.” → the success roll → a spell-specific flavor line (e.g. “A prismatic globe encircles you.”) — instead of one generic combined line
- Examining another player now shows a spell-specific line for these four buffs (e.g. “She is outlined in glowing armor.”) instead of the same generic “shimmering magical aura” line used for every defense spell

## v11.11.0 — 2026-07-15

Script engine execution-order rework, monster dialogue scripting, and the full “Enter the Fold” lens-of-worlds quest chain fixed end-to-end.

### Script Engine
- Rebuilt how conditional script blocks execute — actions and nested `IFVAR` checks now run in their original source order instead of every flat action always running before any nested block, fixing scripts that pick a new value and then reference it later in the same block
- `ELSE` branches now correctly support `PLREVENT`/`SETEVENT` delayed continuations — previously anything after a delay inside an ELSE branch ran immediately instead of waiting
- `IFSAY` blocks nested inside a room’s `IFVAR` tree (multi-stage conversations gated on quest flags) are now reachable — previously only top-level `IFSAY` blocks could ever match
- `KNEEL`/`SIT`/`STAND`/`LAY` and spoken commands (`SAY`/`'`) now correctly schedule delayed script continuations (`PLREVENT`/`CONTPLREVENT`) instead of silently dropping everything after the delay
- Fixed a parser bug where a `CEVENT` (cyclic world event) immediately followed by a `MACRO` definition would swallow the entire macro into its own body and fire it on a timer, unconditionally, with no acting player
- Added support for `CALL N` as an inline, mid-script subroutine call (e.g. resolving a quest item’s name for display on demand) — previously only the static room/item-level `CALL` attachment worked
- Monsters can now have their own `EXAMINE`/`GIVE`/etc. scripts attached via `SCRIPTMACRO` — this directive was silently ignored before, so no monster dialogue script ever ran
- `GIVE <item> TO <monster>` now actually works — previously GIVE only ever considered player targets
- `IFCARRY` now resolves variable arguments (not just literal numbers) and checks all three adjective slots instead of only the first — fixes quest checks for store-bought items, whose adjective is stored in the third slot
- Text set via `STRCPY` now converts underscores to spaces (e.g. “some_meteoric_dust” → “some meteoric dust”), matching how `IFSAY` patterns already worked

### Ming-K’Tuk & The Fold
- The riddle trial in the Phosphorescent Catacombs (speaking the pass phrases, then kneeling at the altar) now actually teleports you onward instead of just running a normal kneel
- Ming-K’Tuk will now accept tribute, ask his questions, and grant passage through his lair when properly flattered — the entire conversation and tribute sequence was previously unreachable
- The “Enter the Fold” lens-of-worlds scavenger hunt (Cellar → Beyond the Breach → The Void) now works start to finish — turning in a quest component is correctly detected, the hint for the next component names it correctly, and the finished lens is awarded after the full trial

### GM Tools
- New `@give [#] <item> to <player>` and `@take [#] <item> from <player>` — silently move an item between a GM’s inventory and a player’s, with ordinal and adjective matching
- Fixed `@whisper` — it reported the message as sent but never actually delivered it to the target

### Crafting
- Weapons and jewelry crafted from oiled/oily metal now correctly show both the iridescent and oiled adjectives (needed to avoid round-time in certain caves) instead of losing the oiled adjective entirely
- Mortar (used in alchemy) was incorrectly classified as an instantly-created Weaponsmithing item — it’s now properly a Jeweler craft using the same `CRAFT` → `WORK` sequence as its pestle counterpart

### Fixes
- `SNEAK` could never actually keep you hidden while moving, win or lose — fixed
- Enchantment spells no longer overwrite an item’s existing adjective when all three adjective slots are already full — the magical bonus is still applied, the item’s look is just left alone
- `OPEN`/`CLOSE` on an already-open or already-closed item now says so instead of silently repeating the action

### World
- Sharkhor spawn rate along the Inner Sea shore increased slightly (10% → 15% per check)

## v11.10.0 — 2026-07-14

Alchemy & potion brewing rebuilt from the ground up, plus fixes to spell targeting, item script dispatch, and Wood Lore crafting.

### Alchemy & Brewing
- Reagent recognition rebuilt — `BREW` now identifies every catalyst and reagent by its actual name (e.g. mandrake root, babich root, muur crystal), matching the original “Art of Alchemy” reference notes, instead of a per-instance field almost no real ingredient in the world ever had set
- Brewing order no longer matters — the catalyst and two reagents can be added in any order, as documented
- Fixed reagent/container name matching so `BREW MANDRAKE ROOT IN FLASK` and similar actually find the named items
- Potion sip count now matches the container’s real capacity (Vial 2, Flask 5, Flagon 6, Bottle 10, Ewer 6) instead of a random 2–5 regardless of vessel
- Brewed potions now get a random liquid-appearance adjective like any other potion (e.g. “a fuming potion”) instead of showing as plain “some liquid” when examined
- Removed the reference “Bottle Color” column from the recipe list and completion messages — that was only the original compiling player’s personal notes, not a game mechanic
- Each `BREW` step now takes a 15-second round (7 sec under Haste), matching other crafting skills
- Successfully brewing a potion now awards experience scaled to its recipe level (level × 20), the same formula used for jewelry/wood/weaving crafts

### Spellcasting
- `CAST` now supports ordinal targeting for item-targeted spells like Enchantment — `CAST 2 RAPIER` correctly enchants the second rapier instead of failing

### Item Scripts
- `LIGHT` and `EXTINGUISH` now check a named item’s own script before falling back to standard lighting logic
- `PLAY` now runs a named instrument’s own script instead of always printing generic flavor text — fixes instruments like the teak flute that have distinct wielded/not-wielded text
- Script engine now supports `IFITEM -1 WIELDED` checks (previously only WORN was recognized), and wielded/off-hand items are now correctly tagged when their scripts run

### Wood Lore & Crafting
- Fixed a parser bug where FLETCHER-tagged rooms (bowyer/fletcher shops) were never recognized as valid crafting workshops, blocking Wood Lore `CRAFT` entirely
- Fixed several Wood Lore items (musical instruments, staves, and other non-launcher items) awarding zero experience on completion due to missing source data — they now fall back to a sensible difficulty-based reward

## v11.9.0 — 2026-07-13

New potions system — random potion drops, containers, alchemy analysis, and pouring — plus several spell-duration fixes and a new APPEARANCE command.

### Potions
- Potion containers (bottle, flask, vial) now drop as loot from monster kills or turn up inside chests, filled with 2–10 sips of a random spell scaled to the monster’s treasure level
- Flask and vial get a random material adjective (glass, jade, obsidian, crystal, etc.); every container reveals a random liquid appearance (crimson, cloudy, fizzing, reeking, etc.) once opened
- `SIP`/`DRINK` now actually casts the potion’s spell instead of printing “[Spell effect coming soon.]”
- `LOOK IN` and plain `EXAMINE` on a potion container report fullness (full, 3/4 full, half full, 1/4 full, almost empty) and describe the liquid inside
- New `POUR <container> INTO <container>` command to transfer liquid between containers
- `ANALYZE` can now identify a potion’s magical properties for players with Alchemy skill
- Potions (and items generally) sitting inside an open container can now be targeted directly by `SIP`, `EXAMINE`, `ANALYZE`, `@iexamine`, and `@editem` — including by the potion’s own color rather than the vessel holding it (e.g. `SIP CRIMSON POTION`)

### Spell Duration Fixes
- Fly (224) and the new Heat Shield/Cold Shield (507/508) are now proper 20-minute timed buffs like other spells, extending on recast instead of lasting forever or doing nothing at all
- Heat Shield and Cold Shield now actually reduce heat and cold damage taken by 50% while active
- Potion/scroll/wand-triggered Strength, Agility, Mystic Armor, and defense-spell buffs are now temporary (20 minutes, extending on recast) instead of permanent stacking increases
- Psionic Flight (Mind over Matter 10) no longer persists across logout, and players no longer get stuck “hovering in the air” after landing

### New Commands
- `APPEARANCE` (abbr. `APPEAR`) lets you set a custom line shown when others examine you, appended after your worn equipment
- `COMMAND LOOK` lets a summoned creature look around its room and report back

## v11.8.0 — 2026-07-11

New TARGET command for multi-target spellcasting; Chain Lightning, Flaming Arrows, and Siryx’s Terrible Tentacles rebuilt to hit every creature you’ve targeted.

### TARGET Command
- `TARGET <creature>` (or `TAR`) builds a list of up to 6 creatures in the room for multi-target spells — supports ordinal disambiguation (`TARGET 2 werewolf`) just like ATTACK
- Targeting the same creature twice is rejected (“That is already being targeted.”), and the list reports when it’s full (6/6)
- Targets are automatically dropped from the list if they die, flee, or otherwise leave the room, freeing a slot for a new target

### Chain Lightning (Conjuration 132)
- With an active TARGET list, the bolt now arcs from you to the named target and then chains through every other targeted creature in turn, each taking its own independently-rolled damage
- Falls back to a single bolt at one target if no TARGET list has been built, same as before

### Flaming Arrows (Conjuration 131)
- Now fires one flaming arrow at every creature in your TARGET list, each independently rolled for damage

### Siryx’s Terrible Tentacles (Conjuration 134)
- New immobilize spell — black tentacles burst from the ground and grab hold of every creature in your TARGET list (or a single named target), with no body-point size limit on what it can restrain
- Entangled creatures can’t attack or flee, and take crushing damage once every minute until they die, break free, or the spell expires

## v11.7.0 — 2026-07-10

Necromancy overhaul — undead-only Turn/Destroy Undead, Control/Animate Undead, Reconstruction, Regeneration, Speak with Dead, Summon Spectral Warrior; SAY command and “.” repeat-last-command.

### Undead-Only Spells
- Turn Undead I/II (301/302) and Destroy Undead I/II/III (339–341) now only affect actual undead creatures (RACE 22, e.g. skeletons and zombies) — casting on a living creature now fizzles with no effect instead of dealing damage

### Control Undead I/II (Necromancy 308/309)
- Dominate an existing undead creature in the room — `CAST control undead i <creature>` — and command it with the same `COMMAND FOLLOW/GUARD/ATTACK/BEGONE` verbs used on summoned elementals
- Control Undead I only works on undead with 100 or fewer body points; Control Undead II raises the cap to 200
- Control lasts 40 minutes — afterward the undead breaks free of its bonds and turns hostile again

### Animate Skeleton / Animate Zombie (Necromancy 306/307)
- Summon a fresh skeleton or zombie fully under your control, permanent until dismissed — same command set as a summoned elemental

### Summon Spectral Warrior (Necromancy 353)
- Summons a spectral warrior under your command — requires some ghoul dust as a reagent, consumed at `PREPARE`

### Undead Healing & Harm
- Body Restoration I/II/III (316–318) now sear an undead target with holy energy as damage instead of healing them
- Reconstruction (337) now only heals undead targets — casting it on a living creature fizzles with no effect
- Regeneration (343) is now a heal-over-time spell: heals once immediately on cast, then heals the same amount again once per minute for 5 more minutes

### Speak with Dead (Necromancy 311)
- Cast on a dead player’s body to grant them the power of speech again, even though they remain otherwise incapacitated until DEPART or resurrection

### Commands
- `SAY <message>` now works as an alias for `'<message>` / `"<message>`, including automatic ask/exclaim detection and speech-manner overrides
- Entering `.` by itself now repeats your last entered command (e.g. `attack fire giant` then `.` attacks again)

## v11.6.8 — 2026-07-08

INLAY and INSET jeweler commands, spell reagent fixes, and a new Breath of Life resurrection spell.

### Jeweler Commands
- Added `INLAY <item> WITH <gem>` and `INSET <item> WITH <gem>` alongside `ENCRUST` — same requirements (Jeweler 3, forge/workshop, 2 free adjective slots), different resulting adjective (e.g. an emerald inlaid ring, a sapphire inset amulet)

### Spell Reagents
- Fixed Enchantment II and Enchantment III incorrectly demanding a spider eye/imp toe at cast time even when the spell was chanted from a scroll instead of self-prepared — reagents are now only required when a player prepares and casts the spell themselves
- Closed a loophole where casting a reagent-requiring spell in one step (`CAST <spell> <target>` without a prior `PREPARE`) skipped the reagent check entirely — it now directs you to `PREPARE ... WITH ...` instead

### Breath of Life (Necromancy 305)
- New resurrection spell — requires some mandrake root as a reagent (consumed at `PREPARE`), then `CAST <dead player>` to raise them where they fell, restoring 1-10 body points

## v11.6.7 — 2026-07-05

Guard-redirected combat message routing, Crescent muldragun lair over-spawning, GM ordinal targeting for lock commands.

### Combat
- Fixed guard-redirected monster attacks not sending the defender their own `[ToHit: X, Roll: Y]` detail — when a player guards another and an attack is redirected to them, the guard now sees their private combat roll instead of only the room’s simplified broadcast line
- Guard-redirected combat now saves the correct player’s health and status afterward — was previously saving the original target instead of the guard who actually took the hit

### Monster Spawning
- Fixed the Crescent muldragun/bhulondag lair spawning far too many monsters at once — duplicate `MLIST` spawn entries defined in both MONSTERS.SCR and HAVEN.SCR were being loaded twice, doubling every spawn roll for that group
- Monster groups no longer fill to their full population cap in a single check — each check now adds at most one monster per spawn entry, so numbers build up gradually instead of bursting in all at once
- Periodic respawn checks slowed roughly 3.5× (about every 30 seconds → about every 105 seconds)
- Added a 20-second per-room spawn cooldown — a group of players arriving together, or rapidly leaving and re-entering, no longer triggers a separate spawn roll for every single arrival

### GM Tools
- `@unlock`/`@lock`/`@open`/`@close` now support ordinal targeting (e.g., `@unlock 2 door`) to disambiguate when multiple matching items are in the room

## v11.6.6 — 2026-07-03

Spell cast message overhaul — original per-spell flavor text, caster/onlooker perspective split, and shared damage lines.

### Spell Flavor Text
- Damage spells with no custom text now correctly fall back to generic per-damage-type flavor (bolt of energy, ball of flame, freezing sphere, bolt of lightning, force blast) — and the caster now sees a second-person line (“You form a freezing sphere…”) while onlookers see the third-person version (“Chandra forms a freezing sphere…”)
- Call Meteor (112) restored to its original two-part effect — hammers the target with independently-rolled heat and crushing damage, each shown as its own damage line (“…burn to …” / “…blow to …”)
- Frost Ray (120), Lightning Bolt (103), Spectral Sword (345), and Earth Spike (523) now use their original spell-specific cast text instead of the generic elemental fallback
- Web spell now shows “\<target\> is covered with strands of sticky webbing!” to both caster and room, replacing the old generic entangle text
- Onlookers in the room now see the same damage line as the caster (e.g. “Minor blast to body. [20 Damage]”) for every damage spell, instead of only the caster seeing the result

## v11.6.5 — 2026-07-02

Foraging overhaul (terrain tables, Wood Lore gating, skill-scaled rarity), crafting material search fix, free player-taught training.

### Foraging
- Fixed a script-loading bug that silently dropped almost every real `FORAGEDEF` entry — seasonal terrain tables (forest/mountain/plain/swamp) and the jungle table were never being loaded, so `FORAGE` fell back to generic bare-noun items with no adjectives
- Foraged items now correctly carry their proper adjective (e.g. “yulman leaf” instead of just “leaf”)
- `FORAGE` now requires Wood Lore skill (18) — “You have no training in Wood Lore.” if untrained
- Success chance now scales with Wood Lore skill and Perception (base 30%, capped 90%), with a 10-second round time
- Higher Wood Lore skill biases the odds toward rarer finds (e.g. mandrake root) instead of only ever turning up common plants

### Crafting
- `WORK <material>` now keeps searching your inventory when it finds a matching-but-unsuitable item instead of giving up immediately — e.g. a finished “cotton jacket” no longer blocks raw cotton later in your pack from being found
- Material matching now recognizes fully-qualified names (e.g. `WORK BROWN SNAKE SKIN`) to disambiguate between multiple similarly-named materials, in addition to a bare noun like `WORK SKIN`

### Training
- `TRAIN` no longer charges gold for ranks made available by a fellow player’s `TEACH` — only ranks within an organization’s own posted training cap are charged

## v11.6.4 — 2026-06-30

Jewelry, weaving, and wood lore multi-step crafting; ENCRUST and ENGRAVE commands for Jewelers.

### Crafting
- Jewelry, dyeing/weaving, and wood lore now use the same multi-step `CRAFT → WORK` cycle as weaponsmithing — `CRAFT <item>` begins planning, then `WORK <material>` each step to complete it
- Each craft type has its own step flavor: jewelry (shape → refine → polish), weaving (mount → cut & stitch → hem), wood lore (carve → sand → oil)
- Material is matched by noun name — if an item matches the noun but isn’t a valid crafting material, you receive a “not suitable” message instead of a confusing error

### ENCRUST Command (Jeweler 3)
- `ENCRUST <item> WITH <gem>` — set a gem stone into an encrustable item
- Item must have the ENCRUSTABLE flag and at least 2 free adjective slots; gem is consumed
- Result: gem name and “encrusted” are applied as adjectives (e.g. a gold ring becomes “an emerald encrusted gold ring”)
- Requires a forge or jeweler’s workshop; 30-second round time (15 sec under Haste)

### ENGRAVE Command (Jeweler 3)
- `ENGRAVE <item> WITH <inscription>` — engrave a custom inscription onto an item
- Works on any ENCRUSTABLE item or any item made of hard metal
- Inscription is set verbatim as the item tail — `ENGRAVE mug WITH etched with a howling wolf` produces “a mug etched with a howling wolf”
- Requires a forge or jeweler’s workshop; 30-second round time (15 sec under Haste)

## v11.6.3 — 2026-06-29

Summoning system overhaul: elemental guard combat, immunities, reagent timing, and creature lifetime management.

### Summoned Creature Guard
- `COMMAND GUARD ME` now actually intercepts melee attacks — monsters targeting you hit the elemental instead
- Ranged attacks (bows, guns, thrown weapons) bypass the guard and land on you normally
- Guard combat uses the same `[ToHit: X, Roll: Y] Hit! / Miss!` format seen in normal combat
- Elementals with `MAGICWEAPON` requirements are immune to non-magical attacks — the TEXI flavor text is shown on hits that do no damage (e.g. “The attack serves only to scatter a little water about.”), while misses show a normal miss line
- When the elemental is destroyed while guarding, its TEXD death text is shown and the summoner receives a psychic shock stun (2–5 seconds)

### COMMAND Improvements
- `COMMAND FOLLOW ME` is now a toggle — issuing it again stops the creature from following (“The water elemental stops following you.”)
- All COMMAND actions (FOLLOW, GUARD, ATTACK, BEGONE) now broadcast a confirmation message to the room
- Summoned creatures follow their summoner through named portals and GO exits, not just cardinal directions

### Reagent Timing
- Elemental summoning reagents (garnet, opal, aquamarine, tourmaline, diamond) are now consumed at PREPARE time, not on CAST
- Message shown on prepare: “The [gem] turns to dust as it is consumed by the spell.”

### Creature Lifetime
- Summoned creatures are automatically dismissed if the summoner dies or disconnects
- Summoned elementals no longer auto-aggro nearby players or wander between rooms

## v11.6.2 — 2026-06-28

Physicians’ Guild healer rooms and full Bank of Fayd banking services.

### Healer Rooms
- Rooms tagged as `HEALER` (e.g. Physicians’ Guild, room 312 in Fayd) now provide medical treatment when you SIT or LAY down
- Physicians heal all wounds instantly — charged 1 copper penny per body point restored
- If you have no money at all you are turned away; if you have some but not enough you are charged everything you have and still fully healed
- Cure poison: 10 gold crowns — removes all poison and poison severity
- Cure disease: 10 gold crowns — removes all disease and disease severity
- Poison and disease are treated independently before wounds; each checks your current purse after any prior charge
- If you are already in perfect health and have no conditions, the physician tells you so

### Banking
- Rooms tagged as `BANK` (e.g. Bank of Fayd, room 355) now offer full banking services
- `DEPOSIT <amount>` — deposit copper pennies; suffix with GOLD or SILVER to use those denominations (e.g. `DEPOSIT 5 GOLD`)
- `WITHDRAW <amount>` — withdraw from your account, same denomination suffix supported; coins returned as carried currency
- No fee to deposit or withdraw money
- `DEPOSIT <item name>` — place an item into your safety deposit box for 1 gold crown; item matched by name and adjectives (e.g. `DEPOSIT steel longsword`)
- `WITHDRAW <item name>` — retrieve an item from your safety deposit box, no charge
- Safety deposit box holds up to 20 items; you must afford the 1 gold fee to deposit an item
- `BALANCE` — shows your cash balance and lists all items stored in your safety deposit box (X/20 slots used)

## v11.6.1 — 2026-06-26

Spell mastery system, 46 missing spells, poison/disease severity levels, herb fixes, Cure Poison and Cure Disease spells, wave emote.

### Spell Mastery
- Spell mastery system implemented — when your school skill reaches double a spell’s level, you achieve first mastery rank; each additional rank requires one more spell level of skill
- Each mastery rank reduces mana cost by 2, floored at half the base cost (minimum 1)
- SPELLS command now shows a School column and mastery stars: `(*)`, `(**)`, `(***)`, etc.
- SPELLS list is sorted by school, then level, then spell ID

### New Spells
- 46 spells added to the registry across all five schools (Conjuration, Enchantment, Necromancy, General, Druidic) that were present in the original spell list but missing from the engine — they now appear when learned and are castable
- Cure Poison (spell 303, Necromancy level 11) — removes poison and all poison levels from self or a target in the room
- Cure Disease (spell 319, Necromancy level 12) — removes disease and all disease levels from self or a target in the room

### Poison & Disease Severity
- Poison and disease now have severity levels (1–5) — each level deals 1 BP damage per regen tick
- Trap severity: minor traps (types 1, 2) = level 1; moderate (types 5, 12) = level 2; major (type 7) = level 3; lethal black needle (type 13) = level 5
- Monster hits that poison or disease default to level 1 and never downgrade an existing higher level
- Regen tick message shows damage amount and current BP: `The poison courses through your veins! [-2 BP, 45/80]`

### Herbs & Consumables
- Rowik berry now reliably cures poison and clears poison level, even when the item script has no ITEMADJ3 block
- Babich root and piece of babich root now correctly heal 5–15 BP (Body Restoration I)

### Emotes
- WAVE with no target now emotes “You wave.” / “[Name] waves.” instead of returning “Wave what?”

## v11.5.18 — 2026-06-25

Vertical passage movement fix, magical item charge detection, timed defense spells target other players.

### Movement & Script Engine
- Fixed infinite wait loop in vertical passages (Devil’s Chimney, deep crevices) — `EQUAL ROUNDTIME` in an `IFPREVERB` block without `CLEARVERB` now applies roundtime as a post-move penalty, not a pre-move block
- Roundtime check now happens before room scripts fire on directional movement, matching original engine behavior

### Magical Items & Detect Magic
- Detect Magic (spell 400) now correctly reads charges from Val2 (was incorrectly checking Val4)
- Detect Magic now reports remaining charges in its output: e.g. `glows a soft blue (3 charges remaining)`
- Items with a stored spell but no charges remaining now show “a faint magical residue — completely drained of power” instead of glowing
- Magical item treasure drops now correctly store charges in Val2 (was Val4)

### Defense Spells
- All timed defense spells (Globe of Protection, Globe of Protection II, Spectral Shield, and others) can now be cast on yourself or on another player in the room, matching Mystic Armor behavior
- Casting on another player shows distinct messages to caster, target, and room

## v11.5.17 — 2026-06-24

Flying movement overhaul, script MOVEGROUP fixes, geyser ROOMCOPY, round time display.

### Flying & Movement
- ASCEND and DESCEND now work correctly for players who used the FLY command — previously only worked for Drakin or players with psionic flight
- EXIT ABOVE and EXIT BELOW no longer appear in the exits list — they are fly-only passages, accessible only via ASCEND/DESCEND
- IFPREVERB ASCEND/DESCEND scripts now fire before movement — rooms can block or redirect ascent/descent with a message
- ASCEND/DESCEND abbreviations (e.g. `asc`) now resolve correctly

### Script Engine
- MOVEGROUP in IFVERB blocks now fires correctly — was only being checked in IFPREVERB results
- PLREVENT and CONTPLREVENT (player-timed events) are now implemented — enables deferred room entry sequences such as river currents pulling players downstream
- ROOMCOPY script command implemented — copies exits and description from a template room into the current room; powers the geyser system that periodically opens new exits
- ECHO OTHGROUP now implemented — sends a message to room occupants who are not in the player’s group
- EQUAL ROUNDTIME now shows a `[Round: X sec]` message to the player, matching original engine behavior
- MOVEGROUP in CLIMB scripts now fires correctly and shows round time; success and failure echoes no longer both display after a successful group move

## v11.5.16 — 2026-06-23

Weaponsmith XP scaling, cyclic world events, smelting round time and XP, gold and silver ore.

### Crafting
- Weaponsmith XP now scales correctly: 25 XP per skill level required by the weapon, plus a metal quality bonus (up to +150 for truesteel) and a blade sharpness bonus (up to +100)
- Smelting now applies a 5-second round time (2 seconds if Hasted) before the smith can act again
- Smelting now awards XP on success: 5 for copper/tin, scaling up to 50 for truesteel and 100 for exotic metals
- Gold and silver ore can now be found when mining — grade A mines yield silver (8%) and gold (4%), grade B mines yield silver (5%) and gold (1%); grade C mines do not produce precious metals

### World Events
- Fixed a parser bug where all 51 cyclic world events (CEVENT) were silently discarded at server startup — timed events such as Menelian’s wandering cottage now fire correctly
- Menelian’s cottage cycles through rooms 128, 156, and 944 every ~12.5 minutes, appearing as a door that players can enter

## v11.5.15 — 2026-06-22

Herb effects, elevator mechanics, script engine AFFECT fixes.

### Herbs & Consumables
- Stat-boosting herbs (yarrow lichen, quillim fruit, malatrin leaf, shay-ahm blossoms, zarus stem, coriam seed, kurkan pollen) now correctly limited to one dose each
- Stat boost range corrected to 1–4 (was 2–4)
- Rowik berry now cures poison
- Disease-curing herb now implemented
- Thesnia leaf (mindlink) now refreshes on every bite, not just the first

### Script Engine
- FLIP switch now correctly fires IFVERB scripts after the state change — fixes power-on scripts in Technologists Hall
- Room-level IFTOUCH blocks now execute during IFPREVERB handling — fixes the Technologists Hall identity scanner
- IFSAY pattern matching now strips trailing punctuation — “computer, identify” matches whether or not the player adds a period
- PLRSINROOM and MONINROOM now count players/monsters in the AFFECTed room rather than always the player’s room — fixes the Technologists Hall elevator auto-return

## v11.5.14 — 2026-06-21

Group disconnect handling, GUARD command overhaul, improved guard messages, STATUS build point fix.

### Group System
- When a group leader disconnects or logs off, the entire group is disbanded and all members are notified
- When a group member disconnects or logs off, they are removed from the leader’s group and the leader is notified

### GUARD Command Overhaul
- GUARD now requires Combat Maneuvering skill at level 2 or higher to guard a player
- Combat Maneuvering level 3 unlocks portal guard — anyone trying to pass a guarded gate or doorway must make a skill check
- Any Combat Maneuvering level can guard an item on the ground — others must make a skill check to pick it up
- You can now guard multiple targets simultaneously — GUARD \<target\> toggles each guard on or off independently
- Cannot guard a player who is already themselves on guard duty
- Hidden or invisible bypassers are announced as “something” rather than by name in all guard messages
- Skill check formula: roll d100 + CM×5 + (AGI−50)/10 + (QUI−50)/10 vs 50 + guard’s CM×5

### Improved Guard Messages
- Skill roll results now shown on both success and failure: `[Roll: 73 vs 60]`
- Successful bypass now shows “You slip past X’s guard” before the pickup or movement confirmation
- Guard sees their own success/failure with the roll result
- Room bystanders see appropriate success/failure echoes for portal and item guard attempts

### Fixes
- STATUS command now shows correct “Build Points to date” and “Unspent Build Points” values — was previously double-subtracting spent points

## v11.5.13 — 2026-06-20

Weapon repair fixes, damaged adjective, oily metal crafting correction.

### Combat
- Weapons damaged in a clash now show a “damaged” adjective — a damaged longsword appears as “a damaged longsword”
- Successfully repairing a weapon removes the damaged adjective

### Crafting
- Oily and oiled metals no longer produce weapons with icy elemental crit damage — they were incorrectly inheriting cold-damage properties
- Weapons and jewelry crafted from oily or oiled metal now show as iridescent instead

### Fixes
- REPAIR now matches weapons by adjective — “repair damaged longsword” correctly finds the item
- REPAIR no longer says “doesn’t need repair” when the weapon isn’t in your inventory — now says “you aren’t carrying that”

## v11.5.12 — 2026-06-19

Weapon clash metal bonuses, off-hand inventory fix, crafted weapon sharpness system.

### Combat
- Weapon clashes now account for metal type — harder metals resist damage: steel +25, truesteel +40, elkyri +55, albescent +65, randar +75
- Soft metals (copper, silver, gold, tin) provide no clash bonus and are more likely to be damaged or broken

### Crafting
- Crafted weapons now receive a non-magical to-hit bonus (Val1) based on the ore’s hidden color quality
- Ore color (purple / indigo / blue) determines the sharpness ceiling — blue is the best currently mineable
- Smith’s Weaponsmithing skill level shifts the bonus range — higher skill yields sharper weapons
- Weapon completion message now reports blade quality and bonus (e.g., “The blade is very sharp. (+7 non-magical bonus)”)
- ANALYZE ore at Mining 5+ now also reveals the metal’s color quality alongside purity

### Fixes
- Off-hand weapon (Two Weapons skill) now appears at the top of the INVENTORY list alongside wielded and worn items, labeled “(off-hand weapon)”

## v11.5.11 — 2026-04-29

Crafting pipeline fix, LOOK IN containers, CRAFT recipe list, manuscript protection.

### Crafting
- Smelting now preserves ore type — iron ore smelts to iron metal, not generic metal
- WORK command matches instance adjectives — “work steel” finds steel metal in inventory
- CRAFT with no args lists all available recipes

### Fixes
- LOOK IN now works for containers in rooms and inventory (chest, sack, coffer, etc.)
- Manuscripts and FIXED items can no longer be picked up

## v11.5.10 — 2026-04-28

Players no longer look dead, date/time on login, mining tools from inventory, ambient text reduced.

### Critical Fix
- Players no longer appear dead when examined — fixed divide-by-zero when MaxBodyPoints was 0

### Features
- Date, time, and weather shown when entering the game world
- Mining tools work from inventory (not just wielded)
- Mining accepts “pick-axe” (hyphenated) in addition to “pickaxe”

### Balance
- Ambient NPC text frequency reduced (15% → 8%) for less screen clutter

## v11.5.9 — 2026-04-28

Build point fix, level-up mana/psi, open-ended crits, typed mining ore, GM idle exempt.

### Critical Fix
- Build Points no longer reset when viewing EXP/STATUS — spent BP are now properly tracked

### Combat
- Open-ended critical hits (96-100): bonus damage roll, with double-open-ended for devastating crits
- Level-up now increases MaxMana and MaxPsi (based on Willpower/Empathy)

### Crafting
- Mining now produces typed ore (iron, copper, bronze, tin, steel) based on mine grade

### Fixes
- GMs exempt from 30-minute idle timeout
- SET accepts TRUE/FALSE/YES/NO in addition to ON/OFF
- LOOK IN container shows contents (or “empty”)

## v11.5.8 — 2026-04-28

Ethereal Projection persistence, Flight cancel, Manipulate Lock, LOOK direction items.

### Psionics
- Ethereal Projection now persists through movement (psi-maintained hiding)
- SEARCH no longer reveals players using Ethereal Projection
- Canceling Flight/Levitate properly resets position and CanFly
- Manipulate Lock (PSI 8) now works — psionically unlocks locked items

### Fixes
- LOOK direction now shows room items in addition to players and monsters
- Bot login/logout messages suppressed
- @vis/@invis now reports if already in that state
- @set FATIGUE always sets both current and max
- TEND self says “You don’t need healing” instead of “yourself doesn’t need healing”

## v11.5.7 — 2026-04-25

PSI Flight fix, GM invisibility, Ethereal Projection, emotes on dead monsters.

### Psionics
- PSI Flight (10) now properly sets CanFly — flying players can go UP/ASCEND
- Ethereal Projection (PSI 15) is now a maintained buff that makes you hidden

### GM Invisibility
- Invisible/hidden GMs no longer broadcast login/logout messages
- Prompt ‘H’ indicator now shows when GM is invisible (persists across sessions)

### Fixes
- Emotes now work on dead monsters (kick that dead rat!)
- Prompt ‘P’ indicator now uppercase for Poisoned (per original manual)
- ‘C’ indicator used for combat-joined (was conflicting with P for Poisoned)

## v11.5.6 — 2026-04-23

Flight movement, PSI persistence, INITIATE, per-IP limit, NPC pacing, ordinals.

### Psionics
- Prepared psionic power now persists after projection — no need to re-prepare each time
- Flying players can now move normally (PSI Flight no longer blocks movement)
- Added Ethereal Projection (PSI 15) discipline

### Commands
- INITIATE \<player\> \<org#\> — GM command to set player organization
- “GO COUNTER 2” now works (trailing numbers parsed as ordinals)
- “WORK METAL” now prompts for specific metal type instead of confusing error

### GM Tools
- @set BP/FAT/MANA/PSI now sets both current and max values
- Use BODYPOINTS/MAXBODYPOINTS for fine-grained control

### Balance & Polish
- NPC ambient wander rate halved (less screen clutter)
- Per-IP connection limit raised from 5 to 8
- Chrome desktop: additional focus retention fix

## v11.5.5 — 2026-04-23

PSI teleportation, partial item names, @echoplr fix, emote articles, custom descriptions.

### Psionics
- PSI Teleportation (ID 12) now works — teleports to your marked location

### Fixes
- `@echoplr` now works with any capitalization of player name
- DROP/GET now match partial item names (e.g., “drop tooth” matches “rat tooth”)
- Emotes with monsters now include articles (“You sniff a lost mutt” not “You sniff lost mutt”)
- Custom @line descriptions now replace the auto-generated race/gender line in LOOK
- VER command now shows correct version number
- Character names must be at least 3 characters (prevents single-letter names)
- Chrome desktop: improved input focus retention after command output

## v11.5.4 — 2026-04-23

GM edit target, shields, skinning, position checks, CANT, @answer, STOMP.

### GM Tools
- `@edpl <player>` now sets an edit target — subsequent `@set` commands modify that player
- `@edpl` (no args) clears the edit target
- `@answer` — teleports GM to the last player who used ASSIST
- @yank offline message clarified

### Fixes
- Shields now have a default worn slot — WIELD/WEAR shield works correctly
- Can no longer skin the same corpse multiple times
- Can no longer SIT/STAND/KNEEL/LAY if already in that position
- Mark list no longer shows room numbers to non-GM players
- THUMP and script-interactive verbs now fire IFVERB scripts with emote fallback

### Features
- CANT (thieves’ cant) now requires Stealth or Legerdemain skill — only skilled players can hear it
- STOMP emote added

## v11.5.3 — 2026-04-23

Stun, immobilize, healing cooldown, multi-attacker defense, NPC flight restrictions.

### Combat
- Stun now works: stunned monsters skip their next attack and are easier to hit (+20 to-hit)
- Multi-attacker defense penalty: -5 per 2 additional monsters attacking you
- Polearm fatigue reduced (capped at 3 per swing instead of scaling with weight)

### Psionics
- PSI Immobilize now implemented — freezes target in place (monsters and players)
- PSI damage disciplines now give a clear message when targeting players

### Skills
- TEND/HEAL now has a 5-second round timer and cannot be used in combat

### World
- NPCs can no longer wander through ABOVE/UP exits into sky rooms
- THUMP and similar verbs now trigger room scripts before falling back to emotes
- Immobilized players cannot move

### UI
- Mobile keyboard now auto-capitalizes after periods

## v11.5.2 — 2026-04-23

Seasonal rooms, martial arts, healing, whisper, idle timeout, and more fixes.

### Seasonal World
- Seasonal room overrides now load at startup and on season change — fixes missing rooms like Grymwood
- Rooms defined in spring/summer/autumn/winter scripts now apply their descriptions, exits, and items

### Combat & Skills
- Martial Arts now boosts unarmed damage: +1 per rank + expanded damage range
- Monsters lose their target when the player hides (no more attacking hidden players)

### Fixes
- Whisper target no longer sees both the whisper content and 3rd-person room message
- HEAL another player now saves the target’s updated health
- WIELD shield now properly routes to WEAR for any shield or wearable item
- Fixed “You look looks sickly” double-verb in self-examine when poisoned/diseased
- Fixed double space in skin command output (“skin a  giant rat”)

### GM Tools
- `@setp <player> <var> <value>` — set variables on another player (fixes @set after @edpl)

### Server
- 30-minute idle timeout — inactive players are now disconnected automatically

## v11.5.1 — 2026-04-23

Second round of player-reported fixes: scripts, groups, combat, GM tools.

### Scripts & World
- Scripts can now charge gold — `SUB WEALTH` / `SET WEALTH` in scripts now works (fixes herb shops, vendors)
- `MOVEGROUP` script command implemented — moves all players in a room to a destination
- Monsters no longer spawn in sky/ABOVE rooms
- Death telepathy — psionic characters now sense when another player dies

### Group & Follow
- Group members now follow through portals, gates, and doorways
- Stale “J” prompt indicator cleared when leader goes offline

### Combat & Abilities
- BERSERK stance now available to all races (not just Murg)
- Psionic Levitate now grants flight (fixes FLY command for non-Drakin psionics)

### Crafting
- Dyeing now preserves material adjective — color goes to second adjective slot instead of overwriting the material name

### GM Tools
- `@gop` no longer reveals invisible GMs (ExitEcho/EntryEcho suppressed when @invis)
- `@yank` now shows the yanked player their new location

## v11.5 — 2026-04-23

Feedback pipeline, 16 bug fixes from player reports, email deliverability.

### Feedback Pipeline
- REPORT command now forwards player feedback to VibeCtl for triage and issue tracking
- Reports include player name, room location, and full message text
- Backfilled 205 existing reports from game logs

### New Features
- `APPRAISE <item>` — see what a merchant will pay before selling
- `SEARCH` (no target) — scan the area to reveal hidden players (perception vs stealth check)
- `HEAL <player>` — now routes to TEND for healing instead of showing health stats
- `TEACH` now accepts skill names in addition to skill numbers
- WHO list is now sorted alphabetically
- Merchant payments now show proper coin names: gold crowns, silver shillings, copper pennies

### Combat & Movement
- Wolfling wolf form now uses claw attacks with higher damage (3-10 + STR/10) instead of fist damage
- Wolf form shows “claws” instead of “fists” in combat messages
- Can no longer climb ladders, go through portals, or use GO while lying down, sitting, or kneeling
- CONTACT now requires psionic abilities (Psionics or psionic school skill)

### Group & Follow
- Following now breaks when you manually move away from your leader
- Two players can no longer follow each other (circular follow prevented)

### Items & Containers
- Shields can now be equipped via WIELD (auto-routes to WEAR)
- Containers in your inventory can now be opened and closed

### GM Tools
- `@echoplr` now sends the message to the target player (was GM-only echo)
- `@set ORG <value>` — set a player’s organization for script conditionals

### Other Fixes
- Fixed invisible player bug — reconnecting no longer creates duplicate sessions
- Capture recording button now resets properly on reconnect
- Fixed input losing focus after pressing Enter on Windows and Android Chrome
- SPF, DKIM, and DMARC DNS records configured for email deliverability

## v11.4 — 2026-04-12

### GM Script Editor
- New "GM" nav button visible when playing a GM character
- Web-based script editor: upload, edit, and manage `.scr` files
- Scripts parsed, validated, and hot-loaded into running engine without restart
- Stored in MongoDB with priority ordering (higher priority loads first)
- Version history: last 10 versions with one-click restore
- Shared GM filespace with audit trail (who uploaded what, when)
- Upload `.scr` files from disk or edit directly in the text editor
- Script size limit: 262 KB (110% of largest disk script)

## v11.3 — 2026-04-11

### Mobile UI
- Responsive layout throughout — status bar, nav, admin panel, character creation all adapt to small screens
- Send button (↵) visible on touch devices; keyboard dismissed after each command so iOS zooms back out
- Predictive text / QuickType bar suppressed — autocorrect no longer mangles game commands
- Input font-size 16px on mobile, preventing iOS auto-zoom on focus
- `dvh` viewport units account for mobile browser chrome
- Login/logout messages off by default for new characters

### GM Server Banner
- `@banner <text>` — set a login banner and broadcast notice to all online players
- `@banner` (no args) — clears the banner silently
- Banner displayed on web login screen and telnet/SSH login menu
- In-memory primary storage (works even if MongoDB is offline); persisted to MongoDB as backup for restarts

### Fixes
- `@announce` now broadcasts to all online players (was only echoing back to the sender)
- Weather room descriptions now show prose instead of "The weather is Heavy Snow"
- Rare Weapons Exchange parlor door now works with GO as well as PUSH (`modern_fixes.scr`)

## v11.2.2 — 2026-04-09

### Game Calendar & Seasons
- **In-game time** runs at 6:1 ratio (4 real hours = 1 game day)
- **Full calendar**: 12 named months, 28 days each, with tracked years
- **Seasons** follow game calendar: Spring (months 1-3), Summer (4-6), Autumn (7-9), Winter (10-12)
- **Hot-swap seasonal MLISTs**: monster spawns change with the seasons
- **Season change broadcasts** to outdoor players with atmospheric messages
- **Game time persistence**: calendar state saved to MongoDB, survives restarts
- **TIME command** shows date, season, and dual moon phases (Great Moon + Phulcrus)

### Passive Regeneration
- **Fatigue** regenerates based on Constitution (/20 per tick, min 1)
- **Mana** regenerates based on Willpower + Empathy (/30 per tick, min 1)
- **PSI** regenerates based on Willpower (/20 per tick, min 1)
- **Body Points** regenerate slowly based on Constitution (/50 per tick, min 1)
- **Position multipliers**: laying 3x, sitting 2x, kneeling 1.5x, standing/flying 1x
- Ticks every real minute for all online players

## v11.2.1 — 2026-04-08

### Player Manual
- Full original Player Manual (V3.1, 1994) as in-game modal overlay
- Sticky table of contents with 29 sections
- Accessible from top nav and footer

### SET Command & Settings
- SET command: view/toggle all display settings
- Full/Brief room descriptions, prompt mode
- Filter logon/logoff/disconnect, RPbrief, Battlebrief, Actionbrief, Actbrief

### World
- Dynamic weather system with regional transitions
- Sunrise/sunset broadcasts to outdoor rooms
- PSI command: list disciplines, activate by number, toggle maintained powers
- Fixed monster spawning for seasonal scripts
- Fixed height/weight assignment on character creation

## v10.0.5 — 2026-04-06

### GM Tools
- **@zap <monster>**: actually destroys monsters (was a stub)
- **@trace**: toggle script execution debug output (shows IFVAR conditions, block types)
- **@verb/@verbs**: comprehensive alphabetical listing of every game command with parameters
- **@go/@goplr**: broadcast entry/exit echoes using custom @entry/@exit text
- **@invis** GMs move completely silently (no echoes of any kind)
- **GM flags persist**: @invis, @hide, @gm survive reconnections and server restarts

### Combat & Movement
- **ADVANCE <target>**: properly engages a monster or player in combat
- **RETREAT**: disengages from combat without fleeing the room
- **YELL heard in adjacent rooms**: "You hear someone yell..." through all exits
- **DEPART**: sends to City Gate (201) instead of tutorial room (3950)

### Fixes
- **USE**: now an item interaction verb (triggers IFPREVERB/IFVERB scripts) instead of alias for WIELD
- **ACT**: preserves original text case (was lowercasing)
- **THINK**: preserves full text including first word and punctuation
- **GIVE**: no longer shows duplicate messages to recipient (TargetMsg instead of Whisper)
- **QUIT**: no longer broadcasts "left the Realms" twice
- **Directional LOOK**: shows players and monsters in adjacent rooms
- **SEARCH**: fixed double-space in "You search a skeleton"
- **RECITE**: backslash line breaks for poetry/songs

## v10.0.4 — 2026-04-06

### Multiplayer
- **Room arrival/departure on login/logout**: players in a room see "X arrives." when someone logs in and "X fades from the Realms." on logout
- **Fix GIVE double-message**: recipient was seeing both room broadcast and whisper; now uses TargetName/TargetMsg (excluded from broadcast)
- **Fix QUIT double-broadcast**: "left the Realms" no longer sent twice (once from QUIT command, once from disconnect cleanup)
- **Fix @heal on live players**: GM commands now modify the live session player, not a stale DB copy

### Commands
- **Fix THINK preserving text**: was dropping the first word ("mmm..." in "think mmm... thesnia leaf"); now uses original input text with correct case
- **RECITE line breaks**: backslash (`\`) creates multi-line poetry/song output

### Security
- **Two-tier rate limiting**: 4 commands/sec burst + 10 commands per 10 seconds sustained
- **Max 8 characters per account**

## v10.0.3 — 2026-04-04

### Bot API (New Feature)
Bots were not part of the original game — this is a new feature added to help repopulate the world. Bots can't do anything a human player can't; they follow the same rules, rate limits, and game mechanics.

- **API key system**: generate per-character keys from the main menu (SHA-256 hashed, shown once)
- **WebSocket `auth_apikey`**: bots connect and authenticate via the same WebSocket as regular players
- **[Bot] marker** on WHO list so other players can see who's a bot
- **GM scope control**: GM bots can optionally be restricted from using GM commands
- **SDK examples**: Python, Node.js, and TypeScript in `/bots/` directory with full documentation

### Currency & Trading
- **GIVE money**: `GIVE 5 GOLD TO Taliesin`, `GIVE 10 SILVER TO player`, etc.
- **All currency types**: gold crowns, silver shillings, copper pennies
- **Regional currencies**: kragenmark, danir, shard, darktar, dollar (transferred as MONEY inventory items)
- Accepts plural forms and full names (crowns, shillings, pennies, etc.)

## v10.0.2 — 2026-04-04

### Script Engine
- **ELSE branches** in script conditionals now work (IFVAR, IFITEM, IFNOITEM, IFCARRY, IFFULLDESC, IFIN)
- **Case-insensitive file loading** for DOS-era script filenames — scans directory for matching files regardless of case
- **PORTAL_CLIMBUP, PORTAL_CLIMBDOWN, PORTAL_OVER** types added to parser (were silently falling back to MISC)
- **New script variables**: WARRANT, GFLAG1-4, NUMPLRS, ARENADEATH, SITTING, LAYING, STANDING, KNEELING, WIELDED, WEALTH, REGION

### GM Tools
- **@line1/2/3 [player] <text>** — set persistent description lines on any character, visible on EXAMINE
- **@entry/@exit <text>** — custom room entry/exit messages (replaces default "arrives"/"goes north")
- **@speech <player> <verb phrase>** — set custom speech patterns (e.g., "says grimly", "squawks", "giggle")
- **REPORT <message>** — player command to file reports (broadcast to all GMs, logged with "report" event type)
- **@help** updated with all new commands

### Security Hardening
- **Per-IP WebSocket connection limit** (5 max per IP, 500 total)
- **Command rate limiting** (10 commands/sec, then throttled)
- **Chat flood protection** (5 broadcast messages per 10 seconds)
- **HTML entity sanitization** on all outgoing messages (defense in depth against XSS)
- **REST API auth gates** — /api/characters GET and POST now require authentication
- **MongoDB regex injection fix** in resolvePlayerByName (regexp.QuoteMeta)
- **JWT lifetime reduced** from 30 days to 7 days
- **Dead player command restriction** — limited to DEPART, LOOK, WHO, QUIT, EXP, STATUS, HEALTH

### Fixes
- **Room 225 stairway** — PORTAL_CLIMBUP type was parsed as MISC, blocking GO STAIR
- **Temple of Amilor** — room 592 missing due to AMILOR.SCR case mismatch in git; ELSE branches not firing
- **SKILLS command** — was hardcoded to "no skills yet", now shows actual trained skills
- **Grammar fixes** — "A a fist" → proper article, "your an axe" → "your axe", natural weapons don't drop
- **SPEECH** — removed as player command, now GM-only via @speech

## v10.0.1 — 2026-04-04

### Character Management
- **Soft-delete characters** from main menu with confirmation modal
- **Unique first names** enforced on character creation
- **Admin character recovery**: browse deleted characters, recover with optional rename if name conflicts
- **Name validation split**: monster/game names blocked as exact match only; slurs blocked as substring
- Removed over-aggressive substring matching (e.g., "Pendragon" no longer blocked for containing "dragon")

### Script Engine Variables
- **WARRANT**: player warrant level for law/crime system
- **GFLAG1-4**: global flags accessible in scripts
- **NUMPLRS**: total online player count
- **ARENADEATH**: whether player died in arena
- **SITTING, LAYING, STANDING, KNEELING**: position state variables
- **WIELDED, WEALTH, REGION**: equipment and location variables

### Fixes
- **Fixed Taliesin login**: reserved-word validation was running on login, not just creation — "Pendragon" was blocked because it contains "dragon"
- **Name validation moved to creation only** — existing characters log in without name checks
- **Version notes deeplink**: /version-notes URL works as direct link

### Infrastructure
- **Engine test harness**: first unit test (TestLoadPlayerTaliesin) verifying MongoDB player lookup
- **Warrant field** added to Player struct for future law system

## v10.0.0 — 2026-04-04

### Crafting System
- **MINE** ore from MINEA/B/C rooms — ore purity based on grade (A=50-100%, B=30-70%, C=10-40%)
- **SMELT** ore into refined material at FORGE rooms — purity + skill determines success
- **CRAFT/FORGE** weapons and armor at FORGE rooms (Weaponsmithing skill ≥ weapon PARAMETER1)
- **WEAVE** clothing at LOOM rooms (Dyeing/Weaving skill ≥ PARAMETER2)
- **CRAFT** wood items at FLETCHER rooms (Wood Lore skill ≥ PARAMETER1)
- **FORAGE** terrain-based gathering using ForageDef tables (forest/mountain/plain/swamp/jungle)
- **DYE** materials at LOOM rooms — apply color adjectives from DYE items to DYEABLE materials
- **ANALYZE** ore purity (Mining 3+) and reagent properties (Alchemy)
- **BREW** potions via alchemy — 32 recipes from original alchemy.bin (3 reagents = catalyst + 2 types)
- Recipes: Body Restoration, Strength I-III, Agility I-III, Mystic Armor, Haste, Invisibility, Globe of Protection, Cure Poison/Disease, and more
- Potions yield 2-5 sips, skill level gates recipe access

### Full Skill System (36 skills)
- **Build point costs** from skills.txt (first rank + per-rank costs, e.g., Edged 12/5, Healing 20/2)
- **Skill prerequisites** enforced (magic schools need Spellcraft, Mind over Mind/Matter needs Psionics, Dodge needs weapon skill, Disguise needs Stealth)
- **Weapon skills**: +5 attack per rank (Edged, Crushing, Polearms, Missile, Drakin, Natural, Thrown)
- **Dodge & Parry**: +5 defense per rank (requires any weapon skill)
- **Martial Arts**: +5 attack unarmed, +2 defense unarmed per rank; 10+ ranks = hit magic-required monsters
- **Combat Maneuvering**: -1 sec roundtime per rank; 2% per rank chance to dodge monster special attacks (max 95%)
- **Endurance**: +4 max body points per rank; 1% elemental damage reduction per rank (max 50%)
- **ANOINT**: apply poison to wielded weapon (Trap & Poison Lore skill, level = rank)
- **TEND**: heal wounds with Healing skill (2 + skill×2 + random, +50% same-race bonus, stops bleeding)
- **UNLEARN**: remove one skill rank, get back build points minus one
- 30 starting build points for new characters

### Treasure & Loot System
- **Treasure tables** based on monster TREASURE level (0-127)
- **Coin drops** always (copper amount scales with level)
- **Weapon drops**: level-appropriate from item database, chance for magic bonus and premium materials
- **Armor drops**: level-appropriate with magic bonus chance
- **Spell scrolls**: random learnable spell on scroll (spell level = treasure/3)
- **Locked chests**: with lock difficulty and traps (13 types + spell glyphs)
- **Monster weapon drops** on death (with WeaponPlus bonus)
- **SEARCH** dead monsters for treasure (one-time loot, corpses decay after 60s)
- **MONEY items** auto-convert to currency on GET; visible as "some coins" in LOOK
- **SELL** uses VAL1 as copper value (merchants pay 50%)

### Combat Fidelity (from LEGENDS.DOC / shirla.cap)
- **Fatigue drain** on melee attacks based on weapon weight (ranged exempt)
- **Fatigue ToHit penalties**: -10 at half fatigue, -25 at quarter fatigue
- **Weapon clash** on roll < 3 vs weapon-wielding monsters (2d100 vs weapon strength)
- **Damaged weapons**: -10 ToHit penalty; break on second clash
- **Backstab** requires puncture weapon only (daggers, rapiers)
- **Death**: 90% XP penalty toward current build point
- **Spellcraft**: base 25% + EMP/10 + spellcraft×5 (max 95%), fumble on 98+
- **Mana cost** = spell level (from LEGENDS.DOC)
- **NOCK/LOAD** ranged weapons with ammunition; must reload between shots
- **Invisible vs Hidden** distinction: Invisibility spell not broken by movement

### Monster System Improvements
- **Demand-based spawning**: monsters only spawn when players enter rooms
- **Correct MLIST format**: probability/maxCount (not min/max)
- **Monster unloading**: despawn after 3 minutes with no players (ETERNAL exempt)
- **Psi defense auto-activation** on spawn (Wall of Force, Psychic Shield, etc.)
- **Local-only monster broadcasts** (no cross-machine ghost monsters)
- **Corpse decay** after 60 seconds; dead monsters show as "(dead)" in LOOK

### Other
- **@mlist** GM command: show all spawned monsters worldwide
- **@lsk** GM command: list all skills with IDs
- **@edpl** alias for @edplayer, **@edsk** alias for @eds
- **@help** sorted alphabetically with all commands listed
- **CREDITS** updated with full original team and 2026 re-release info
- **REVEAL/UNHIDE** commands; auto-reveal on movement, emotes, attacks
- Item value fields (VAL1-5) fully implemented for all documented uses

## v0.97 — 2026-04-04

### Combat System
- **ATTACK/KILL <monster>** — original format output: `[ToHit: X, Roll: Y] Hit!/Miss./Excellent Hit!`
- **Damage severity tiers**: Puny, Grazing, Insignificant, Minor, Passable, Good, Masterful, Grisly, Severe, Ghastly
- **Attack verbs by weapon type**: swings (slash), thrusts (pierce/pole), slashes (claw)
- **Body part targeting**: head, body, right/left arm, right/left leg, back; animal parts for non-humanoids
- **Weapon elemental crits (VAL3)**: 10-50% chance heat/cold/electric bonus damage with VAL5 max
- **Racial slayer weapons (VAL3 21-32)**: bonus damage vs Undead, Dragon, Giant, Troll, etc.
- **Weapon poison (VAL4 51-100)**: delivers poison on hit
- **MAGICWEAPON gating**: some monsters require enchanted weapons (level 1/2/3), shows TEXI text
- **Monster guard behavior**: guard monsters intercept attacks on their charge
- **Cry for law (strategy 1-25, 101-125)**: attacking lawful NPCs alerts nearby guards
- **Monster poison/disease/fatigue**: % chance per hit to inflict conditions
- **EXTRABODY**: monsters have 50-100% extra HP not counted toward XP
- **Weather combat modifiers**: rain -10, heavy rain -20, thunderstorm -30, gale -40, hurricane -50 (outdoor only)
- **Arena rooms**: ExtraMod ARENA prevents lethal damage
- **Alignment shifts**: killing evil monsters → more good, good → more evil
- **Hostile monsters** (strategy 301+) auto-attack players entering rooms
- **Monster flee AI**: strategy-based (fight to death 501+, flee when wounded 301-500, animals flee 201-300)
- **Monster special attacks**: SPECUSE/SPECDMG with TEXX text overrides
- **FLEE command**: escape combat (quickness/agility based chance, random exit)
- **Combat stances**: OFFENSIVE (+15/-15), DEFENSIVE (-15/+15), BERSERK (+25/-25, Murg), WARY (-5/+5)
- **Roundtime**: `[Round: X sec]` format, quickness-based
- **Death**: Eternity, Inc. message (matching original text) → DEPART to respawn
- **Real XP table**: 100-level XP/build-point table from GM Manual, build points = 20 + 10*level

### Magic System
- **PREPARE <spell>** then **CAST [target]** — two-step spell casting
- **60+ spells** across 5 schools with mana costs and casting times
- **Offensive**: Flame Bolt, Force Blade, Lightning Bolt, Freezing Sphere, Call Meteor, Chain Lightning, Energy Maelstrom
- **Healing**: Body Restoration I/II/III, Invigoration I/II, Reconstruction, Regeneration
- **Defense**: Mystic Armor (+20), Globe of Protection (+50/+100), Spectral Shield (+20), Spell Shield (+25)
- **Buffs**: Strength I/II/III, Agility I/II/III, Fly, Invisibility, Haste
- **Spellcraft skill checks**: school skill + spellcraft + willpower vs spell level
- **Magic resistance**: monsters can resist spells, elemental immunities apply

### Psionic System
- **PSI <discipline>** then **PROJECT [target]** — two-step psionic projection
- **Mind over Matter**: Kinetic Thrust, Pyrokinetics, Cryokinetics, Electrify, Wall of Force (+25), Force Field (+75), Flight
- **Mind over Mind**: Psychic Blast, Psychic Crush, Terror, Pain, Screen/Shield/Barrier/Fortress defense chain
- **Psi point costs**, skill checks (Psionics + school skill + willpower), psi resistance

### World Systems
- **SKIN <dead monster>** — weighted random drops from SkinItem definitions with SkinAdj
- **Container traps (VAL4)**: 13 trap types (needle, gas, acid, blade, explosive, glyph spells) trigger on OPEN
- **Highlander BLEND**: hide in mountain/cave terrain (race-restricted)

### Race-Specific Emotes
- **Drakin**: flick tongue, bare teeth, spread/fold wings, swish tail
- **Aelfen**: rub ears · **Highlander**: pull beard
- **Wolfling**: scratch behind ear, bare fangs, chase tail, scent air, whine, droop tail
- **23 self-emotes**: fume, squint, hum, sneeze, crack knuckles, bat eyelashes, bounce, strike pose, and more

### Infrastructure
- **PlayerMessageFunc**: targeted messages from background tasks (combat, aggro)
- **Monster combat tick**: integrated into 3-second behavior loop
- **Combat disengagement**: moving rooms, fleeing, or target dying clears combat state
- **Character name validation**: reserved-word filter (monster names, profanity, game terms)

## v0.96 — 2026-04-04

### Monster System
- **Monster spawning**: Monsters appear in room descriptions from original MonsterList data
- **TEX1-4 ambient text**: Monsters emit random flavor text on a timer based on Speed
- **Wandering**: Non-hostile monsters (strategy < 301) wander between rooms via exits
- **TEXG/TEXE/TEXM**: Custom text for spawn, room entry, and movement (direction appended)
- **Examine monsters**: `exam skeleton` shows monster description
- **Emote targeting**: Monsters targetable by all emotes (`point skeleton`, etc.)
- **@spawn/@genmon fixed**: GM commands now actually create monster instances
- **@genmon creates sedated** (inactive), **@spawn creates active** (will act)
- **Speed-based ticking**: Speed 1 = every 3s, Speed 3 (default) = every 9s

### New Emotes (30+)
- **lick, nibble, bark, claw, curse, duck, hiss, hold, hula, jig, moan, massage, pinch, play, purr, roar, snarl, snuggle, wag, wait, write, yowl, thump, applaud, peer, grunt, dip, handraise, handshake, headshake, pick, gesture**
- **Self-targeting overrides**: `spit me` → drool, `lick me` → lick lips, `laugh me` → laugh at self, `kick me` → kick self, `thump me` → thump head
- **KISS body parts**: head, nose, lips, ears, neck, chest, hand, navel, leg, knee, feet
- **Submit-gated**: lips, navel, leg, knee, feet require target to be submitting

### Submit System
- **SUBMIT/UNSUBMIT** commands — accept intimate emotes from other players
- **LICK** behavior changes: non-submitted → passionate kiss, submitted → full body lick
- Moving to a new room automatically clears submit state

### Infrastructure
- **CEVENT ECHO delivery**: Script ECHO messages from cyclic events now broadcast to players in rooms
- **RoomBroadcast callback**: Engine background tasks (monsters, CEVENTs) can push messages to rooms via WebSocket
- Monster article handling: "an orc" vs "a skeleton", unique monsters without article

## v0.95 — 2026-04-03

### Lock & Unlock
- **LOCK/UNLOCK commands** — match KEY items via Val3, proper feedback for missing/wrong keys
- Latched items block open; locked items block passage

### Ordinal Targeting
- **"2 gate", "other gate", "second gate"** — target the Nth matching item in a room
- Works across all 19 item-matching functions (get, drop, look, open, close, wield, wear, etc.)

### Mechanoid Emote
- **EMOTE/UNEMOTE** — Mechanoid racial ability to toggle emotional state (race 7 only)
- **ACT** remains the general-purpose roleplaying command for all races

### Verb Aliases & Commands
- **ORDER** as BUY synonym
- **UNLIGHT/IGNITE** as EXTINGUISH/LIGHT aliases
- **QUAFF** as DRINK alias, **SHOUT** as YELL alias, **PLACE** as DROP alias
- **RECALL** with no args runs room-level IFVERB RECALL -1 scripts
- **ACTBRIEF/RPBRIEF** toggle commands
- **POUR** verb stub
- **A** as shorthand for ACT (freeform roleplay)

## v0.94 — 2026-04-03

### Script Engine
- **226 named global variables** (DANWATER, TECHSWITCH, etc.) loaded from VARIABLE definitions, synchronized across machines
- **MUL/DIV/MOD** arithmetic operations in scripts
- **GENMON/ZAPMON** — spawn and remove monsters from scripts
- **GFLAG** — set flags for all players in a room
- **RELOGIN/IFLOGIN** — force player room on login
- **IFFULLDESC/IFIN** — new conditional blocks
- **NEWPUT** — place items inside containers from scripts
- **DUMMY1-5** volatile scratch variables
- **PVAL 0-19** persistent system-wide variables (MongoDB-backed)
- **CEVENT system** — cyclic timed events parsed and executing on 3-second ticks
- **Implicit ENDIF** — parser handles missing ENDIF when new verb block starts (matches original engine)
- **Room-level IFVERB blocks** now fire from doItemInteraction and position commands
- **Direction verb resolution** — "O" resolves to "OUT" for IFPREVERB matching, all abbreviations expanded

### Parsing
- **REGIONDEF** expanded: DepartRoom, Weather, Treasure, Teleport, Summoning, spell modifiers, MineAdj
- **MONEYDEF** — multi-currency per region definitions
- **FORAGEDEF/MINDEF** — full forage and mining definitions parsed
- **Room REGION field** parsed and accessible in scripts
- **Monster psionics** — PSI, PSIUSE, PSISKILL, PSIRESIST, PSILEVEL, DISCIPLINE
- **Monster combat** — IMMUNITY (10 types), WEAPON, WEAPONPLUS, MAGICWEAPON, SPECUSE/SPECDMG, EXTRABODY, FATIGUE

### Food & Drink
- **Bite/sip tracking** — PARAMETER1 defines bites for food, sips for liquid. VAL2 tracks remaining. Items removed only when empty.
- **Spell on consume** — ITEMVAL3 triggers spell effect on first bite (e.g., Mindlink from thesnia leaf)
- **Item scripts run on eat** — root-level IFVAR blocks execute before checking spell effects

### Items & Verbs
- **Store adjective in ADJ3** — bought items place variety adjective in last slot, leaving ADJ1/ADJ2 for crafting
- **Worn/wielded item visibility** — all verbs (touch, sniff, wave, emotes) now find inventory, worn, and wielded items
- **SIT/LAY/KNEEL/STAND trigger IFVERB scripts** — position commands run room-level verb scripts
- **CLIMB handles MOVE** — non-portal items with IFPREVERB CLIMB + MOVE now work
- **SNIFF/SMELL/LISTEN** try item interaction first, fall back to emote
- **@peek named variables** — `@peek DANWATER` works, plus PVAL and full getVar fallback
- **@yank fix** — updates live session player, not just database

### Telepathy & Spells
- **Mindlink (#403)** — first working spell, grants 1hr telepathy via food consumption
- **Telepathy system** — THINK broadcasts to telepathy-enabled players only
- **Ephemeral innate telepathy** — race 8 starts with telepathy active
- **TELEPATHY command** — toggle on/off

### Original Fidelity
- Global messages: "has just entered/left the Realms"
- WHO: 4-column grid with "There are N adventurers in the Realms."
- EAT: "You take a bite of X. (N bites remaining)"
- Death: Eternity, Inc. message + DEPART stub
- HP → BP (Body Points) throughout UI
- Article fix: "an axe" not "a axe"

### Admin & Infrastructure
- Event Monitor: real-time admin WebSocket feed
- Session Capture: record, view, download gameplay
- Capture delete with confirmation
- /healthz endpoint for vibectl
- Backend connection indicator
- Admin Players/Users sorted by recency with relative timestamps

## v0.93 — 2026-04-03

### New Commands
- **HIDE/SNEAK**: Stealth system — conceal yourself and move while hidden
- **FLY/ASCEND/DESCEND/LAND**: Flight system for Drakin race and spell-granted flight
- **MARK**: Set teleport anchor points (1-10) for future spell use
- **BALANCE**: View bank account balance
- **SPELL**: List all known spells
- **THINK**: Telepathy broadcast to all players
- **CANT**: Covert message requiring Legerdemain skill
- **UNDRESS**: Remove outermost worn item
- **UNPROMPT**: Toggle off prompt indicators
- **VERSION/CREDITS**: Game info

### Combat & Spell Stubs (40+ verbs)
- All combat verbs stubbed with TODO: ATTACK, KILL, ADVANCE, RETREAT, GUARD, BACKSTAB, BITE, AVOID
- Combat stances: BERSERK, FRENZY, DEFENSIVE, OFFENSIVE, WARY, NORMAL
- Spell verbs: INVOKE, PREPARE, CHANT, COMMAND, MASTER
- Ranged: NOCK, LOAD, SPECIALIZE
- Skills: DISARM, STEAL, STALK, TEACH, SELFTRAIN, UNLEARN, ANOINT, TRAP, SURVEY, SPLIT
- Racial: BLEND, CALL, TRANSFORM, MOLD, DISGUISE
- Social: SUBMIT, ARREST, ENROLL, INITIATE, JOIN, FOLLOW, LEAVE, DISBAND
- Other: TEND, BREAK, PUT, FILL, SKIN, REPAIR, WORK

### Script Variables Expansion
- Physical attributes: HEI/HEIT, WEI/WEIT, AGE/AGET
- Form states: WOLFFORM, SLIMEFORM, OTHERFORM, UNDEAD, DISGUISED
- Status: SLEEPING, SUBMITTING, ROUNDTIME, SPELLNUM, POSITION
- Wealth: WEALTH (total copper)
- Room: WILDERNESS, ASTRAL, TERRAIN (numeric), OBJWEIGHT

### Player System
- New fields: Height, Weight, Age (with true variants), Marks, PreparedSpell
- Form states: WolfForm, SlimeForm, Disguised
- Status: Sleeping, Submitting, Undead

### Fixes
- **Article fix**: "an axe" instead of "a axe" — auto-detects vowel sounds
- **Body Points**: All "HP" references changed to "BP" throughout UI
- **GO command**: Non-portal items with IFPREVERB GO scripts now work (e.g., stairways)
- **CLEARVERB + MOVE**: Scripts that block default GO but provide MOVE now work correctly
- **Session Capture**: Fixed null array bug preventing line recording
- **Copy-paste**: Terminal text now selectable on Windows
- **Stop Recording**: Fixed stale WebSocket reference

### Infrastructure
- `/healthz` endpoint for vibectl health monitoring
- Backend connection indicator in nav bar
- Character list shows "Connecting to server..." when backend is unavailable
- Create Character button hidden until backend responds
- Event Monitor disconnect messages

## v0.92 — 2026-04-03

### Admin Event Monitor
- Real-time WebSocket event feed for admin monitoring
- Categories: system, time, monster, script, world state, weather
- Background time cycle publishes hour/day/night transitions
- Category filter, clear button, auto-scroll

### Session Capture
- Record game sessions to MongoDB
- Start/stop from modal overlay during gameplay
- View previous captures with color-coded output
- Download as .txt file
- Auto-stops on disconnect

## v0.91 — 2026-04-03

### Script Engine Expansion
- **80+ variables** supported in IFVAR conditions: player stats (STR/AGI/CON/QUI/WIL/PER/EMP), resources (BODYPOINTS/MANAPOINTS/PSIPOINTS/FATPOINTS), state (DEAD/FLYING/KNEELING/HIDDEN), organization (ORG/ORGRANK/ALIGN), room info (RNUM/OUTDOOR/TERRAIN/PLRSINROOM/MONINROOM), time (TIM/DAY/NIGHT/DATE/MONTH/YEAR), weather (WEA), exits (EXITN/EXITS/etc.), flags (FLAG1-4), skills (SKILL0-35)
- **IFSAY** script blocks now execute on player speech (enables NPC dialogue and quest triggers)
- **IFCARRY** condition checks if player carries a specific item
- **IFNOITEM** condition (negation of IFITEM)
- **IFTOUCH** block execution for touch-type verbs
- **AFFECT** action switches script context to another room for multi-room effects
- **RANDOM** action for random number generation in scripts
- **DAMAGEPLR** action for environmental damage (traps, falls)
- **STRCVT** action with %0-%9 string substitution in ECHO text
- **POSITION** action forces player position from scripts
- **String substitution** expanded: %s (he/she), %i (him/her), %h (his/her), %p (group name), %m (monster), %c (newline), %0-%9 (STRCVT)

### World Systems
- **In-game clock and calendar**: 343-day year, 12 months, day/night cycle (1 real minute = 1 game hour)
- **Monster spawning**: Monsters populate rooms from MonsterList data, displayed in room descriptions
- **Weather system**: 15 weather states per region, shown in outdoor room descriptions
- **Room lighting**: Dark rooms require light sources; DAY_LIGHT rooms dark at night

### New Commands
- **DRINK/SIP**: Consume liquid/food items
- **LIGHT/EXTINGUISH**: Light and douse LIGHTABLE items for dark room navigation
- **FLIP**: Toggle FLIPPED/UNFLIPPED state on FLIPABLE items
- **LATCH/UNLATCH**: Toggle latched state; latched items can't be opened
- **DEPOSIT/WITHDRAW**: Banking in BANK rooms
- **TRAIN**: Skill training in rooms with TRAINING definitions (36 skills)
- **MINE**: Mining in MINEA/B/C rooms (stub)
- **FORAGE**: Foraging in wilderness terrain (stub)
- **CAST**: Spell casting with 150+ registered spells across 5 schools (stub behavior)
- **CRAFT/FORGE/SMELT/WEAVE/DYE/BREW/ANALYZE**: Crafting stubs

### Spell Registry
- 150+ spells registered across 5 schools: Conjuration (100-144), Enchantment (200-250), Necromancy (301-356), General Magic (400-415), Druidic (500-538)
- CAST command recognizes all spells by name prefix

### Monster System
- Expanded MonsterDef with 20+ new fields: alignment, magic resistance, mana, spell use, poison/disease, skin items, text overrides, guard, stealable, eternal, discorporate
- All monster fields parsed from scripts
- Passive monster spawning in rooms

### Player Systems
- Organization/Guild fields (ORG, ORGRANK)
- Alignment tracking
- Build points
- Banking (BankGold/Silver/Copper)
- Known spells registry
- Transient flags (FLAG1-4, reset on room entry)
- 36 named skills with training system

### Training System
- TRAINING room definitions parsed from scripts
- TRAIN command with skill listing, level caps, and gold cost
- Cost formula: (current_level+1)^2 * 10 copper per level

## v0.9 — 2026-04-03

### Game Engine
- Script parser loads 2273 rooms, 1990 items, 297 monsters from original .SCR files
- Fixed description parser: `*DESCRIPTION_START ITEM READ/EXAM` no longer overwrites room descriptions
- Fixed item deduplication: later script definitions correctly override earlier ones (ANTI.SCR placeholders replaced)
- Fixed script block parser: missing ENDIF no longer consumes subsequent item definitions
- Seasonal scripts (ASCRIPT/WSCRIPT/SSCRIPT/PSCRIPT) skipped; base definitions used
- Formatted text preservation: descriptions with leading whitespace or blank lines retain formatting (poems, maps, etc.)

### Movement & Portals
- Portal traversal via GO command with VAL2 destination
- PORTAL_CLIMB, PORTAL_UP, PORTAL_DOWN, PORTAL_OVER, PORTAL_THROUGH support
- CLIMB verb for climbable portals
- Closed/locked portals block passage
- Directional look (LOOK N, LOOK NORTH, etc.) shows adjacent room description

### Script Execution
- IFENTRY blocks execute on room entry (movement, login, character creation)
- IFVAR conditions: INTNUM, ITEMBIT, ITEMVAL, ITEMADJ, LEV, RAC, SKILL, MISTFORM
- IFITEM conditions: check item open/closed/locked/unlocked state
- IFPREVERB/IFVERB execution on item interactions
- Actions: ECHO (PLAYER/ALL/OTHERS), EQUAL, ADD, SUB, NEWITEM, GMMSG, CLEARVERB, MOVE, SHOWROOM
- Actions: LOCK, UNLOCK, OPEN, CLOSE, REMOVEITEM, SETITEMVAL
- Nested conditional blocks supported
- Script text placeholders: %N, %n, %a, %h, %e, %o

### Commands
- Movement: N/S/E/W/NE/NW/SE/SW/UP/DOWN/OUT, GO, CLIMB
- Looking: LOOK, EXAMINE, INSPECT, LOOK IN/ON/UNDER/BEHIND, directional LOOK
- Items: GET, DROP, INVENTORY, WIELD, UNWIELD, WEAR, REMOVE
- Containers: OPEN, CLOSE
- Interaction: PULL, PUSH, TURN, RUB, TAP, TOUCH, SEARCH, DIG, RECALL
- Communication: SAY ('), WHISPER, YELL, RECITE, EMOTE
- Commerce: BUY, SELL (1482 store items across 20+ shops)
- Social: 60+ emotes (SMILE, BOW, KICK, etc.) with targeted second-person messages
- Info: STATUS, HEALTH, WEALTH, SKILLS, WHO, HELP, ASSIST
- Position: SIT, STAND, KNEEL, LAY
- Roleplay: ACT, EMOTE, RECITE
- READ items with room-scoped descriptions

### Multiplayer
- WebSocket-based real-time gameplay
- Cross-machine coordination via MongoDB Change Streams
- Player presence synced across multiple Fly.io machines
- Room state changes (item state, drops, pickups, script mutations) coordinated
- Emote targeting: second-person messages to target, third-person to room
- Room broadcasts, global broadcasts, GM broadcasts, whispers
- WebSocket reconnection with exponential backoff

### Characters & Auth
- Google OAuth login with 30-day JWT sessions
- Character creation with 8 races, stat rolling, name validation
- Character name uniqueness enforced (no logging into others' characters)
- Session persistence across server restarts
- Starting gear granted via IFENTRY scripts at room 201

### Commerce
- BUY command with store items (STOREITEM parsed from scripts)
- Efficient currency deduction (copper first, then silver, then gold with change)
- SELL command in rooms with BUY_ARMOR/BUY_SKINS/BUY_JEWELRY modifiers
- Price display as gold/silver/copper

### Admin Interface
- Rooms tab: searchable index with detail view, clickable exits, enriched items
- Items tab: searchable index with full properties, type/slot expansion
- Monsters tab: searchable index with combat stats and properties
- Players tab: character detail, GM toggle, account reassignment
- Users tab: account detail, admin toggle, character list
- Logs tab: filterable game event log with hyperlinked players/users
- Exact number match sorting in search results

### Logging
- MongoDB-backed game log with 90-day TTL
- Events: user login/logout, character game enter/exit, character creation, GM grant/revoke
- Log entries include user name, email, account ID
- Admin logs UI with event/player filtering

### Security
- Admin auth required for GM toggle, game world API endpoints
- WebSocket origin validation against frontend URL
- 64KB WebSocket message size limit
- Character name validation (alpha + ' + -, max 20 chars)
- Race/gender input validation
- Command input truncated to 500 characters

### Infrastructure
- Go + gorilla/mux backend, React 19 + TypeScript + Vite + Tailwind 4 frontend
- MongoDB Atlas for persistence
- Fly.io production deployment (2 machines, ord region)
- Custom domain: lofp.metavert.io
- Multi-stage Docker build (14MB production image)
- Separate dev/prod configs
