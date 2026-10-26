# Network combat setup

The client defaults to `wss://magic.jonrafkind.com`. Use
`go run ./game/network-combat -local` to connect to `localhost:5000` instead.
The server address can still be edited in the UI after disconnecting.

## Authentication

The server uses the trimmed contents of `util/combat-simulator/server/key/key.txt`
for multiplayer authentication. The client embeds
`game/network-combat/engine/key/key.txt` and sends its trimmed contents in the
initial auth message. Keep these files identical and rebuild both binaries after
changing the key. Both binaries embed `key/*` and then read `key/key.txt` from
that filesystem: a missing key does not prevent compilation, but causes a clear
runtime error. Empty keys are rejected. Use `wss://` outside local development;
an embedded shared key is not a substitute for per-user authentication.

The creator sets a **uint64 budget per player** (default: 1,000) and chooses the
attacker. The other participant is the defender; the creator can play either
side. Apply shared settings before marking ready.

Each participant selects and saves their own army, books, mana, known spells,
and wizard portrait in the game window. Click the current portrait to open the
three-column portrait picker, then click a portrait to select it:

- Units cost their production cost, or casting cost for summoned units.
- Magic books cost 100 each, with at most 11 books total.
- Mana costs 10 per point (at most 1,000,000 points).
- Known combat spells cost their research cost / 10, rounded down. Non-arcane spells
  require at least one book in the corresponding realm. Spell labels show casting
  mana after the current book discount: each realm book above seven reduces that
  realm's casting cost by 10%, rounded down, matching combat. Purchase prices do
  not change.
- Wizard portraits are free.
- Armies must contain 1–20 units to become ready. Drafts may temporarily be empty.

The game window includes private chat for its participants (Enter or Send,
up to 2,000 bytes per message), with local timestamps. Game chat is separate
from room chat and does not reset readiness; its local history clears on exit.

The UI shows spending and unsaved changes. Unready before editing a loadout.
The server validates all selections and spending, restricts shared settings to
the creator, and resets both readiness flags after any saved setup change.
Revision checks reject stale saves/readiness requests. Lowering the budget does
not silently discard purchases: each player must fix any over-budget loadout.

When both participants are ready, they connect to the combat relay, exchange
keys, and exchange their full approved loadouts (including mana, spells, books,
and portrait). Each peer verifies the other's data against the lobby snapshot
before constructing the two armies and wizards. Combat has no free spells,
mana, or city enchantments from the CLI's demonstration scenario.

## Catalog generation

`protocol/catalog.json` is shared metadata, not artwork or executable game
logic. It lets the server validate purchases without graphics dependencies or
LBX files. Regenerate after changing canonical unit or spell data:

```sh
xvfb-run -a go run ./game/network-combat/protocol/internal/generate \
  -data magic-data -out game/network-combat/protocol/catalog.json
```

`xvfb-run` is only needed on headless Linux. Use matching game assets and the
same catalog on both clients and the server.
