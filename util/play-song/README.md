# Play a music entry

Run from the repository root:

```sh
go run ./util/play-song 115
go run ./util/play-song -lbx magic-data/MUSIC.LBX 115
```

Entries are zero-based. Without `-lbx`, the engine's automatic LBX cache searches
for game data. Put options before the entry number.

This uses the engine's XMI-to-MIDI conversion, soundfont discovery, and audio
player unchanged. A working graphical/audio environment and a soundfont are
required, just as for the game. Playback loops; press Escape, close the window,
or press Ctrl+C in the terminal to stop.
