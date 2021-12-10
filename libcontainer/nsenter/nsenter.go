//go:build linux && !gccgo
// +build linux,!gccgo

package nsenter

/*
#cgo CFLAGS: -Wall
#cgo pkg-config: libcap
extern void nsexec();
void __attribute__((constructor)) init(void) {
	nsexec();
}
*/
import "C"
