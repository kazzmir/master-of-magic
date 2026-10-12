#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_SDK="$HOME/Android/Sdk"
SDK_DIR="${ANDROID_SDK_ROOT:-${ANDROID_HOME:-$DEFAULT_SDK}}"
EMULATOR_BIN="$SDK_DIR/emulator/emulator"
ADB_BIN="$SDK_DIR/platform-tools/adb"
APK_PATH="$ROOT_DIR/dist/arena-debug.apk"
PACKAGE_NAME="org.masterofmagic.arena"
ACTIVITY_NAME="org.masterofmagic.arena.MainActivity"
BUILD_APK=false
AVD_NAME=""
LIST_AVDS=false
NO_START=false

usage() {
    cat <<'EOF'
Usage: scripts/run-emulator.sh [options]

Starts or reuses an Android emulator, installs dist/arena-debug.apk,
and launches Master of Magic Arena.

Options:
  --build       Rebuild the APK using android/build.sh before installing.
  --apk PATH    Install a different arena APK (cannot be combined with --build).
  --avd NAME    Use the specified AVD instead of the first available AVD.
  --list-avds   Print available AVD names and exit.
  --no-start    Require an already running emulator instead of starting one.
  --help        Show this help text.
EOF
}

require_tool() {
    local tool_path="$1"
    local description="$2"

    if [[ ! -x "$tool_path" ]]; then
        echo "Missing $description at: $tool_path" >&2
        exit 1
    fi
}

available_avds() {
    "$EMULATOR_BIN" -list-avds | sed '/^[[:space:]]*$/d'
}

first_running_emulator() {
    "$ADB_BIN" devices | awk '/^emulator-[0-9]+[[:space:]]+device$/ { print $1; exit }'
}

wait_for_boot() {
    local device_serial="$1"

    "$ADB_BIN" -s "$device_serial" wait-for-device >/dev/null

    until [[ "$("$ADB_BIN" -s "$device_serial" shell getprop sys.boot_completed 2>/dev/null | tr -d '\r')" == "1" ]]; do
        sleep 2
    done

    "$ADB_BIN" -s "$device_serial" shell input keyevent 82 >/dev/null 2>&1 || true
}

start_emulator() {
    local avd_name="$1"

    echo "Starting emulator for AVD: $avd_name" >&2
    nohup "$EMULATOR_BIN" -avd "$avd_name" >/tmp/arena-emulator.log 2>&1 &

    for _ in {1..60}; do
        local serial
        serial="$(first_running_emulator)"
        if [[ -n "$serial" ]]; then
            echo "$serial"
            return 0
        fi
        sleep 2
    done

    echo "Timed out waiting for the emulator to appear in adb." >&2
    echo "See /tmp/arena-emulator.log for emulator output." >&2
    exit 1
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --build)
            BUILD_APK=true
            shift
            ;;
        --apk)
            if [[ $# -lt 2 ]]; then
                echo "--apk requires a value." >&2
                exit 1
            fi
            APK_PATH="$2"
            shift 2
            ;;
        --avd)
            if [[ $# -lt 2 ]]; then
                echo "--avd requires a value." >&2
                exit 1
            fi
            AVD_NAME="$2"
            shift 2
            ;;
        --list-avds)
            LIST_AVDS=true
            shift
            ;;
        --no-start)
            NO_START=true
            shift
            ;;
        --help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown argument: $1" >&2
            usage >&2
            exit 1
            ;;
    esac
done

if [[ "$LIST_AVDS" == true ]]; then
    require_tool "$EMULATOR_BIN" "Android emulator"
    available_avds
    exit 0
fi

if [[ "$BUILD_APK" == true ]]; then
    if [[ "$APK_PATH" != "$ROOT_DIR/dist/arena-debug.apk" ]]; then
        echo "--build cannot be combined with a custom --apk path." >&2
        exit 1
    fi
    echo "Building arena debug APK"
    ANDROID_HOME="$SDK_DIR" "$ROOT_DIR/android/build.sh"
fi

if [[ ! -f "$APK_PATH" ]]; then
    echo "APK not found at $APK_PATH. Run android/build.sh or use --build." >&2
    exit 1
fi

require_tool "$ADB_BIN" "adb"
DEVICE_SERIAL="$(first_running_emulator)"

if [[ -z "$DEVICE_SERIAL" ]]; then
    if [[ "$NO_START" == true ]]; then
        echo "No running emulator found and --no-start was specified." >&2
        exit 1
    fi

    require_tool "$EMULATOR_BIN" "Android emulator"
    if [[ -z "$AVD_NAME" ]]; then
        AVD_NAME="$(available_avds | awk 'NR == 1 { first = $0 } END { print first }')"
    fi

    if [[ -z "$AVD_NAME" ]]; then
        echo "No AVDs are available. Create an API 24+ ARM64 or x86-64 AVD in Android Studio." >&2
        exit 1
    fi

    DEVICE_SERIAL="$(start_emulator "$AVD_NAME")"
fi

echo "Waiting for emulator boot: $DEVICE_SERIAL"
wait_for_boot "$DEVICE_SERIAL"

echo "Installing APK"
"$ADB_BIN" -s "$DEVICE_SERIAL" install -r "$APK_PATH"

echo "Launching app"
"$ADB_BIN" -s "$DEVICE_SERIAL" shell am start -n "$PACKAGE_NAME/$ACTIVITY_NAME" >/dev/null

echo "App launched on $DEVICE_SERIAL"
