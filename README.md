# Grimhold

A text-based dungeon crawler that runs in your terminal. You explore cursed
regions, stumble into encounters, fight in dice-based combat, and record your
kills in a bestiary. Monster data comes live from the
[Open5e API](https://open5e.com) (D&D 5e SRD, no API key); the world map is the
`world.json` file.

Standard library only. Requires Go 1.22 or newer and an internet connection.

## Build and run

```bash
go build -o grimhold ./cmd/grimhold   # build
./grimhold                            # run (or: go run ./cmd/grimhold)
```

Run it from the root of the repository, because it reads `world.json` from the current directory.

## Checking the Open5e API yourself

First check that the API answers and has the expected shape (on Windows PowerShell use `curl.exe`):

```bash
curl https://api.open5e.com/v1/monsters/goblin/
curl "https://api.open5e.com/v1/monsters/?limit=2"
```

You should see `armor_class`, `hit_points`, `challenge_rating` and an `actions` list with
`attack_bonus` and `damage_dice`. The second call should contain `count`, `next` and `results`.

Then try it in the game: `inspect goblin`, `bestiary`, `next`, `travel ashen-marsh`, `explore`.
If a foe from `world.json` is not found, `explore` says so: fix its name in `world.json`
(the name is the monster's slug, for example `giant-spider`).

| Flag | Meaning |
|---|---|
| `--continue` | load `save.json` at startup (starts a new game if there is none) |
| `--world path` | use another world file (default `world.json`) |
| `--save path` | use another save file (default `save.json`) |
| `--api url` | use another Open5e base URL (default `https://api.open5e.com/v1`) |
| `--no-color` | disable ANSI colors (also disabled when `NO_COLOR` is set or the output is not a terminal) |

## Commands

| Command | What it does |
|---|---|
| `help` | list all commands (generated from the command map) |
| `map` | list all regions with their danger level and exits |
| `travel <region>` | move along an exit of the current region (`travel ashen marsh` works too) |
| `look` | describe the current region |
| `explore` | wander the region; may start an encounter |
| `fight` | play **one round** of combat |
| `flee` | try to run away (one roll; failure gives the foe a free hit) |
| `rest` | heal fully, once per region visit |
| `character` | show the character sheet |
| `slain` | list the monsters you have slain, with kill counts |
| `bestiary`, `next`, `prev` | browse the Open5e monsters page by page |
| `inspect <monster>` | show a monster's stat block |
| `save`, `load` | write / read `save.json` |
| `quit` | leave the game |

## How it plays

- **Combat**: attack roll `d20 + attack bonus` against the foe's armor class, then
  damage dice on a hit. The foe answers with one of its own attacks from the API
  (its `attack_bonus` and `damage_dice`) against your armor class.
- **Encounters**: `explore` picks a random foe from the region's `foes`. In regions
  of danger 2 and above it rolls twice and keeps the foe with the higher challenge rating.
- **Victory** gives XP (100 × challenge rating, at least 10), maybe some gold, and
  adds the monster to your slain bestiary. Level `n` to `n+1` needs `n² × 50` XP in total;
  each level gives +8 max HP and +1 attack bonus.
- **Defeat** sends you back to the crossroads with 1 HP. The save file is never touched.
- `save` is refused during a fight. The file is written to a temporary file and then
  renamed, so a crash can never leave a half-written save.

## Project layout

```text
cmd/grimhold/main.go     thin entry point: flags, wiring
internal/world/          world.json loading and validation
internal/open5e/         API client, typed structs, TTL cache with mutex + reaper goroutine
internal/game/           dice, character, combat, encounters, save / load
internal/repl/           command loop and commands
internal/ansi/           hand-rolled ANSI colors
```

The cache stores API responses for 5 minutes in a `map` guarded by a `sync.Mutex`.
A goroutine removes expired entries every minute.

## Bonuses

- **Test suite (+6)**: `go test ./...`. None of the tests touch the network: combat
  with a seeded `rand` and fake monsters, cache expiry with an injected clock, Open5e
  decoding and pagination against an `httptest.Server`, plus world, dice and REPL tests.
- **Faster dungeon (+4)**: when you `travel`, the foes of the new region are downloaded
  in parallel (one goroutine per foe, joined with a `sync.WaitGroup`), so `explore`
  answers instantly from the cache.
- **ANSI colors (+2)**: colored HP bars (green, yellow, red) and danger levels.

## Sample session

```text
Welcome to Grimhold. Type help to see the commands.
grimhold > look
Crossroads
A weathered signpost leans where five roads meet. The wind carries the smell of rot.
Exits: ashen-marsh, whispering-woods, bone-quarry
Danger: safe
grimhold > travel ashen-marsh
You travel to Ashen Marsh. Fog, iron, and things that breathe under the water.
grimhold > explore
Goblin emerges from the shadows! (challenge 1/4, AC 15, HP 7)
Type fight to attack or flee to run.
Goblin         [##########] 7/7
grimhold > fight
You [Rusty Sword]: rolled 3 + 4 = 7 vs AC 15. Miss.
Goblin [Shortbow]: rolled 10 + 4 = 14 vs AC 13. Hit for 3 damage.
You            [#########-] 27/30
Goblin         [##########] 7/7
grimhold > fight
You [Rusty Sword]: rolled 16 + 4 = 20 vs AC 15. Hit for 9 damage.
The Goblin falls! You gain 25 XP.
You find 6 gold.
You            [#########-] 27/30
grimhold > character
Character sheet
  Level    1 (XP 25/50)
  HP       [#########-] 27/30
  Attack   +4
  Armor    13
  Weapon   Rusty Sword (1d8+1)
  Gold     6
grimhold > slain
Slain bestiary
  Goblin               x1
grimhold > save
Game saved to save.json.
grimhold > quit
Farewell, wanderer.
```
