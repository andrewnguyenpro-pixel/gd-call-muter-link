package main

// Key combos: text like "Ctrl+Shift+D" <-> keys, plus pressing them.

import (
	"strconv"
	"strings"
	"unsafe"
)

const (
	vkShift     = 0x10
	vkControl   = 0x11
	vkMenu      = 0x12 // Alt
	vkEscape    = 0x1B
	vkF1        = 0x70
	vkF24       = 0x87
	modAlt      = 0x1
	modCtrl     = 0x2
	modShift    = 0x4
	modNoRepeat = 0x4000
)

type Bind struct {
	VK               int
	Ctrl, Shift, Alt bool
}

var namedKeys = map[string]int{
	"SPACE": 0x20, "TAB": 0x09, "INSERT": 0x2D, "DELETE": 0x2E, "HOME": 0x24, "END": 0x23,
	"PAGEUP": 0x21, "PAGEDOWN": 0x22, "PAUSE": 0x13, "UP": 0x26, "DOWN": 0x28, "LEFT": 0x25, "RIGHT": 0x27,
	"NUM0": 0x60, "NUM1": 0x61, "NUM2": 0x62, "NUM3": 0x63, "NUM4": 0x64, "NUM5": 0x65,
	"NUM6": 0x66, "NUM7": 0x67, "NUM8": 0x68, "NUM9": 0x69,
}

func parseBind(s string) Bind {
	var b Bind
	for _, part := range strings.Split(strings.ToUpper(s), "+") {
		part = strings.TrimSpace(part)
		switch {
		case part == "":
		case part == "CTRL" || part == "CONTROL":
			b.Ctrl = true
		case part == "SHIFT":
			b.Shift = true
		case part == "ALT":
			b.Alt = true
		case len(part) == 1 && ((part[0] >= 'A' && part[0] <= 'Z') || (part[0] >= '0' && part[0] <= '9')):
			b.VK = int(part[0])
		case len(part) >= 2 && part[0] == 'F':
			if n, err := strconv.Atoi(part[1:]); err == nil && n >= 1 && n <= 24 {
				b.VK = vkF1 + n - 1
			}
		default:
			if vk, ok := namedKeys[part]; ok {
				b.VK = vk
			}
		}
	}
	return b
}

func keyName(vk int) string {
	switch {
	case (vk >= 'A' && vk <= 'Z') || (vk >= '0' && vk <= '9'):
		return string(rune(vk))
	case vk >= vkF1 && vk <= vkF24:
		return "F" + strconv.Itoa(vk-vkF1+1)
	}
	for n, v := range namedKeys {
		if v == vk {
			return strings.ToUpper(n[:1]) + strings.ToLower(n[1:])
		}
	}
	return ""
}

func (b Bind) String() string {
	k := keyName(b.VK)
	if k == "" {
		return ""
	}
	s := ""
	if b.Ctrl {
		s += "Ctrl+"
	}
	if b.Shift {
		s += "Shift+"
	}
	if b.Alt {
		s += "Alt+"
	}
	return s + k
}

func (b Bind) Mods() uintptr {
	m := uintptr(modNoRepeat)
	if b.Ctrl {
		m |= modCtrl
	}
	if b.Shift {
		m |= modShift
	}
	if b.Alt {
		m |= modAlt
	}
	return m
}

func keyDown(vk int) bool {
	r, _, _ := pGetAsyncKeyState.Call(uintptr(vk))
	return r&0x8000 != 0
}

// INPUT for SendInput (64-bit layout: 4-byte type, 4 pad, 32-byte union)
type kbInput struct {
	typ   uint32
	_     uint32
	vk    uint16
	scan  uint16
	flags uint32
	time  uint32
	_     uint32
	extra uintptr
	_     [8]byte
}

// pressBind presses and releases a key combo, e.g. Discord's deafen keybind
func pressBind(b Bind) {
	if b.VK == 0 {
		return
	}
	var seq []kbInput
	add := func(vk int, up bool) {
		in := kbInput{typ: 1, vk: uint16(vk)}
		sc, _, _ := pMapVirtualKeyW.Call(uintptr(vk), 0)
		in.scan = uint16(sc)
		if up {
			in.flags = 2 // KEYEVENTF_KEYUP
		}
		seq = append(seq, in)
	}
	var mods []int
	if b.Ctrl {
		mods = append(mods, vkControl)
	}
	if b.Shift {
		mods = append(mods, vkShift)
	}
	if b.Alt {
		mods = append(mods, vkMenu)
	}
	for _, m := range mods {
		add(m, false)
	}
	add(b.VK, false)
	add(b.VK, true)
	for i := len(mods) - 1; i >= 0; i-- {
		add(mods[i], true)
	}
	pSendInput.Call(uintptr(len(seq)), uintptr(unsafe.Pointer(&seq[0])), unsafe.Sizeof(seq[0]))
}
