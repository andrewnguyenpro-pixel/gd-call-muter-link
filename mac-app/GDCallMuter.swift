// GD Call Muter for Mac
// Gets your exact Geometry Dash % from the "GD Call Muter Link" Geode mod
// (http://127.0.0.1:47321/state). Two separate switches, each with its own %:
//   Instagram: mute the instagram.com tab (browser extension), and/or the mic.
//   Discord:   deafen discord.com (extension) and/or the Discord app (keybind).
// The browser extension talks to this app on http://127.0.0.1:47322/wait.

import AppKit
import ApplicationServices
import Carbon.HIToolbox
import Network
import SwiftUI

// MARK: - Colours

extension Color {
    init(hex: UInt32) {
        self.init(red: Double((hex >> 16) & 0xFF) / 255,
                  green: Double((hex >> 8) & 0xFF) / 255,
                  blue: Double(hex & 0xFF) / 255)
    }
}

let cBg = Color(hex: 0x0E1014)
let cCard = Color(hex: 0x171A21)
let cBorder = Color(hex: 0x262A34)
let cText = Color(hex: 0xE7E9EE)
let cMuted = Color(hex: 0x8C93A1)
let cFaint = Color(hex: 0x4B515E)
let cAccent = Color(hex: 0x6E8BFF)
let cGreen = Color(hex: 0x34C77B)
let cRed = Color(hex: 0xF06464)
let cControl = Color(hex: 0x222632)
let cControlH = Color(hex: 0x2C3140)

// MARK: - Key combos

struct KeyCombo: Codable, Equatable {
    var keyCode: UInt16
    var name: String
    var cmd = false
    var ctrl = false
    var opt = false
    var shift = false

    var display: String {
        var s = ""
        if ctrl { s += "⌃" }
        if opt { s += "⌥" }
        if shift { s += "⇧" }
        if cmd { s += "⌘" }
        return s + name
    }

    var cgFlags: CGEventFlags {
        var f: CGEventFlags = []
        if cmd { f.insert(.maskCommand) }
        if ctrl { f.insert(.maskControl) }
        if opt { f.insert(.maskAlternate) }
        if shift { f.insert(.maskShift) }
        return f
    }

    var carbonMods: UInt32 {
        var m: UInt32 = 0
        if cmd { m |= UInt32(cmdKey) }
        if shift { m |= UInt32(shiftKey) }
        if opt { m |= UInt32(optionKey) }
        if ctrl { m |= UInt32(controlKey) }
        return m
    }
}

let fKeys: [UInt16: String] = [
    122: "F1", 120: "F2", 99: "F3", 118: "F4", 96: "F5", 97: "F6", 98: "F7", 100: "F8",
    101: "F9", 109: "F10", 103: "F11", 111: "F12", 105: "F13", 107: "F14", 113: "F15",
    106: "F16", 64: "F17", 79: "F18", 80: "F19", 90: "F20",
]
let namedKeys: [UInt16: String] = [
    49: "Space", 48: "Tab", 36: "Return", 51: "Delete", 123: "←", 124: "→", 125: "↓", 126: "↑",
    115: "Home", 119: "End", 116: "Page Up", 121: "Page Down",
]

func keyName(_ e: NSEvent) -> String? {
    if let f = fKeys[e.keyCode] { return f }
    if let n = namedKeys[e.keyCode] { return n }
    guard let c = e.charactersIgnoringModifiers?.uppercased(), let ch = c.first,
          !ch.isWhitespace, ch.unicodeScalars.allSatisfy({ $0.value >= 0x21 && $0.value < 0xF700 })
    else { return nil }
    return String(ch)
}

// MARK: - Settings

struct Config: Codable {
    var instaOn = true
    var instaThr = 50.0
    var instaTab = true
    var mic = false
    var discOn = true
    var discThr = 50.0
    var web = true
    var desk = false
    var deskKey: KeyCombo? = nil
    var hotKey: KeyCombo? = KeyCombo(keyCode: 98, name: "F7")
    var undeafen = true
    var ignorePrac = true
}

