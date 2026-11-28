package main

// GD Call Muter v10
// Gets your exact Geometry Dash % from the "GD Call Muter Link" Geode mod.
// Two separate switches, each with its own %:
//   Instagram: mute the instagram.com tab (extension), an app, and/or the mic.
//   Discord:   deafen discord.com (extension) and/or the Discord app (keybind).
//
// The whole window is drawn by hand (GDI+ for smooth rounded shapes, GDI for
// text) so it can be dark, with rounded cards, switches, checkboxes and sliders.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

const appTitle = "GD Call Muter v10"

// shotPath: run with --shot out.png to save a picture of the window and exit
// (used to make the README screenshot).
var shotPath string

func saveShot(bmp uintptr, path string) {
	var img uintptr
	pGdipBmpFromHBITMAP.Call(bmp, 0, uintptr(unsafe.Pointer(&img)))
	if img == 0 {
		return
	}
	// PNG encoder CLSID {557CF406-1A04-11D3-9A73-0000F81EF32E}
	clsid := struct {
		d1 uint32
		d2 uint16
		d3 uint16
		d4 [8]byte
	}{0x557CF406, 0x1A04, 0x11D3, [8]byte{0x9A, 0x73, 0x00, 0x00, 0xF8, 0x1E, 0xF3, 0x2E}}
	p, _ := syscall.UTF16PtrFromString(path)
	pGdipSaveImageToFile.Call(img, uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(&clsid)), 0)
}

var (
	user32  = syscall.NewLazyDLL("user32.dll")
	gdi32   = syscall.NewLazyDLL("gdi32.dll")
	gdiplus = syscall.NewLazyDLL("gdiplus.dll")
	dwmapi  = syscall.NewLazyDLL("dwmapi.dll")

	pRegisterClassExW     = user32.NewProc("RegisterClassExW")
	pCreateWindowExW      = user32.NewProc("CreateWindowExW")
	pDefWindowProcW       = user32.NewProc("DefWindowProcW")
	pGetMessageW          = user32.NewProc("GetMessageW")
	pTranslateMessage     = user32.NewProc("TranslateMessage")
	pDispatchMessageW     = user32.NewProc("DispatchMessageW")
	pPostQuitMessage      = user32.NewProc("PostQuitMessage")
	pSetTimer             = user32.NewProc("SetTimer")
	pShowWindow           = user32.NewProc("ShowWindow")
	pLoadCursorW          = user32.NewProc("LoadCursorW")
	pLoadIconW            = user32.NewProc("LoadIconW")
	pSetCursor            = user32.NewProc("SetCursor")
	pRegisterHotKey       = user32.NewProc("RegisterHotKey")
	pUnregisterHotKey     = user32.NewProc("UnregisterHotKey")
	pGetAsyncKeyState     = user32.NewProc("GetAsyncKeyState")
	pGetKeyState          = user32.NewProc("GetKeyState")
	pSendInput            = user32.NewProc("SendInput")
	pMapVirtualKeyW       = user32.NewProc("MapVirtualKeyW")
	pMessageBoxW          = user32.NewProc("MessageBoxW")
	pAdjustWindowRect     = user32.NewProc("AdjustWindowRect")
	pSetProcessDPIAware   = user32.NewProc("SetProcessDPIAware")
	pGetDpiForSystem      = user32.NewProc("GetDpiForSystem")
	pInvalidateRect       = user32.NewProc("InvalidateRect")
	pBeginPaint           = user32.NewProc("BeginPaint")
	pEndPaint             = user32.NewProc("EndPaint")
	pGetClientRect        = user32.NewProc("GetClientRect")
	pFillRect             = user32.NewProc("FillRect")
	pDrawTextW            = user32.NewProc("DrawTextW")
	pSetCapture           = user32.NewProc("SetCapture")
	pReleaseCapture       = user32.NewProc("ReleaseCapture")
	pTrackMouseEvent      = user32.NewProc("TrackMouseEvent")
	pCreatePopupMenu      = user32.NewProc("CreatePopupMenu")
	pAppendMenuW          = user32.NewProc("AppendMenuW")
	pTrackPopupMenu       = user32.NewProc("TrackPopupMenu")
	pDestroyMenu          = user32.NewProc("DestroyMenu")
	pClientToScreen       = user32.NewProc("ClientToScreen")
	pSetForegroundWindow  = user32.NewProc("SetForegroundWindow")
	pCreateFontW          = gdi32.NewProc("CreateFontW")
	pSetTextColor         = gdi32.NewProc("SetTextColor")
	pSetBkMode            = gdi32.NewProc("SetBkMode")
	pCreateSolidBrush     = gdi32.NewProc("CreateSolidBrush")
	pDeleteObject         = gdi32.NewProc("DeleteObject")
	pSelectObject         = gdi32.NewProc("SelectObject")
	pCreateCompatibleDC   = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBmp  = gdi32.NewProc("CreateCompatibleBitmap")
	pDeleteDC             = gdi32.NewProc("DeleteDC")
	pBitBlt               = gdi32.NewProc("BitBlt")
	pGdiplusStartup       = gdiplus.NewProc("GdiplusStartup")
	pGdipCreateFromHDC    = gdiplus.NewProc("GdipCreateFromHDC")
	pGdipDeleteGraphics   = gdiplus.NewProc("GdipDeleteGraphics")
	pGdipSetSmoothingMode = gdiplus.NewProc("GdipSetSmoothingMode")
	pGdipSetPixelOffset   = gdiplus.NewProc("GdipSetPixelOffsetMode")
	pGdipBmpFromHBITMAP   = gdiplus.NewProc("GdipCreateBitmapFromHBITMAP")
	pGdipSaveImageToFile  = gdiplus.NewProc("GdipSaveImageToFile")
	pGdipCreatePath       = gdiplus.NewProc("GdipCreatePath")
	pGdipDeletePath       = gdiplus.NewProc("GdipDeletePath")
	pGdipAddPathArcI      = gdiplus.NewProc("GdipAddPathArcI")
	pGdipClosePathFigure  = gdiplus.NewProc("GdipClosePathFigure")
	pGdipCreateSolidFill  = gdiplus.NewProc("GdipCreateSolidFill")
	pGdipDeleteBrush      = gdiplus.NewProc("GdipDeleteBrush")
	pGdipFillPath         = gdiplus.NewProc("GdipFillPath")
	pGdipFillEllipseI     = gdiplus.NewProc("GdipFillEllipseI")
	pGdipCreatePen1       = gdiplus.NewProc("GdipCreatePen1")
	pGdipDeletePen        = gdiplus.NewProc("GdipDeletePen")
	pGdipDrawPath         = gdiplus.NewProc("GdipDrawPath")
	pGdipDrawLineI        = gdiplus.NewProc("GdipDrawLineI")
	pGdipSetPenStartCap   = gdiplus.NewProc("GdipSetPenStartCap")
	pGdipSetPenEndCap     = gdiplus.NewProc("GdipSetPenEndCap")
	pDwmSetWindowAttr     = dwmapi.NewProc("DwmSetWindowAttribute")
)

