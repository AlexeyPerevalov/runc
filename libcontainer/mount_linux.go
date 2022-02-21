package libcontainer

import (
	"strconv"

	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	OPEN_TREE_CLONE         = 1
	OPEN_TREE_CLOEXEC       = unix.O_CLOEXEC
	MOVE_MOUNT_F_EMPTY_PATH = 0x00000004
)

const (
	FSCONFIG_SET_FLAG        = 0 /* Set parameter, supplying no value */
	FSCONFIG_SET_STRING      = 1 /* Set parameter, supplying a string value */
	FSCONFIG_SET_BINARY      = 2 /* Set parameter, supplying a binary blob value */
	FSCONFIG_SET_PATH        = 3 /* Set parameter, supplying an object by path */
	FSCONFIG_SET_PATH_EMPTY  = 4 /* Set parameter, supplying an object by (empty) path */
	FCONFIG_SET_FD           = 5 /* Set parameter, supplying an object by fd */
	FSCONFIG_CMD_CREATE      = 6 /* Invoke superblock creation */
	FSCONFIG_CMD_RECONFIGURE = 7 /* Invoke superblock reconfiguration */

)

// mountError holds an error from a failed mount or unmount operation.
type mountError struct {
	op     string
	source string
	target string
	procfd string
	flags  uintptr
	data   string
	err    error
}

// Error provides a string error representation.
func (e *mountError) Error() string {
	out := e.op + " "

	if e.source != "" {
		out += e.source + ":" + e.target
	} else {
		out += e.target
	}
	if e.procfd != "" {
		out += " (via " + e.procfd + ")"
	}

	if e.flags != uintptr(0) {
		out += ", flags: 0x" + strconv.FormatUint(uint64(e.flags), 16)
	}
	if e.data != "" {
		out += ", data: " + e.data
	}

	out += ": " + e.err.Error()
	return out
}

// Unwrap returns the underlying error.
// This is a convention used by Go 1.13+ standard library.
func (e *mountError) Unwrap() error {
	return e.err
}

// mount is a simple unix.Mount wrapper. If procfd is not empty, it is used
// instead of target (and the target is only used to add context to an error).
func mount(source, target, procfd, fstype string, flags uintptr, data string) error {
	dst := target
	if procfd != "" {
		dst = procfd
	}
	if err := unix.Mount(source, dst, fstype, flags, data); err != nil {
		return &mountError{
			op:     "mount",
			source: source,
			target: target,
			procfd: procfd,
			flags:  flags,
			data:   data,
			err:    err,
		}
	}
	return nil
}

// unmount is a simple unix.Unmount wrapper.
func unmount(target string, flags int) error {
	err := unix.Unmount(target, flags)
	if err != nil {
		return &mountError{
			op:     "unmount",
			target: target,
			flags:  uintptr(flags),
			err:    err,
		}
	}
	return nil
}

//move_mount(int from_dfd, const char *from_pathname, int to_dfd,
// const char *to_pathname, unsigned int flags)
func moveMount(from_dfd int, fromPathName string, to_dfd int, toPathName string, flags uint) (err error) {
	var _p0, _p1 *byte
	_p0, err = syscall.BytePtrFromString(fromPathName)
	if err != nil {
		return
	}
	_p1, err = syscall.BytePtrFromString(toPathName)
	if err != nil {
		return
	}
	_, _, e1 := syscall.Syscall6(unix.SYS_MOVE_MOUNT, uintptr(from_dfd), uintptr(unsafe.Pointer(_p0)), uintptr(to_dfd), uintptr(unsafe.Pointer(_p1)), uintptr(flags), 0)
	if e1 != 0 {
		err = error(e1)
	}
	return
}

// int dfd, const char *filename, unsigned int flags)
func openTree(dfd int, fileName string, flags uint) (r int, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(fileName)
	if err != nil {
		return
	}
	r1, _, e1 := syscall.Syscall6(unix.SYS_OPEN_TREE, uintptr(dfd), uintptr(unsafe.Pointer(_p0)), uintptr(flags), 0, 0, 0)
	if e1 != 0 {
		err = error(e1)
	}
	r = int(r1)
	return
}

func fsOpen(fstype string, flags uint) (r int, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(fstype)
	if err != nil {
		return
	}

	r1, _, e1 := syscall.Syscall(unix.SYS_FSOPEN, uintptr(unsafe.Pointer(_p0)), uintptr(flags), 0)
	if e1 != 0 {
		err = error(e1)
	}
	r = int(r1)
	return
}

func fsMount(mfd int, flags uint, ms_flags uint) (r int, err error) {
	r1, _, e1 := syscall.Syscall(unix.SYS_FSMOUNT, uintptr(flags), uintptr(ms_flags), 0)
	if e1 != 0 {
		err = error(e1)
	}
	r = int(r1)
	return
}

func fsConfig(fs_fd int, cmd int, key string, val string, aux int) (err error) {
	var (
		_p0 *byte
		_p1 *byte
	)
	_p0, err = syscall.BytePtrFromString(key)
	if err != nil {
		return
	}

	_p1, err = syscall.BytePtrFromString(val)
	if err != nil {
		return
	}

	_, _, e1 := syscall.Syscall6(unix.SYS_FSCONFIG, uintptr(fs_fd), uintptr(cmd), uintptr(unsafe.Pointer(_p0)), uintptr(unsafe.Pointer(_p1)), uintptr(aux), 0)
	if e1 != 0 {
		err = error(e1)
	}
	return
}