let configURL: URL = {
    let dir = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask)[0]
        .appendingPathComponent("GDCallMuter", isDirectory: true)
    try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
    return dir.appendingPathComponent("config.json")
}()

// MARK: - Browser extension server (127.0.0.1:47322)

final class ExtServer {
    static let shared = ExtServer()
    private let q = DispatchQueue(label: "ext-server")
    private var listener: NWListener?
    private var timer: DispatchSourceTimer?
    private var seq = 0, dseq = 0, iseq = 0
    private var discord = "none", insta = "none"
    private var waiting: [(NWConnection, Date)] = []
    private var lastPoll = Date.distantPast

    var connected: Bool { q.sync { Date().timeIntervalSince(lastPoll) < 25 } }

    func start() throws {
        let p = NWParameters.tcp
        p.requiredLocalEndpoint = .hostPort(host: "127.0.0.1", port: 47322)
        let l = try NWListener(using: p)
        l.newConnectionHandler = { [weak self] c in self?.accept(c) }
        l.start(queue: q)
        listener = l
        let t = DispatchSource.makeTimerSource(queue: q)
        t.schedule(deadline: .now() + 1, repeating: 1)
        t.setEventHandler { [weak self] in self?.flushOld() }
        t.resume()
        timer = t
    }

    private func accept(_ c: NWConnection) {
        c.start(queue: q)
        c.receive(minimumIncompleteLength: 1, maximumLength: 4096) { [weak self] data, _, _, _ in
            guard let self = self, let data = data, let req = String(data: data, encoding: .utf8) else {
                c.cancel()
                return
            }
            self.lastPoll = Date()
            var known = -1
            if let r = req.range(of: "seq=") {
                let digits = req[r.upperBound...].prefix { $0.isNumber || $0 == "-" }
                known = Int(digits) ?? -1
            }
            if known == self.seq {
                self.waiting.append((c, Date()))
            } else {
                self.reply(c)
            }
        }
    }

    private func reply(_ c: NWConnection) {
        let body = "{\"seq\":\(seq),\"dseq\":\(dseq),\"discord\":\"\(discord)\",\"iseq\":\(iseq),\"insta\":\"\(insta)\"}"
        let resp = "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nAccess-Control-Allow-Origin: *\r\n"
            + "Cache-Control: no-store\r\nContent-Length: \(body.utf8.count)\r\nConnection: close\r\n\r\n" + body
        c.send(content: resp.data(using: .utf8), completion: .contentProcessed { _ in c.cancel() })
    }

    private func bump() {
        seq += 1
        lastPoll = Date()
        let w = waiting
        waiting = []
        for (c, _) in w { reply(c) }
    }

    private func flushOld() {
        let now = Date()
        var keep: [(NWConnection, Date)] = []
        for (c, t) in waiting {
            if now.timeIntervalSince(t) > 20 { reply(c) } else { keep.append((c, t)) }
        }
        waiting = keep
    }

    func sendDiscord(_ a: String) { q.async { self.dseq += 1; self.discord = a; self.bump() } }
    func sendInsta(_ a: String) { q.async { self.iseq += 1; self.insta = a; self.bump() } }
}

// MARK: - Geode mod state

struct ModState: Codable {
    var inLevel: Bool
    var percent: Double
    var practice: Bool
    var attempt: Int
}

// MARK: - System helpers

@discardableResult
func appleScript(_ src: String) -> NSAppleEventDescriptor? {
    var err: NSDictionary?
    return NSAppleScript(source: src)?.executeAndReturnError(&err)
}

func accessibilityOK(prompt: Bool) -> Bool {
    let key = kAXTrustedCheckOptionPrompt.takeUnretainedValue() as String
    return AXIsProcessTrustedWithOptions([key: prompt] as CFDictionary)
}