const (
	WS_OVERLAPPED_APP = 0x00CA0000 // caption, sysmenu, minimize
	WM_DESTROY        = 0x2
	WM_PAINT          = 0xF
	WM_CLOSE          = 0x10
	WM_ERASEBKGND     = 0x14
	WM_SETCURSOR      = 0x20
	WM_KEYDOWN        = 0x100
	WM_TIMER          = 0x113
	WM_MOUSEMOVE      = 0x200
	WM_LBUTTONDOWN    = 0x201
	WM_LBUTTONUP      = 0x202
	WM_MOUSELEAVE     = 0x2A3
	WM_HOTKEY         = 0x312
)

// ---------- colours (0xRRGGBB) ----------
type col uint32

const (
	cBg       col = 0x000000
	cCard     col = 0x0A0A0A
	cBorder   col = 0x1F1F1F
	cText     col = 0xEDEDED
	cMuted    col = 0x8A8A8A
	cFaint    col = 0x474747
	cAccent   col = 0xFFFFFF
	cGreen    col = 0xFFFFFF
	cRed      col = 0xE5484D
	cTrack    col = 0x262626
	cControl  col = 0x111111
	cControlH col = 0x1C1C1C
	cWhite    col = 0xFFFFFF
)

func argb(c col, a uint8) uintptr { return uintptr(uint32(a)<<24 | uint32(c)) }
func cref(c col) uintptr {
	return uintptr((uint32(c)&0xFF)<<16 | (uint32(c)>>8&0xFF)<<8 | uint32(c)>>16&0xFF)
}
func f32(f float32) uintptr { return uintptr(math.Float32bits(f)) }

// ---------- settings ----------
type Config struct {
	InstaOn    bool   `json:"instaOn"`
	InstaThr   int    `json:"instaThreshold"`
	InstaTab   bool   `json:"muteInstagramTab"`
	InstaApp   bool   `json:"muteApp"`
	App        string `json:"app"`
	Mic        bool   `json:"muteMic"`
	DiscOn     bool   `json:"discordOn"`
	DiscThr    int    `json:"discordThreshold"`
	Web        bool   `json:"deafenDiscordWeb"`
	Desk       bool   `json:"deafenDiscordApp"`
	DeskKey    string `json:"discordAppKeybind"`
	HotKey     string `json:"toggleHotkey"`
	Undeafen   bool   `json:"undeafenOnDeath"`
	IgnorePrac bool   `json:"ignorePractice"`
}

var cfg = Config{InstaOn: true, InstaThr: 50, InstaTab: true, App: "opera.exe",
	DiscOn: true, DiscThr: 50, Web: true, DeskKey: "Ctrl+Shift+F12", HotKey: "F7",
	Undeafen: true, IgnorePrac: true}

func cfgPath() string {
	d, _ := os.UserConfigDir()
	return filepath.Join(d, "GDCallMuter", "v10.json")
}
func loadCfg() {
	if b, err := os.ReadFile(cfgPath()); err == nil {
		json.Unmarshal(b, &cfg)
	}
}
func saveCfg() {
	os.MkdirAll(filepath.Dir(cfgPath()), 0755)
	b, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(cfgPath(), b, 0644)
}

// ---------- small helpers ----------
func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

var dpi = 96

func sc(v int) int { return v * dpi / 96 }

var hwndMain uintptr

func redraw() { pInvalidateRect.Call(hwndMain, 0, 0) }

