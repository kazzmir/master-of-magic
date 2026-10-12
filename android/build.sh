#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
: "${ANDROID_HOME:?Set ANDROID_HOME to your Android SDK (platform 36 and NDK required)}"
# Oto 3.4's bundled Oboe is incompatible with preview NDK 30 headers.
export ANDROID_NDK_HOME="${ANDROID_NDK_HOME:-$ANDROID_HOME/ndk/28.2.13676358}"
DATA="${1:-$ROOT/Master of Magic.zip}"
TARGET="${ANDROID_TARGET:-android/arm64,android/amd64}"
mkdir -p game/arena/mobile/assets android/app/libs dist

# Include only game resources, not the original DOS executables.
python3 - "$DATA" "$ROOT/game/arena/mobile/assets/game.zip" <<'PY'
import pathlib, sys, zipfile
source, output = map(pathlib.Path, sys.argv[1:])
files = {}
if source.is_dir():
    files = {p.name.upper(): p.read_bytes() for p in source.iterdir()
             if p.is_file() and p.suffix.lower() == '.lbx'}
else:
    with zipfile.ZipFile(source) as archive:
        for name in archive.namelist():
            if name.lower().endswith('.lbx'):
                files[pathlib.PurePosixPath(name).name.upper()] = archive.read(name)
required = {'MAGIC.LBX', 'BACKGRND.LBX', 'INTRO.LBX', 'CITYSCAP.LBX',
            'FIGURE10.LBX', 'UNITS1.LBX', 'UNITS2.LBX', 'TERRAIN.LBX'}
missing = required - files.keys()
if missing:
    sys.exit('Missing game resources: ' + ', '.join(sorted(missing)))
with zipfile.ZipFile(output, 'w', zipfile.ZIP_DEFLATED) as archive:
    for name, content in sorted(files.items()):
        archive.writestr(name, content)
print(f'Bundled {len(files)} LBX files')
PY

# Use exactly the engine version selected by this module.
VERSION="$(go list -m -f '{{.Version}}' github.com/hajimehoshi/ebiten/v2)"
GOBIN="$ROOT/android/.tools" go install "github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@$VERSION"
"$ROOT/android/.tools/ebitenmobile" bind -target "$TARGET" -androidapi 24 \
    -javapkg org.masterofmagic.arena -trimpath -ldflags='-s -w' \
    -o "$ROOT/android/app/libs/arena.aar" ./game/arena/mobile

cd android
./gradlew --no-daemon assembleDebug
cp app/build/outputs/apk/debug/app-debug.apk ../dist/arena-debug.apk
printf '\nAPK: %s/dist/arena-debug.apk\n' "$ROOT"