func pressCombo(_ k: KeyCombo) {
    guard accessibilityOK(prompt: true) else { return }
    let src = CGEventSource(stateID: .hidSystemState)
    for down in [true, false] {
        let e = CGEvent(keyboardEventSource: src, virtualKey: CGKeyCode(k.keyCode), keyDown: down)
        e?.flags = k.cgFlags
        e?.post(tap: .cghidEventTap)
    }
}

// Global hotkey (Carbon). The C callback can't capture, so it uses this global.
var sharedModel: Model?
var hotKeyRef: EventHotKeyRef?
var hotHandlerInstalled = false

func registerHotkey(_ k: KeyCombo?) {
    if let r = hotKeyRef {
        UnregisterEventHotKey(r)
        hotKeyRef = nil
    }
    guard let k = k else { return }
    if !hotHandlerInstalled {
        var spec = EventTypeSpec(eventClass: OSType(kEventClassKeyboard), eventKind: UInt32(kEventHotKeyPressed))
        InstallEventHandler(GetApplicationEventTarget(), { _, _, _ in
            DispatchQueue.main.async { sharedModel?.hotkeyPressed() }
            return noErr
        }, 1, &spec, nil, nil)
        hotHandlerInstalled = true
    }
    let id = EventHotKeyID(signature: OSType(0x4744_4D55), id: 1)
    RegisterEventHotKey(UInt32(k.keyCode), k.carbonMods, id, GetApplicationEventTarget(), 0, &hotKeyRef)
}

// MARK: - Model

enum Rec { case desk, hot }

final class Model: ObservableObject {
    @Published var cfg = Config() {
        didSet {
            if !cfg.instaOn && oldValue.instaOn { holdI = false; release(insta: true) }
            if !cfg.discOn && oldValue.discOn { holdD = false; release(insta: false) }
            save()
        }
    }
    @Published var instaActive = false
    @Published var discActive = false
    @Published var modOK = false
    @Published var modText = "Mod not connected — start Geometry Dash"
    @Published var extConnected = false
    @Published var recording: Rec? = nil
    @Published var serverError = false

    private var holdI = false, holdD = false
    private var lastAttempt = -1
    private var lastOK = Date.distantPast
    private var inFlight = false
    private var started = false
    private var keyMonitor: Any?
    private var tabMuted = false, micMuted = false, webDeaf = false, deskPressed = false
    private var savedInputVolume: Int32 = 50
    private lazy var session: URLSession = {
        let c = URLSessionConfiguration.ephemeral
        c.connectionProxyDictionary = [:]
        c.timeoutIntervalForRequest = 0.3
        c.urlCache = nil
        return URLSession(configuration: c)
    }()

    init() {
        if let d = try? Data(contentsOf: configURL), let c = try? JSONDecoder().decode(Config.self, from: d) {
            cfg = c
        }
    }

    func save() {
        if let d = try? JSONEncoder().encode(cfg) { try? d.write(to: configURL) }
    }

    func start() {
        if started { return }
        started = true
        sharedModel = self
        do { try ExtServer.shared.start() } catch { serverError = true }
        registerHotkey(cfg.hotKey)
        Timer.scheduledTimer(withTimeInterval: 0.01, repeats: true) { [weak self] _ in self?.poll() }
        Timer.scheduledTimer(withTimeInterval: 0.5, repeats: true) { [weak self] _ in
            guard let self = self else { return }
            let c = ExtServer.shared.connected
            if c != self.extConnected { self.extConnected = c }
            if Date().timeIntervalSince(self.lastOK) > 1 { self.setMod(ok: false, text: "Mod not connected — start Geometry Dash") }
        }
    }

    private func setMod(ok: Bool, text: String) {
        if ok != modOK { modOK = ok }
        if text != modText { modText = text }
    }