// ---------- the two groups: Instagram and Discord ----------
type group struct {
	on     *bool
	thr    *int
	active bool // muted / deafened right now
	hold   bool // turned off by hand mid-run: wait for next attempt
	doOn   func()
	doOff  func()
}

var (
	instaTabMuted, appMutedFlag, micMuted bool
	appMuted                              string
	webDeafened, deskPressed              bool
)

var insta = &group{on: &cfg.InstaOn, thr: &cfg.InstaThr,
	doOn: func() {
		if cfg.InstaTab {
			webInsta("mute")
			instaTabMuted = true
		}
		if cfg.InstaApp {
			appMuted = cfg.App
			muteApp(appMuted, true)
			appMutedFlag = true
		}
		if cfg.Mic {
			muteMic(true)
			micMuted = true
		}
	},
	doOff: func() {
		if instaTabMuted {
			webInsta("unmute")
			instaTabMuted = false
		}
		if appMutedFlag {
			muteApp(appMuted, false)
			appMutedFlag = false
		}
		if micMuted {
			muteMic(false)
			micMuted = false
		}
	}}

var disc = &group{on: &cfg.DiscOn, thr: &cfg.DiscThr,
	doOn: func() {
		if cfg.Web {
			webDiscord("deafen")
			webDeafened = true
		}
		if cfg.Desk {
			pressBind(parseBind(cfg.DeskKey))
			deskPressed = true
		}
	},
	doOff: func() {
		if webDeafened {
			webDiscord("undeafen")
			webDeafened = false
		}
		if deskPressed {
			pressBind(parseBind(cfg.DeskKey))
			deskPressed = false
		}
	}}

var groups = []*group{insta, disc}

func (g *group) activate() {
	if g.active {
		return
	}
	g.active = true
	g.doOn()
	redraw()
}

func (g *group) release() {
	if !g.active {
		return
	}
	g.active = false
	g.doOff()
	redraw()
}

const modOff = "Mod not connected — start Geometry Dash"

var (
	lastAttempt int64 = -1
	modText           = modOff
	modOK       bool
)

func tick() {
	st, ok := getMod()
	txt := modOff
	switch {
	case ok && st.InLevel:
		pr := ""
		if st.Practice {
			pr = "  ·  practice"
		}
		txt = fmt.Sprintf("Connected  ·  %.1f%%%s", st.Percent, pr)
	case ok:
		txt = "Connected  ·  not in a level"
	}
	if txt != modText || ok != modOK {
		modText, modOK = txt, ok
		redraw()
	}

	newAttempt := ok && lastAttempt != -1 && st.Attempt != lastAttempt
	if ok {
		lastAttempt = st.Attempt
	}
	for _, g := range groups {
		if newAttempt {
			g.hold = false
		}
		if !*g.on {
			g.release()
			continue
		}
		if g.active && cfg.Undeafen && (newAttempt || (ok && !st.InLevel)) {
			g.release()
			continue
		}
		if !g.active && !g.hold && ok && st.InLevel &&
			!(st.Practice && cfg.IgnorePrac) && st.Percent >= float64(*g.thr) {
			g.activate()
		}
	}
}

// ---------- hotkey + recording ----------
const (
	recNone = iota
	recDesk
	recHot
)

var recording = recNone

func registerHotkey() {
	pUnregisterHotKey.Call(hwndMain, 1)
	b := parseBind(cfg.HotKey)
	if b.VK != 0 {
		pRegisterHotKey.Call(hwndMain, 1, b.Mods(), uintptr(b.VK))
	}
}

func startRecording(which int) {
	pUnregisterHotKey.Call(hwndMain, 1) // so the current hotkey can be recorded
	recording = which
	redraw()
}

func pollRecording() {
	if recording == recNone {
		return
	}
	if keyDown(vkEscape) {
		finishRecording("")
		return
	}
	for vk := 0x08; vk <= 0x87; vk++ {
		if vk == vkShift || vk == vkControl || vk == vkMenu || vk == vkEscape {
			continue
		}
		if !keyDown(vk) {
			continue
		}
		b := Bind{VK: vk, Ctrl: keyDown(vkControl), Shift: keyDown(vkShift), Alt: keyDown(vkMenu)}
		if s := b.String(); s != "" {
			finishRecording(s)
			return
		}
	}
}

func finishRecording(s string) {
	if s != "" {
		if recording == recDesk {
			cfg.DeskKey = s
		} else {
			cfg.HotKey = s
		}
	}
	recording = recNone
	saveCfg()
	registerHotkey()
	redraw()
}

// ---------- drawing ----------
type shape struct {
	kind          int // 0 rounded rect, 1 circle, 2 check mark
	x, y, w, h, r int
	fill          uintptr // ARGB, 0 = none
	stroke        uintptr // ARGB, 0 = none
	strokeW       float32
}
type text struct {
	s          string
	x, y, w, h int
	c          col
	font       uintptr
	flags      uintptr
}
type hit struct {
	x, y, w, h int
	id         int
}

var (
	shapes                               []shape
	texts                                []text
	hits                                 []hit
	fTitle, fHead, fBody, fSmall, fValue uintptr
	hoverID, focusID, dragID             int
)

const (
	dtLeft   = 0x24 | 0x8000 // vcenter, single line, end ellipsis
	dtRight  = 0x26
	dtCenter = 0x25 | 0x8000
)

