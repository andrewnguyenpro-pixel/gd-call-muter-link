package main

// Windows audio: mute one app's sound and the microphone, using the Core
// Audio COM interfaces directly (no outside libraries).

import (
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"unsafe"
)

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var (
	clsidMMDeviceEnumerator  = GUID{0xBCDE0395, 0xE52F, 0x467C, [8]byte{0x8E, 0x3D, 0xC4, 0x57, 0x92, 0x91, 0x69, 0x2E}}
	iidIMMDeviceEnumerator   = GUID{0xA95664D2, 0x9614, 0x4F35, [8]byte{0xA7, 0x46, 0xDE, 0x8D, 0xB6, 0x36, 0x17, 0xE6}}
	iidIAudioSessionManager2 = GUID{0x77AA99A0, 0x1BD6, 0x484F, [8]byte{0x8B, 0xC7, 0x2C, 0x65, 0x4C, 0x9A, 0x9B, 0x6F}}
	iidIAudioSessionControl2 = GUID{0xBFB7FF88, 0x7239, 0x4FC9, [8]byte{0x8F, 0xA2, 0x07, 0xC9, 0x50, 0xBE, 0x9C, 0x6D}}
	iidISimpleAudioVolume    = GUID{0x87CE5498, 0x68D6, 0x44E5, [8]byte{0x92, 0x15, 0x6D, 0xA4, 0x7E, 0xF8, 0x83, 0xD8}}
	iidIAudioEndpointVolume  = GUID{0x5CDF2C82, 0x841E, 0x4546, [8]byte{0x97, 0x22, 0x0C, 0xF7, 0x40, 0x78, 0x22, 0x9A}}
)

const (
	eRender           = 0
	eCapture          = 1
	eCommunications   = 2
	deviceStateActive = 1
	clsctxAll         = 0x17
)

var (
	ole32                       = syscall.NewLazyDLL("ole32.dll")
	pCoInitializeEx             = ole32.NewProc("CoInitializeEx")
	pCoCreateInstance           = ole32.NewProc("CoCreateInstance")
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	pOpenProcess                = kernel32.NewProc("OpenProcess")
	pQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
	pCloseHandle                = kernel32.NewProc("CloseHandle")
)

// call method number idx of a COM object
func comCall(obj uintptr, idx int, args ...uintptr) uintptr {
	vtbl := *(*uintptr)(unsafe.Pointer(obj))
	fn := *(*uintptr)(unsafe.Pointer(vtbl + uintptr(idx)*unsafe.Sizeof(uintptr(0))))
	all := append([]uintptr{obj}, args...)
	r, _, _ := syscall.SyscallN(fn, all...)
	return r
}

func comRelease(obj uintptr) {
	if obj != 0 {
		comCall(obj, 2)
	}
}

func comInit() {
	pCoInitializeEx.Call(0, 2) // apartment threaded, on the UI thread
}

func deviceEnumerator() uintptr {
	var e uintptr
	r, _, _ := pCoCreateInstance.Call(uintptr(unsafe.Pointer(&clsidMMDeviceEnumerator)), 0, clsctxAll,
		uintptr(unsafe.Pointer(&iidIMMDeviceEnumerator)), uintptr(unsafe.Pointer(&e)))
	if r != 0 {
		return 0
	}
	return e
}

// processName returns e.g. "opera.exe" for a process ID
func processName(pid uint32) string {
	if pid == 0 {
		return ""
	}
	h, _, _ := pOpenProcess.Call(0x1000, 0, uintptr(pid)) // PROCESS_QUERY_LIMITED_INFORMATION
	if h == 0 {
		return ""
	}
	defer pCloseHandle.Call(h)
	buf := make([]uint16, 1024)
	n := uint32(len(buf))
	r, _, _ := pQueryFullProcessImageNameW.Call(h, 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&n)))
	if r == 0 {
		return ""
	}
	return strings.ToLower(filepath.Base(syscall.UTF16ToString(buf[:n])))
}

// forEachSession calls fn for every audio session on every active output device
func forEachSession(fn func(name string, ctrl2 uintptr)) {
	en := deviceEnumerator()
	if en == 0 {
		return
	}
	defer comRelease(en)
	var coll uintptr
	if comCall(en, 3, eRender, deviceStateActive, uintptr(unsafe.Pointer(&coll))) != 0 {
		return
	}
	defer comRelease(coll)
	var nDev uint32
	comCall(coll, 3, uintptr(unsafe.Pointer(&nDev)))
	for d := uint32(0); d < nDev; d++ {
		var dev uintptr
		if comCall(coll, 4, uintptr(d), uintptr(unsafe.Pointer(&dev))) != 0 {
			continue
		}
		var mgr uintptr
		if comCall(dev, 3, uintptr(unsafe.Pointer(&iidIAudioSessionManager2)), clsctxAll, 0, uintptr(unsafe.Pointer(&mgr))) == 0 {
			var senum uintptr
			if comCall(mgr, 5, uintptr(unsafe.Pointer(&senum))) == 0 {
				var nSes int32
				comCall(senum, 3, uintptr(unsafe.Pointer(&nSes)))
				for i := int32(0); i < nSes; i++ {
					var ctrl uintptr
					if comCall(senum, 4, uintptr(i), uintptr(unsafe.Pointer(&ctrl))) != 0 {
						continue
					}
					var ctrl2 uintptr
					if comCall(ctrl, 0, uintptr(unsafe.Pointer(&iidIAudioSessionControl2)), uintptr(unsafe.Pointer(&ctrl2))) == 0 {
						var pid uint32
						comCall(ctrl2, 14, uintptr(unsafe.Pointer(&pid)))
						if name := processName(pid); name != "" {
							fn(name, ctrl2)
						}
						comRelease(ctrl2)
					}
					comRelease(ctrl)
				}
				comRelease(senum)
			}
			comRelease(mgr)
		}
		comRelease(dev)
	}
}

// audioApps lists apps that currently have sound sessions
func audioApps() []string {
	seen := map[string]bool{}
	forEachSession(func(name string, _ uintptr) { seen[name] = true })
	var out []string
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// muteApp mutes or unmutes every sound session of an app (e.g. "opera.exe")
func muteApp(name string, mute bool) int {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return 0
	}
	if !strings.HasSuffix(name, ".exe") {
		name += ".exe"
	}
	count := 0
	forEachSession(func(n string, ctrl2 uintptr) {
		if n != name {
			return
		}
		var vol uintptr
		if comCall(ctrl2, 0, uintptr(unsafe.Pointer(&iidISimpleAudioVolume)), uintptr(unsafe.Pointer(&vol))) == 0 {
			b := uintptr(0)
			if mute {
				b = 1
			}
			comCall(vol, 5, b, 0)
			comRelease(vol)
			count++
		}
	})
	return count
}

// muteMic mutes or unmutes the default communications microphone
func muteMic(mute bool) {
	en := deviceEnumerator()
	if en == 0 {
		return
	}
	defer comRelease(en)
	var dev uintptr
	if comCall(en, 4, eCapture, eCommunications, uintptr(unsafe.Pointer(&dev))) != 0 {
		return
	}
	defer comRelease(dev)
	var ev uintptr
	if comCall(dev, 3, uintptr(unsafe.Pointer(&iidIAudioEndpointVolume)), clsctxAll, 0, uintptr(unsafe.Pointer(&ev))) != 0 {
		return
	}
	b := uintptr(0)
	if mute {
		b = 1
	}
	comCall(ev, 14, b, 0)
	comRelease(ev)
}
