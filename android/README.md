# Arena for Android

Uses [ebitenmobile](https://ebitengine.org/en/documents/mobile.html) to bind
`game/arena/mobile` into an AAR, then packages it in a native Android Activity.
The desktop entry point remains `go run ./game/arena`; both use
`game/arena/engine`.

## Build

Requirements:
- Go (the version required by the root `go.mod`), Python 3, and JDK 17 or 21.
- Android SDK platform 36, build tools, and NDK **28.2.13676358**.
- A legally obtained Master of Magic data ZIP or directory containing LBX files.
- Internet access on the first build to download Go/Gradle dependencies.

```sh
export JAVA_HOME=/path/to/jdk-21
export PATH="$JAVA_HOME/bin:$PATH"
export ANDROID_HOME="$HOME/Android/Sdk"
"$ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager" \
  'platforms;android-36' 'build-tools;35.0.0' 'ndk;28.2.13676358'
./android/build.sh '/path/to/Master of Magic.zip'
# Or: ./android/build.sh /path/to/LBX-directory
```

Without an argument, the script uses `Master of Magic.zip` at the repository
root. It stages only LBX files in an ignored embedded archive. The existing
embedded SoundFont is also available for music. Proprietary resources, AARs,
and APKs are not committed. Do not redistribute bundled game data without
permission.

Output: **`dist/arena-debug.apk`** (debug-signed, not a Play Store release).

```sh
"$ANDROID_HOME/platform-tools/adb" install -r dist/arena-debug.apk
"$ANDROID_HOME/platform-tools/adb" shell am start \
  -n org.masterofmagic.arena/.MainActivity
```

## Signed release APK

After building the Go AAR with `android/build.sh`, run:

```sh
./android/sign-release.sh
```

This builds the non-debug release variant, aligns it, signs it using
`~/keystore/magic-arena.jks` (alias `arena`), and verifies the signature.
It prompts for the keystore password without echoing it. Output:
**`dist/arena-release.apk`**. Set `ARENA_KEYSTORE` or `ARENA_KEY_ALIAS` to
use a different keystore or alias. Use JDK 17 or 21 as above.

For non-interactive signing, set `ARENA_STORE_PASSWORD_FILE` to a protected
file containing the password on its first line. If the private key has a
different password, also set `ARENA_KEY_PASSWORD_FILE`. Never commit these
files. `ANDROID_BUILD_TOOLS` can override the default SDK build-tools/36.0.0
directory.

A release-signed APK cannot update a debug-signed installation. Uninstall the
debug app first (this deletes its app data). Keep the same release key for
future updates and increment `versionCode` in `android/app/build.gradle`.

Supports Android 7.0/API 24+ on ARM64 phones/tablets and x86-64 emulators.
Override `ANDROID_TARGET` with an ebitenmobile target list if needed; other
ABIs may require a compatible NDK. `ANDROID_NDK_HOME` can override the pinned
NDK (preview NDK 30 is incompatible with the current Oto/Oboe dependency).

The app is fullscreen landscape. The shop uses a fixed 1200×1050 logical
canvas so widgets do not shrink according to device pixel density; combat
uses the engine's original logical resolution. The existing UI supports
finger taps; combat maps a short tap to left-click and a long press to
right-click. Activity pause/resume suspends/resumes the EbitenView. No network
permission is needed.

Build/signature checks do not replace gameplay testing on a real device.