func rr(x, y, w, h, r int, fill, stroke uintptr) {
	shapes = append(shapes, shape{kind: 0, x: x, y: y, w: w, h: h, r: r, fill: fill, stroke: stroke, strokeW: float32(sc(1))})
}
func circle(x, y, d int, fill uintptr) {
	shapes = append(shapes, shape{kind: 1, x: x, y: y, w: d, h: d, fill: fill})
}
func label(s string, x, y, w, h int, c col, font, flags uintptr) {
	texts = append(texts, text{s, x, y, w, h, c, font, flags})
}

func drawShapes(g uintptr) {
	for _, s := range shapes {
		switch s.kind {
		case 0:
			var path uintptr
			pGdipCreatePath.Call(0, uintptr(unsafe.Pointer(&path)))
			d := 2 * s.r
			if d > s.w {
				d = s.w
			}
			if d > s.h {
				d = s.h
			}
			if d < 1 {
				d = 1
			}
			pGdipAddPathArcI.Call(path, uintptr(s.x), uintptr(s.y), uintptr(d), uintptr(d), f32(180), f32(90))
			pGdipAddPathArcI.Call(path, uintptr(s.x+s.w-d), uintptr(s.y), uintptr(d), uintptr(d), f32(270), f32(90))
			pGdipAddPathArcI.Call(path, uintptr(s.x+s.w-d), uintptr(s.y+s.h-d), uintptr(d), uintptr(d), f32(0), f32(90))
			pGdipAddPathArcI.Call(path, uintptr(s.x), uintptr(s.y+s.h-d), uintptr(d), uintptr(d), f32(90), f32(90))
			pGdipClosePathFigure.Call(path)
			if s.fill != 0 {
				var br uintptr
				pGdipCreateSolidFill.Call(s.fill, uintptr(unsafe.Pointer(&br)))
				pGdipFillPath.Call(g, br, path)
				pGdipDeleteBrush.Call(br)
			}
			if s.stroke != 0 {
				var pen uintptr
				pGdipCreatePen1.Call(s.stroke, f32(s.strokeW), 2, uintptr(unsafe.Pointer(&pen)))
				pGdipDrawPath.Call(g, pen, path)
				pGdipDeletePen.Call(pen)
			}
			pGdipDeletePath.Call(path)
		case 1:
			var br uintptr
			pGdipCreateSolidFill.Call(s.fill, uintptr(unsafe.Pointer(&br)))
			pGdipFillEllipseI.Call(g, br, uintptr(s.x), uintptr(s.y), uintptr(s.w), uintptr(s.h))
			pGdipDeleteBrush.Call(br)
		case 2: // check mark inside a box at x,y of size w
			var pen uintptr
			pGdipCreatePen1.Call(s.stroke, f32(s.strokeW), 2, uintptr(unsafe.Pointer(&pen)))
			pGdipSetPenStartCap.Call(pen, 2) // round
			pGdipSetPenEndCap.Call(pen, 2)
			x, y, w := s.x, s.y, s.w
			pGdipDrawLineI.Call(g, pen, uintptr(x+w*26/100), uintptr(y+w*52/100), uintptr(x+w*43/100), uintptr(y+w*68/100))
			pGdipDrawLineI.Call(g, pen, uintptr(x+w*43/100), uintptr(y+w*68/100), uintptr(x+w*74/100), uintptr(y+w*34/100))
			pGdipDeletePen.Call(pen)
		}
	}
}

// ---------- widgets (each draws itself and registers a click area) ----------
const (
	idNone = iota
	idISwitch
	idISlider
	idITab
	idIApp
	idIAppPick
	idIMic
	idDSwitch
	idDSlider
	idDWeb
	idDDesk
	idDeskKey
	idHotKey
	idUndeaf
	idPrac
	idTest
)

var enabledIDs = map[int]bool{}

func addHit(id, x, y, w, h int, enabled bool) {
	if enabled {
		hits = append(hits, hit{x, y, w, h, id})
		enabledIDs[id] = true
	}
}

func focusRing(id, x, y, w, h, r int) {
	if focusID == id {
		p := sc(3)
		rr(x-p, y-p, w+2*p, h+2*p, r+p, 0, argb(cAccent, 255))
	}
}

func toggleSwitch(id, x, y int, on bool) {
	w, h := sc(46), sc(26)
	fill := argb(cTrack, 255)
	if on {
		fill = argb(cGreen, 255)
	} else if hoverID == id {
		fill = argb(cControlH, 255)
	}
	rr(x, y, w, h, h/2, fill, 0)
	k, pad := sc(20), sc(3)
	kx := x + pad
	if on {
		kx = x + w - pad - k
	}
	kc := cWhite
	if on {
		kc = cBg
	}
	circle(kx, y+pad, k, argb(kc, 255))
	focusRing(id, x, y, w, h, h/2)
	addHit(id, x, y, w, h, true)
}

