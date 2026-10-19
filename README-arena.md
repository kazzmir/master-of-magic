## Desktop test armies

Start directly in army preparation with a preconfigured army:

```sh
go run ./game/arena -test 3
```

| Preset | Units | Gold | Difficulty | Level | Mana | Books per realm |
| --- | ---: | ---: | --- | ---: | ---: | ---: |
| 1 | 2 | 500 | Easy | 1 | 20 | 1 (2 realms) |
| 2 | 5 | 2,500 | Normal | 3 | 50 | 2 (3 realms) |
| 3 | 10 | 10,000 | Hard | 6 | 100 | 3 (4 realms) |
| 4 | 18 | 50,000 | Impossible | 10 | 200 | 4 (5 realms) |
| 5 | 30 | 200,000 | Impossible | 15 | 400 | 5 (5 realms) |

Each preset includes mixed units with enchantments, some damaged units for testing
healing, and combat spells from its magic realms. Higher presets add experience
and weapon upgrades; the largest armies also exercise list scrolling.
Omit `-test` (or use `-test 0`) for normal startup with the difficulty picker.

## Android arena

Build an installable APK of `./game/arena` with `ebitenmobile`:

```sh
./android/build.sh '/path/to/Master of Magic.zip'
```

See [Android build instructions](android/README.md) for prerequisites and installation.