    private func poll() {
        if inFlight { return }
        inFlight = true
        let url = URL(string: "http://127.0.0.1:47321/state")!
        session.dataTask(with: url) { [weak self] data, _, _ in
            let st = data.flatMap { try? JSONDecoder().decode(ModState.self, from: $0) }
            DispatchQueue.main.async {
                guard let self = self else { return }
                if let st = st {
                    self.lastOK = Date()
                    self.onMod(st)
                    self.inFlight = false
                } else {
                    // back off while the game isn't running
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) { self.inFlight = false }
                }
            }
        }.resume()
    }

    private func onMod(_ st: ModState) {
        if st.inLevel {
            setMod(ok: true, text: String(format: "Connected  ·  %.1f%%", st.percent) + (st.practice ? "  ·  practice" : ""))
        } else {
            setMod(ok: true, text: "Connected  ·  not in a level")
        }
        let newAttempt = lastAttempt != -1 && st.attempt != lastAttempt
        lastAttempt = st.attempt
        if newAttempt { holdI = false; holdD = false }

        for isInsta in [true, false] {
            let on = isInsta ? cfg.instaOn : cfg.discOn
            let active = isInsta ? instaActive : discActive
            let hold = isInsta ? holdI : holdD
            let thr = isInsta ? cfg.instaThr : cfg.discThr
            if !on { release(insta: isInsta); continue }
            if active && cfg.undeafen && (newAttempt || !st.inLevel) { release(insta: isInsta); continue }
            if !active && !hold && st.inLevel && !(st.practice && cfg.ignorePrac) && st.percent >= thr {
                activate(insta: isInsta)
            }
        }
    }

    func activate(insta: Bool) {
        if insta {
            if instaActive { return }
            instaActive = true
            if cfg.instaTab { ExtServer.shared.sendInsta("mute"); tabMuted = true }
            if cfg.mic {
                if let v = appleScript("input volume of (get volume settings)")?.int32Value { savedInputVolume = v }
                appleScript("set volume input volume 0")
                micMuted = true
            }
        } else {
            if discActive { return }
            discActive = true
            if cfg.web { ExtServer.shared.sendDiscord("deafen"); webDeaf = true }
            if cfg.desk, let k = cfg.deskKey { pressCombo(k); deskPressed = true }
        }
    }

    func release(insta: Bool) {
        if insta {
            if !instaActive { return }
            instaActive = false
            if tabMuted { ExtServer.shared.sendInsta("unmute"); tabMuted = false }
            if micMuted { appleScript("set volume input volume \(savedInputVolume)"); micMuted = false }
        } else {
            if !discActive { return }
            discActive = false
            if webDeaf { ExtServer.shared.sendDiscord("undeafen"); webDeaf = false }
            if deskPressed, let k = cfg.deskKey { pressCombo(k); deskPressed = false }
        }
    }

    func releaseAll() {
        release(insta: true)
        release(insta: false)
    }

    func hotkeyPressed() {
        let anyActive = (cfg.instaOn && instaActive) || (cfg.discOn && discActive)
        for isInsta in [true, false] {
            let on = isInsta ? cfg.instaOn : cfg.discOn
            if !on { continue }
            if anyActive {
                release(insta: isInsta)
                if isInsta { holdI = true } else { holdD = true }
            } else {
                activate(insta: isInsta)
            }
        }
    }

    func test() {
        if instaActive || discActive {
            releaseAll()
        } else {
            if cfg.instaOn { activate(insta: true) }
            if cfg.discOn { activate(insta: false) }
        }
    }

    // Record a key combo for the Discord keybind or the toggle hotkey.
    func startRecording(_ r: Rec) {
        if recording != nil { return }
        recording = r
        registerHotkey(nil) // so the current hotkey can be recorded
        keyMonitor = NSEvent.addLocalMonitorForEvents(matching: .keyDown) { [weak self] e in
            guard let self = self else { return e }
            if e.keyCode == 53 { self.finishRecording(nil); return nil } // Esc
            guard let name = keyName(e) else { return nil }
            let f = e.modifierFlags
            let k = KeyCombo(keyCode: e.keyCode, name: name, cmd: f.contains(.command), ctrl: f.contains(.control),
                             opt: f.contains(.option), shift: f.contains(.shift))
            self.finishRecording(k)
            return nil
        }
    }

    private func finishRecording(_ k: KeyCombo?) {
        if let k = k {
            if recording == .desk { cfg.deskKey = k } else { cfg.hotKey = k }
        }
        recording = nil
        if let m = keyMonitor { NSEvent.removeMonitor(m); keyMonitor = nil }
        registerHotkey(cfg.hotKey)
    }
}