func checkbox(id, x, y, w int, labelText, hint string, checked, enabled bool) {
	b := sc(18)
	rowH := sc(28)
	by := y + (rowH-b)/2
	textC := cText
	if !enabled {
		textC = cFaint
	}
	if checked {
		a := uint8(255)
		if !enabled {
			a = 80
		}
		rr(x, by, b, b, sc(4), argb(cAccent, a), 0)
		shapes = append(shapes, shape{kind: 2, x: x, y: by, w: b, stroke: argb(cBg, a), strokeW: float32(sc(2))})
	} else {
		border := cFaint
		if enabled && hoverID == id {
			border = cMuted
		}
		rr(x, by, b, b, sc(4), argb(cControl, 255), argb(border, 255))
	}
	label(labelText, x+b+sc(10), y, w-b-sc(10), rowH, textC, fBody, dtLeft)
	if hint != "" {
		hc := cMuted
		if !enabled {
			hc = cFaint
		}
		label(hint, x, y, w, rowH, hc, fSmall, dtRight)
	}
	focusRing(id, x, by, b, b, sc(4))
	addHit(id, x, y, w, rowH, enabled)
}

var sliderX = map[int][2]int{}

func slider(id, x, y, w, val int, enabled bool) {
	rowH := sc(28)
	th := sc(6)
	ty := y + rowH/2 - th/2
	rr(x, ty, w, th, th/2, argb(cTrack, 255), 0)
	pos := x + w*(val-1)/99
	a := uint8(255)
	if !enabled {
		a = 70
	}
	if pos > x {
		rr(x, ty, pos-x+th/2, th, th/2, argb(cAccent, a), 0)
	}
	k := sc(18)
	if enabled && (hoverID == id || dragID == id) {
		k = sc(20)
	}
	circle(pos-k/2, y+rowH/2-k/2, k, argb(cWhite, a))
	focusRing(id, pos-k/2, y+rowH/2-k/2, k, k, k/2)
	sliderX[id] = [2]int{x, w}
	addHit(id, x-sc(10), y, w+sc(20), rowH, enabled)
}

func pill(id, x, y, w, h int, s string, enabled, active bool) {
	fill := argb(cControl, 255)
	if enabled && hoverID == id {
		fill = argb(cControlH, 255)
	}
	stroke := argb(cBorder, 255)
	if active {
		stroke = argb(cAccent, 255)
	}
	rr(x, y, w, h, sc(6), fill, stroke)
	c := cText
	if !enabled {
		c = cFaint
	}
	label(s, x+sc(10), y, w-sc(20), h, c, fBody, dtCenter)
	focusRing(id, x, y, w, h, sc(6))
	addHit(id, x, y, w, h, enabled)
}

// sectionCard draws a card with title, status, ON switch and % slider,
// and returns where the option rows start.
func sectionCard(x, y, w, h int, title string, g *group, switchID, sliderID int) int {
	rr(x, y, w, h, sc(8), argb(cCard, 255), argb(cBorder, 255))
	px := x + sc(18)
	label(title, px, y+sc(14), sc(200), sc(24), cText, fHead, dtLeft)

	st, stc, dc := "Off", cMuted, cFaint
	if *g.on {
		st, dc = fmt.Sprintf("On  ·  waiting for %d%%", *g.thr), cAccent
		if g.active {
			st, stc, dc = "On  ·  muted", cGreen, cGreen
			if g == disc {
				st = "On  ·  deafened"
			}
		}
	}
	dot := sc(8)
	circle(px, y+sc(40)+(sc(18)-dot)/2, dot, argb(dc, 255))
	label(st, px+dot+sc(8), y+sc(40), sc(260), sc(18), stc, fSmall, dtLeft)
	toggleSwitch(switchID, x+w-sc(18)-sc(46), y+sc(22), *g.on)

	ry := y + sc(70)
	on := *g.on
	lc, vc := cMuted, cText
	if !on {
		lc, vc = cFaint, cFaint
	}
	label("Trigger at", px, ry, sc(84), sc(28), lc, fBody, dtLeft)
	sx := px + sc(92)
	sw := w - sc(36) - sc(92) - sc(60)
	slider(sliderID, sx, ry, sw, *g.thr, on)
	label(fmt.Sprintf("%d%%", *g.thr), x+w-sc(18)-sc(48), ry, sc(48), sc(28), vc, fValue, dtRight)
	return ry + sc(40)
}

