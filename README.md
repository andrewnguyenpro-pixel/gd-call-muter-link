# GD Call Muter

Mutes your Instagram call and deafens Discord when your Geometry Dash run passes a percentage you choose.

**Downloads:** see the [latest release](../../releases/tag/latest).

| File | What it is |
| --- | --- |
| `GDCallMuter.exe` | The Windows app |
| `GD-Call-Muter-Windows.zip` | The Windows app + browser extension |
| `GD-Call-Muter-Mac.zip` | The Mac app + browser extension |
| `nil.gd-call-muter-link.geode` | The Geode mod (Windows + Mac, GD 2.2091, Geode 6) |

## How it fits together
- **Geode mod** (`src/`): reads your exact level % inside the game and shares it on `127.0.0.1:47321`.
- **App** (`windows-app/`, `mac-app/`): watches that %, and when you pass your number it mutes / deafens.
  Instagram and Discord each have their own on/off switch and %.
- **Browser extension** (`extension/`): mutes the instagram.com tab and clicks Deafen on discord.com.
  Load it in `chrome://extensions` → Developer mode → Load unpacked.

## Setup
1. Put the `.geode` file in Geometry Dash's `geode/mods` folder.
2. Run the app.
3. Load the browser extension, then refresh your Instagram and Discord tabs.
4. Mac only: to deafen the Discord app, allow GD Call Muter in System Settings → Privacy & Security → Accessibility.