// MARK: - Views

struct Card<Content: View>: View {
    let content: Content
    init(@ViewBuilder content: () -> Content) { self.content = content() }
    var body: some View {
        VStack(alignment: .leading, spacing: 10) { content }
            .padding(16)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(RoundedRectangle(cornerRadius: 14, style: .continuous).fill(cCard))
            .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).stroke(cBorder, lineWidth: 1))
    }
}

struct CheckStyle: ToggleStyle {
    func makeBody(configuration: Configuration) -> some View {
        Button { configuration.isOn.toggle() } label: {
            HStack(spacing: 10) {
                ZStack {
                    RoundedRectangle(cornerRadius: 5, style: .continuous)
                        .fill(configuration.isOn ? cAccent : cControl)
                    if configuration.isOn {
                        Image(systemName: "checkmark").font(.system(size: 10, weight: .bold)).foregroundColor(.white)
                    } else {
                        RoundedRectangle(cornerRadius: 5, style: .continuous).stroke(cFaint, lineWidth: 1)
                    }
                }
                .frame(width: 18, height: 18)
                configuration.label.foregroundColor(cText)
                Spacer(minLength: 0)
            }
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }
}

struct PillStyle: ButtonStyle {
    var active = false
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.system(size: 13))
            .foregroundColor(cText)
            .lineLimit(1)
            .frame(width: 130, height: 28)
            .background(RoundedRectangle(cornerRadius: 8, style: .continuous)
                .fill(configuration.isPressed ? cControlH : cControl))
            .overlay(RoundedRectangle(cornerRadius: 8, style: .continuous)
                .stroke(active ? cAccent : cBorder, lineWidth: 1))
    }
}

struct SectionHeader: View {
    let title: String
    @Binding var on: Bool
    let active: Bool
    let threshold: Double
    let activeWord: String

    var body: some View {
        HStack(alignment: .center) {
            VStack(alignment: .leading, spacing: 4) {
                Text(title).font(.system(size: 17, weight: .semibold)).foregroundColor(cText)
                HStack(spacing: 6) {
                    Circle().fill(!on ? cFaint : (active ? cGreen : cAccent)).frame(width: 8, height: 8)
                    Text(!on ? "Off" : (active ? "On  ·  \(activeWord)" : "On  ·  waiting for \(Int(threshold))%"))
                        .font(.system(size: 12))
                        .foregroundColor(active && on ? cGreen : cMuted)
                }
            }
            Spacer()
            Toggle("", isOn: $on).toggleStyle(.switch).tint(cGreen).labelsHidden()
        }
    }
}

struct TriggerRow: View {
    @Binding var value: Double
    var body: some View {
        HStack(spacing: 12) {
            Text("Trigger at").font(.system(size: 13)).foregroundColor(cMuted).frame(width: 72, alignment: .leading)
            Slider(value: $value, in: 1...100, step: 1).tint(cAccent)
            Text("\(Int(value))%").font(.system(size: 15, weight: .semibold)).foregroundColor(cText)
                .frame(width: 48, alignment: .trailing)
        }
    }
}

struct StatusLine: View {
    let ok: Bool
    let text: String
    var body: some View {
        HStack(spacing: 8) {
            Circle().fill(ok ? cGreen : cRed).frame(width: 8, height: 8)
            Text(text).font(.system(size: 12)).foregroundColor(cMuted)
        }
    }
}

struct ContentView: View {
    @ObservedObject var m: Model

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            VStack(alignment: .leading, spacing: 6) {
                Text("GD Call Muter").font(.system(size: 22, weight: .bold)).foregroundColor(cText)
                StatusLine(ok: m.modOK, text: m.modText)
            }
            .padding(.horizontal, 4)