func layout(cw int) {
	x, w := sc(16), cw-sc(32)
	px := x + sc(18)
	iw := w - sc(36)
	pickW := sc(150)
	row := sc(32)

	// header
	label("GD Call Muter", sc(20), sc(12), sc(260), sc(32), cText, fTitle, dtLeft)
	dc := cRed
	if modOK {
		dc = cGreen
	}
	circle(sc(22), sc(46)+sc(5), sc(8), argb(dc, 255))
	label(modText, sc(38), sc(46), cw-sc(60), sc(18), cMuted, fSmall, dtLeft)

	// Instagram
	y := sc(78)
	ch := sc(70) + sc(40) + 3*row
	ry := sectionCard(x, y, w, ch, "Instagram", insta, idISwitch, idISlider)
	on := cfg.InstaOn
	checkbox(idITab, px, ry, iw, "Mute instagram.com tab", "via extension", cfg.InstaTab, on)
	ry += row
	checkbox(idIApp, px, ry, iw-pickW-sc(10), "Mute an app", "", cfg.InstaApp, on)
	app := cfg.App
	if app == "" {
		app = "Pick app"
	}
	pill(idIAppPick, px+iw-pickW, ry, pickW, sc(28), app+"  ▾", on && cfg.InstaApp, false)
	ry += row
	checkbox(idIMic, px, ry, iw, "Also mute my microphone", "", cfg.Mic, on)

	// Discord
	y += ch + sc(12)
	ry = sectionCard(x, y, w, ch, "Discord", disc, idDSwitch, idDSlider)
	on = cfg.DiscOn
	checkbox(idDWeb, px, ry, iw, "Deafen discord.com", "via extension", cfg.Web, on)
	ry += row
	checkbox(idDDesk, px, ry, iw-pickW-sc(10), "Deafen Discord app", "", cfg.Desk, on)
	k := cfg.DeskKey
	if recording == recDesk {
		k = "Press keys…"
	}
	pill(idDeskKey, px+iw-pickW, ry, pickW, sc(28), k, on && cfg.Desk, recording == recDesk)
	ry += row
	hc := cMuted
	if !(on && cfg.Desk) {
		hc = cFaint
	}
	label("Use the same key in Discord → Keybinds → Toggle Deafen", px+sc(28), ry, iw-sc(28), sc(28), hc, fSmall, dtLeft)

	// Settings
	y += ch + sc(12)
	sh := sc(50) + 3*row + sc(6)
	rr(x, y, w, sh, sc(8), argb(cCard, 255), argb(cBorder, 255))
	label("Settings", px, y+sc(14), sc(200), sc(24), cText, fHead, dtLeft)
	ry = y + sc(48)
	label("Toggle hotkey", px, ry, sc(160), sc(28), cText, fBody, dtLeft)
	k = cfg.HotKey
	if k == "" {
		k = "None"
	}
	if recording == recHot {
		k = "Press keys…"
	}
	pill(idHotKey, px+iw-pickW, ry, pickW, sc(28), k, true, recording == recHot)
	ry += row
	checkbox(idUndeaf, px, ry, iw, "Unmute / undeafen when I die or restart", "", cfg.Undeafen, true)
	ry += row
	checkbox(idPrac, px, ry, iw-sc(100), "Ignore practice mode", "", cfg.IgnorePrac, true)
	tl := "Test"
	if insta.active || disc.active {
		tl = "Stop test"
	}
	pill(idTest, px+iw-sc(90), ry, sc(90), sc(28), tl, true, false)

	// footer
	y += sh + sc(14)
	ec, et := cRed, "Browser extension not connected — open discord.com"
	if webConnected() {
		ec, et = cGreen, "Browser extension connected"
	}
	circle(sc(22), y+sc(5), sc(8), argb(ec, 255))
	label(et, sc(38), y, cw-sc(60), sc(18), cMuted, fSmall, dtLeft)
}

const clientW, clientH = 448, 714

type rect struct{ Left, Top, Right, Bottom int32 }

func paint() {
	var ps [80]byte
	hdc, _, _ := pBeginPaint.Call(hwndMain, uintptr(unsafe.Pointer(&ps)))
	var rc rect
	pGetClientRect.Call(hwndMain, uintptr(unsafe.Pointer(&rc)))
	w, h := int(rc.Right), int(rc.Bottom)
	mem, _, _ := pCreateCompatibleDC.Call(hdc)
	bmp, _, _ := pCreateCompatibleBmp.Call(hdc, uintptr(w), uintptr(h))
	oldBmp, _, _ := pSelectObject.Call(mem, bmp)
	br, _, _ := pCreateSolidBrush.Call(cref(cBg))
	pFillRect.Call(mem, uintptr(unsafe.Pointer(&rc)), br)
	pDeleteObject.Call(br)

	shapes, texts, hits = shapes[:0], texts[:0], hits[:0]
	enabledIDs = map[int]bool{}
	layout(w)

	var g uintptr
	pGdipCreateFromHDC.Call(mem, uintptr(unsafe.Pointer(&g)))
	pGdipSetSmoothingMode.Call(g, 4) // anti-alias
	pGdipSetPixelOffset.Call(g, 2)   // high quality
	drawShapes(g)
	pGdipDeleteGraphics.Call(g)

	pSetBkMode.Call(mem, 1)
	for _, t := range texts {
		pSelectObject.Call(mem, t.font)
		pSetTextColor.Call(mem, cref(t.c))
		r := rect{int32(t.x), int32(t.y), int32(t.x + t.w), int32(t.y + t.h)}
		s, _ := syscall.UTF16FromString(t.s)
		pDrawTextW.Call(mem, uintptr(unsafe.Pointer(&s[0])), ^uintptr(0), uintptr(unsafe.Pointer(&r)), t.flags)
	}
	pBitBlt.Call(hdc, 0, 0, uintptr(w), uintptr(h), mem, 0, 0, 0xCC0020)
	pSelectObject.Call(mem, oldBmp)
	if shotPath != "" {
		saveShot(bmp, shotPath)
		shotPath = ""
		pPostQuitMessage.Call(0)
	}
	pDeleteObject.Call(bmp)
	pDeleteDC.Call(mem)
	pEndPaint.Call(hwndMain, uintptr(unsafe.Pointer(&ps)))
}

