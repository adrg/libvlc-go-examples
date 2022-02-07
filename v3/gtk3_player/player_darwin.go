package main

/*
#cgo CFLAGS: -x objective-c
#cgo pkg-config: gdk-3.0
#include <AppKit/AppKit.h>
#include <gdk/gdk.h>

GDK_AVAILABLE_IN_ALL NSView* gdk_quartz_window_get_nsview(GdkWindow *window);
*/
import "C"
import (
	"unsafe"

	vlc "github.com/adrg/libvlc-go/v3"
	"github.com/gotk3/gotk3/gdk"
)

func setPlayerWindow(player *vlc.Player, window *gdk.Window) error {
	handle := unsafe.Pointer(C.gdk_quartz_window_get_nsview((*C.GdkWindow)(unsafe.Pointer(window.GObject))))
	return player.SetNSObject(uintptr(handle))
}
