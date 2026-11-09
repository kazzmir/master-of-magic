# Play a music entry

Run from the repository root:

```sh
go run ./util/play-song 115
go run ./util/play-song -lbx magic-data/MUSIC.LBX 115
go run ./util/play-song -polyphony 128 59
```

Music defaults to 256 synth voices to accommodate layered instruments and
release tails. Use `-polyphony` (8–256) to choose a lower CPU/memory budget;
lower budgets can steal voices in dense passages. The engine reserves synth
output headroom and uses stereo-linked soft limiting to prevent float32 output
from exceeding full scale with louder soundfonts.

Entries are zero-based. Without `-lbx`, the engine's automatic LBX cache searches
for game data. Put options before the entry number.

This uses the engine's XMI-to-MIDI conversion, soundfont discovery, and audio
player unchanged. A working graphical/audio environment and a soundfont are
required, just as for the game. Playback loops; press Escape, close the window,
or press Ctrl+C in the terminal to stop. The window shows the current position
and song length, loop restarts, and total audio playback time across loops.
XMIDI loop controllers are honored, so introductions outside a loop play only
once and looping sections restart at their own loop marker. Finite loop counts
are also honored. Songs without a complete loop region repeat as a whole.