// ---------- input ----------
func hitAt(x, y int) int {
	for i := len(hits) - 1; i >= 0; i-- {
		h := hits[i]
		if x >= h.x && x < h.x+h.w && y >= h.y && y < h.y+h.h {
			return h.id
		}
	}
	return idNone
}

func setSlider(id, mx int) {
	sx := sliderX[id]
	if sx[1] <= 0 {
		return
	}
	v := clamp(1+(mx-sx[0])*99/sx[1], 1, 100)
	if id == idISlider {
		cfg.InstaThr = v
	} else {
		cfg.DiscThr = v
	}
	redraw()
}

func toggleGroup(g *group) {
	*g.on = !*g.on
	g.hold = false
	if !*g.on {
		g.release()
	}
}

func pickApp() {
	m, _, _ := pCreatePopupMenu.Call()
	apps := audioApps()
	found := false
	for _, a := range apps {
		if a == cfg.App {
			found = true
		}
	}
	if !found && cfg.App != "" {
		apps = append([]string{cfg.App}, apps...)
	}
	for i, a := range apps {
		flags := uintptr(0)
		if a == cfg.App {
			flags = 8 // checked
		}
		pAppendMenuW.Call(m, flags, uintptr(1000+i), uintptr(unsafe.Pointer(u16(a))))
	}
	if len(apps) == 0 {
		pAppendMenuW.Call(m, 1, 0, uintptr(unsafe.Pointer(u16("No apps are playing sound"))))
	}
	for _, h := range hits {
		if h.id == idIAppPick {
			pt := [2]int32{int32(h.x), int32(h.y + h.h)}
			pClientToScreen.Call(hwndMain, uintptr(unsafe.Pointer(&pt)))
			pSetForegroundWindow.Call(hwndMain)
			r, _, _ := pTrackPopupMenu.Call(m, 0x100, uintptr(pt[0]), uintptr(pt[1]), 0, hwndMain, 0)
			if r >= 1000 && int(r-1000) < len(apps) {
				cfg.App = apps[r-1000]
			}
			break
		}
	}
	pDestroyMenu.Call(m)
}

func activateID(id int) {
	switch id {
	case idISwitch:
		toggleGroup(insta)
	case idDSwitch:
		toggleGroup(disc)
	case idITab:
		cfg.InstaTab = !cfg.InstaTab
	case idIApp:
		cfg.InstaApp = !cfg.InstaApp
	case idIAppPick:
		pickApp()
	case idIMic:
		cfg.Mic = !cfg.Mic
	case idDWeb:
		cfg.Web = !cfg.Web
	case idDDesk:
		cfg.Desk = !cfg.Desk
	case idDeskKey:
		startRecording(recDesk)
	case idHotKey:
		startRecording(recHot)
	case idUndeaf:
		cfg.Undeafen = !cfg.Undeafen
	case idPrac:
		cfg.IgnorePrac = !cfg.IgnorePrac
	case idTest:
		anyActive := insta.active || disc.active
		for _, g := range groups {
			if anyActive {
				g.release()
			} else if *g.on {
				g.activate()
			}
		}
	}
	saveCfg()
	redraw()
}

func moveFocus(dir int) {
	if len(hits) == 0 {
		return
	}
	idx := -1
	for i, h := range hits {
		if h.id == focusID {
			idx = i
		}
	}
	if idx == -1 && dir < 0 {
		idx = 0
	}
	idx += dir
	if idx < 0 {
		idx = len(hits) - 1
	}
	if idx >= len(hits) {
		idx = 0
	}
	focusID = hits[idx].id
	redraw()
}

func trackLeave() {
	tme := struct {
		size, flags uint32
		hwnd        uintptr
		hover       uint32
	}{flags: 2, hwnd: hwndMain}
	tme.size = uint32(unsafe.Sizeof(tme))
	pTrackMouseEvent.Call(uintptr(unsafe.Pointer(&tme)))
}

var handCursor, arrowCursor uintptr

