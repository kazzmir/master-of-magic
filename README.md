![build linux](https://github.com/kazzmir/master-of-magic/actions/workflows/build-linux.yml/badge.svg)
![build macos](https://github.com/kazzmir/master-of-magic/actions/workflows/build-macos-m1.yml/badge.svg)
![build windows](https://github.com/kazzmir/master-of-magic/actions/workflows/build-windows.yml/badge.svg)

[Discord Channel](https://discord.gg/xCpHBKrU9c)

# Master of Magic clone

* Source language: golang, https://go.dev/
* Graphics library: ebiten, https://ebitengine.org/
* Master of Magic wiki: https://masterofmagic.fandom.com/

# Online demo

Play a wasm build of this game
https://kazzmir.itch.io/magic

Also try the combat simulator, which lets you test different battle situations
https://kazzmir.itch.io/magic-combat-simulator

# Build:

Extra packages needed for ebiten
https://ebitengine.org/en/documents/install.html

```
$ go mod tidy
$ go build -o magic ./game/magic
```
or
```
$ make
```

# Run:
Put the master of magic lbx files in one of the following places
- in the same directory as the game executable
- in any subdirectory of the directory the game executable is in
- in a zip file in the same directory as the game executable
- in a zip file in any subdirectory of the directory the game executable is in
- in a zip file and replace data/data/data.zip, then rebuild the game. This embeds the data into the executable. You can also put the unzipped lbx files in data/data/

Or use the `-data <path>` command line option to `magic` to specify either a directory that contains the lbx files or the `<path>` can be a zip file of the lbx files.
```
$ ./magic
```
or to run without building first
```
$ go run ./game/magic
```

# Music:

To hear the music you must have a midi synthesizer running separately. On linux and macos you can use [fluidsynth](https://www.fluidsynth.org/). Simply run fluidsynth on its own, and `magic` will automatically connect to it.

# Screenshots:
![new wizard](./images/new-custom-wizard.png)

# Directory layout:
- game/ Contains go code that implements the game functionality
- lib/ Supporting code used to load data/fonts
- util/ Extra utility programs for development purposes (sprite viewer, font viewer, etc)
- data/ Put a zip file with the game data to embed the data in the final binary
- test/ Test programs for executing small pieces of functionality at a time

# Community:

Join the Discord server to discuss this project 

[Discord Channel](https://discord.gg/xCpHBKrU9c)
