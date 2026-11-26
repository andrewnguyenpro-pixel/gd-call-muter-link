#!/bin/bash
# Builds "GD Call Muter.app" (Apple Silicon + Intel) and zips it with the browser extension.
set -euo pipefail
cd "$(dirname "$0")"
rm -rf build
APP="build/GD Call Muter.app"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
for arch in arm64 x86_64; do
  swiftc -O -swift-version 5 -parse-as-library -target "$arch-apple-macos13.0" \
    GDCallMuter.swift -o "build/GDCallMuter-$arch"
done
lipo -create build/GDCallMuter-arm64 build/GDCallMuter-x86_64 -output "$APP/Contents/MacOS/GDCallMuter"
cp Info.plist "$APP/Contents/Info.plist"
codesign --force --deep -s - "$APP"   # ad-hoc signature so it can run on Apple Silicon
mkdir -p "build/GD Call Muter (Mac)"
cp -R "$APP" "build/GD Call Muter (Mac)/"
cp -R ../extension "build/GD Call Muter (Mac)/Browser extension"
(cd build && ditto -c -k --keepParent "GD Call Muter (Mac)" "GD-Call-Muter-Mac.zip")
ls -la build
