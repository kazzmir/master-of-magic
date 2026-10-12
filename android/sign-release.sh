#!/usr/bin/env bash
# Sign the release APK without storing passwords in Gradle files or arguments.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SDK="${ANDROID_SDK_ROOT:-${ANDROID_HOME:-$HOME/Android/Sdk}}"
KEYSTORE="${ARENA_KEYSTORE:-$HOME/keystore/magic-arena.jks}"
ALIAS="${ARENA_KEY_ALIAS:-arena}"
INPUT="$ROOT/android/app/build/outputs/apk/release/app-release-unsigned.apk"
OUTPUT="$ROOT/dist/arena-release.apk"

if [[ ! -f "$KEYSTORE" ]]; then
    echo "Keystore not found: $KEYSTORE" >&2
    exit 1
fi

# Prefer a known stable version; allow overriding for other SDK installations.
TOOLS="${ANDROID_BUILD_TOOLS:-$SDK/build-tools/36.0.0}"
for TOOL in zipalign apksigner; do
    if [[ ! -x "$TOOLS/$TOOL" ]]; then
        echo "Missing $TOOLS/$TOOL; set ANDROID_BUILD_TOOLS to your build-tools directory." >&2
        exit 1
    fi
done

if [[ -n "${ARENA_STORE_PASSWORD_FILE:-}" ]]; then
    STORE_PASS="file:$ARENA_STORE_PASSWORD_FILE"
else
    if [[ ! -t 0 ]]; then
        echo "Run this script interactively, or set ARENA_STORE_PASSWORD_FILE to a protected password file." >&2
        exit 1
    fi
    read -r -s -p 'Keystore password: ' ARENA_STORE_PASSWORD
    printf '\n'
    export ARENA_STORE_PASSWORD
    STORE_PASS='env:ARENA_STORE_PASSWORD'
fi

# PKCS12 keystores normally use the same password for the private key.
KEY_PASS="$STORE_PASS"
if [[ -n "${ARENA_KEY_PASSWORD_FILE:-}" ]]; then
    KEY_PASS="file:$ARENA_KEY_PASSWORD_FILE"
fi

# Uses the existing ebitenmobile AAR. Rebuild it with android/build.sh when Go changes.
(cd "$ROOT/android" && ./gradlew --no-daemon :app:assembleRelease)
mkdir -p "$ROOT/dist"
TEMP="$(mktemp -d)"
trap 'rm -rf "$TEMP"; unset ARENA_STORE_PASSWORD' EXIT
"$TOOLS/zipalign" -P 16 -f 4 "$INPUT" "$TEMP/aligned.apk"
"$TOOLS/apksigner" sign --ks "$KEYSTORE" --ks-key-alias "$ALIAS" \
    --ks-pass "$STORE_PASS" --key-pass "$KEY_PASS" \
    --out "$TEMP/signed.apk" "$TEMP/aligned.apk"
"$TOOLS/apksigner" verify --verbose --print-certs "$TEMP/signed.apk"
"$TOOLS/zipalign" -c -P 16 4 "$TEMP/signed.apk"
mv "$TEMP/signed.apk" "$OUTPUT"
printf '\nSigned release APK: %s\n' "$OUTPUT"
