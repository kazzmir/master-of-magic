![build linux](https://github.com/kazzmir/master-of-magic/actions/workflows/build-linux.yml/badge.svg)
![build macos](https://github.com/kazzmir/master-of-magic/actions/workflows/build-macos-m1.yml/badge.svg)
![build windows](https://github.com/kazzmir/master-of-magic/actions/workflows/build-windows.yml/badge.svg)

# Master of Magic clone

* Source language: golang, https://go.dev/
* Graphics library: ebiten, https://ebitengine.org/
* Master of Magic wiki: https://masterofmagic.fandom.com/

# Online demo

Play a wasm build of this game
https://kazzmir.itch.io/magic

# Run/Build:

```
$ go get -u ./game/magic
$ go run ./game/magic
```

```
$ go build -o magic ./game/magic
```
or
```
$ make
```

# Screenshots:
![new wizard](./images/new-custom-wizard.png)

# Directory layout:
- game/ Contains go code that implements the game functionality
- lib/ Supporting code used to load data/fonts
- util/ Extra utility programs for development purposes (sprite viewer, font viewer, etc)
- data/ Put a zip file with the game data to embed the data in the final binary
- test/ Test programs for executing small pieces of functionality at a time
