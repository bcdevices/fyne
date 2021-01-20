package fb565le

import (
	"os"
	"syscall"
	"unsafe"
)

func getFixScreenInfo(f *os.File) (fb_fix_screeninfo, error) {
	info := fb_fix_screeninfo{}
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(f.Fd()), _IOGET_FSCREENINFO, uintptr(unsafe.Pointer(&info))); errno != 0 {
		return info, errno
	}
	return info, nil
}

func getVarScreenInfo(f *os.File) (fb_var_screeninfo, error) {
	info := fb_var_screeninfo{}
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(f.Fd()), _IOGET_VSCREENINFO, uintptr(unsafe.Pointer(&info))); errno != 0 {
		return info, errno
	}
	return info, nil
}

func updateVarScreenInfo(f *os.File, update fb_var_screeninfo) (fb_var_screeninfo, error) {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(f.Fd()), _IOPUT_VSCREENINFO, uintptr(unsafe.Pointer(&update))); errno != 0 {
		return update, errno
	}
	info := fb_var_screeninfo{}
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(f.Fd()), _IOGET_VSCREENINFO, uintptr(unsafe.Pointer(&info))); errno != 0 {
		return info, errno
	}
	return info, nil
}