            Card {
                SectionHeader(title: "Instagram", on: $m.cfg.instaOn, active: m.instaActive,
                              threshold: m.cfg.instaThr, activeWord: "muted")
                Group {
                    TriggerRow(value: $m.cfg.instaThr)
                    HStack {
                        Toggle("Mute instagram.com tab", isOn: $m.cfg.instaTab).toggleStyle(CheckStyle())
                        Text("via extension").font(.system(size: 12)).foregroundColor(cMuted)
                    }
                    Toggle("Also mute my microphone", isOn: $m.cfg.mic).toggleStyle(CheckStyle())
                }
                .disabled(!m.cfg.instaOn)
                .opacity(m.cfg.instaOn ? 1 : 0.4)
            }

            Card {
                SectionHeader(title: "Discord", on: $m.cfg.discOn, active: m.discActive,
                              threshold: m.cfg.discThr, activeWord: "deafened")
                Group {
                    TriggerRow(value: $m.cfg.discThr)
                    HStack {
                        Toggle("Deafen discord.com", isOn: $m.cfg.web).toggleStyle(CheckStyle())
                        Text("via extension").font(.system(size: 12)).foregroundColor(cMuted)
                    }
                    HStack {
                        Toggle("Deafen Discord app", isOn: $m.cfg.desk).toggleStyle(CheckStyle())
                        Button(m.recording == .desk ? "Press keys…" : (m.cfg.deskKey?.display ?? "Set key")) {
                            m.startRecording(.desk)
                        }
                        .buttonStyle(PillStyle(active: m.recording == .desk))
                        .disabled(!m.cfg.desk)
                        .opacity(m.cfg.desk ? 1 : 0.4)
                    }
                    Text("Use the same key in Discord → Settings → Keybinds → Toggle Deafen. "
                         + "Allow GD Call Muter in Privacy & Security → Accessibility.")
                        .font(.system(size: 11)).foregroundColor(cMuted)
                        .fixedSize(horizontal: false, vertical: true)
                        .padding(.leading, 28)
                }
                .disabled(!m.cfg.discOn)
                .opacity(m.cfg.discOn ? 1 : 0.4)
            }

            Card {
                Text("Settings").font(.system(size: 17, weight: .semibold)).foregroundColor(cText)
                HStack {
                    Text("Toggle hotkey").font(.system(size: 13)).foregroundColor(cText)
                    Spacer()
                    Button(m.recording == .hot ? "Press keys…" : (m.cfg.hotKey?.display ?? "None")) {
                        m.startRecording(.hot)
                    }
                    .buttonStyle(PillStyle(active: m.recording == .hot))
                }
                Toggle("Unmute / undeafen when I die or restart", isOn: $m.cfg.undeafen).toggleStyle(CheckStyle())
                HStack {
                    Toggle("Ignore practice mode", isOn: $m.cfg.ignorePrac).toggleStyle(CheckStyle())
                    Button((m.instaActive || m.discActive) ? "Stop test" : "Test") { m.test() }
                        .buttonStyle(PillStyle())
                }
            }

            StatusLine(ok: m.extConnected, text: m.extConnected
                ? "Browser extension connected"
                : "Browser extension not connected — open discord.com")
                .padding(.horizontal, 4)
            if m.serverError {
                Text("Another copy of GD Call Muter is already running.")
                    .font(.system(size: 12)).foregroundColor(cRed).padding(.horizontal, 4)
            }
        }
        .font(.system(size: 13))
        .padding(16)
        .frame(width: 440)
        .background(cBg)
    }
}

// MARK: - App

final class AppDelegate: NSObject, NSApplicationDelegate {
    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool { true }
    func applicationWillTerminate(_ notification: Notification) { sharedModel?.releaseAll() }
}

@main
struct GDCallMuterApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var delegate
    @StateObject private var model = Model()

    var body: some Scene {
        WindowGroup("GD Call Muter") {
            ContentView(m: model)
                .preferredColorScheme(.dark)
                .onAppear { model.start() }
        }
        .windowResizability(.contentSize)
    }
}