func wndProc(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
	mx, my := int(int16(lp&0xFFFF)), int(int16(lp>>16&0xFFFF))
	switch msg {
	case WM_PAINT:
		paint()
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_TIMER:
		if wp == 1 {
			pollRecording()
			tick()
		} else {
			redraw() // extension status
		}
		return 0
	case WM_MOUSEMOVE:
		if dragID != idNone {
			setSlider(dragID, mx)
			return 0
		}
		if h := hitAt(mx, my); h != hoverID {
			hoverID = h
			trackLeave()
			redraw()
		}
		return 0
	case WM_MOUSELEAVE:
		if hoverID != idNone {
			hoverID = idNone
			redraw()
		}
		return 0
	case WM_SETCURSOR:
		if lp&0xFFFF == 1 { // in the client area
			if hoverID != idNone {
				pSetCursor.Call(handCursor)
			} else {
				pSetCursor.Call(arrowCursor)
			}
			return 1
		}
	case WM_LBUTTONDOWN:
		id := hitAt(mx, my)
		focusID = idNone
		if id == idISlider || id == idDSlider {
			dragID = id
			pSetCapture.Call(hwnd)
			setSlider(id, mx)
			return 0
		}
		if id != idNone {
			activateID(id)
		}
		redraw()
		return 0
	case WM_LBUTTONUP:
		if dragID != idNone {
			dragID = idNone
			pReleaseCapture.Call()
			saveCfg()
			redraw()
		}
		return 0
	case WM_KEYDOWN:
		if recording != recNone {
			return 0
		}
		shift, _, _ := pGetKeyState.Call(vkShift)
		step := 1
		if shift&0x8000 != 0 {
			step = 10
		}
		switch wp {
		case 0x09: // Tab
			if shift&0x8000 != 0 {
				moveFocus(-1)
			} else {
				moveFocus(1)
			}
		case 0x20, 0x0D: // Space, Enter
			if focusID != idNone && enabledIDs[focusID] {
				activateID(focusID)
			}
		case 0x25, 0x27: // Left, Right
			if wp == 0x25 {
				step = -step
			}
			if focusID == idISlider && cfg.InstaOn {
				cfg.InstaThr = clamp(cfg.InstaThr+step, 1, 100)
			} else if focusID == idDSlider && cfg.DiscOn {
				cfg.DiscThr = clamp(cfg.DiscThr+step, 1, 100)
			}
			saveCfg()
			redraw()
		}
		return 0
	case WM_HOTKEY:
		// toggle every group that is switched ON
		anyActive := false
		for _, g := range groups {
			if *g.on && g.active {
				anyActive = true
			}
		}
		for _, g := range groups {
			if !*g.on {
				continue
			}
			if anyActive {
				g.release()
				g.hold = true
			} else {
				g.activate()
			}
		}
		return 0
	case WM_CLOSE:
		for _, g := range groups {
			g.release()
		}
		saveCfg()
	case WM_DESTROY:
		pPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, uintptr(msg), wp, lp)
	return r
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

type wndClassEx struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	IconSm                             uintptr
}
type msgT struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      [2]int32
	_       uint32
}

func mkFont(size, weight int) uintptr {
	f, _, _ := pCreateFontW.Call(uintptr(-sc(size)&0xFFFFFFFF), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0,
		uintptr(unsafe.Pointer(u16("Segoe UI"))))
	return f
}

func main() {
	runtime.LockOSThread()
	pSetProcessDPIAware.Call()
	if r, _, _ := pGetDpiForSystem.Call(); r > 0 {
		dpi = int(r)
	}
	for i, a := range os.Args {
		if a == "--shot" && i+1 < len(os.Args) {
			shotPath = os.Args[i+1]
		}
	}
	// a copy named "...shot...exe" saves its picture next to itself when opened
	if exe, err := os.Executable(); err == nil && shotPath == "" &&
		strings.Contains(strings.ToLower(filepath.Base(exe)), "shot") {
		shotPath = filepath.Join(filepath.Dir(exe), "gd-app-black.png")
	}
	comInit()
	gsi := struct {
		version            uint32
		callback           uintptr
		noThread, noCodecs int32
	}{version: 1}
	var gpToken uintptr
	pGdiplusStartup.Call(uintptr(unsafe.Pointer(&gpToken)), uintptr(unsafe.Pointer(&gsi)), 0)
	loadCfg()

	if err := serveWeb(); err != nil && shotPath == "" {
		pMessageBoxW.Call(0, uintptr(unsafe.Pointer(u16("GD Call Muter is already running (or another app is using port 47322). Close the other copy, including v9, and try again."))),
			uintptr(unsafe.Pointer(u16(appTitle))), 0x30)
		return
	}
	go pollMod()

	fTitle, fHead, fBody, fSmall, fValue = mkFont(22, 700), mkFont(17, 600), mkFont(14, 400), mkFont(13, 400), mkFont(16, 600)
	arrowCursor, _, _ = pLoadCursorW.Call(0, 32512)
	handCursor, _, _ = pLoadCursorW.Call(0, 32649)
	icon, _, _ := pLoadIconW.Call(0, 32512)
	wc := wndClassEx{WndProc: syscall.NewCallback(wndProc), Cursor: arrowCursor, Icon: icon, IconSm: icon,
		ClassName: u16("GDCallMuter10")}
	wc.Size = uint32(unsafe.Sizeof(wc))
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	rc := rect{0, 0, int32(sc(clientW)), int32(sc(clientH))}
	pAdjustWindowRect.Call(uintptr(unsafe.Pointer(&rc)), WS_OVERLAPPED_APP, 0)
	hwndMain, _, _ = pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(wc.ClassName)), uintptr(unsafe.Pointer(u16(appTitle))),
		WS_OVERLAPPED_APP, 0x80000000, 0x80000000, uintptr(rc.Right-rc.Left), uintptr(rc.Bottom-rc.Top), 0, 0, 0, 0)

	// dark title bar, matching caption colour, rounded window corners
	dark := int32(1)
	pDwmSetWindowAttr.Call(hwndMain, 20, uintptr(unsafe.Pointer(&dark)), 4)
	capCol := uint32(cref(cBg))
	pDwmSetWindowAttr.Call(hwndMain, 35, uintptr(unsafe.Pointer(&capCol)), 4)
	corner := int32(2)
	pDwmSetWindowAttr.Call(hwndMain, 33, uintptr(unsafe.Pointer(&corner)), 4)

	registerHotkey()
	pShowWindow.Call(hwndMain, 1)
	pSetTimer.Call(hwndMain, 1, 10, 0)  // game check
	pSetTimer.Call(hwndMain, 2, 500, 0) // extension status

	var m msgT
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}
