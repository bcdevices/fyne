package fb565le

// Based on https://github.com/gonutz/framebuffer
//
// Alternatives:
// - https://github.com/NeowayLabs/drm (DRM)
// - https://github.com/gen2brain/framebuffer (fbdev)
import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"syscall"

	//	"periph.io/x/periph/conn/display"

	"github.com/ev3go/ev3dev/fb"
)

// 128x128 (PLT-200A)
// 320x256 (PLT-300A 1/4)
// 640x512 (PLT-300A 1/4)
// 1280x1024 (PLT-300A native)

const (
	displayWidth  = 640
	displayHeight = 512
)

// Open expects a framebuffer device as its argument (such as "/dev/fb0"). The
// device will be memory-mapped to a local buffer. Writing to the device changes
// the screen output.
// The returned Device implements the draw.Image interface. This means that you
// can use it to copy to and from other images.
// The only supported color model for the specified frame buffer is RGB565.
// After you are done using the Device, call Close on it to unmap the memory and
// close the framebuffer file.
func Open(device string) (*Device, error) {
	file, err := os.OpenFile(device, os.O_RDWR, os.ModeDevice)
	if err != nil {
		return nil, err
	}

	varInfo, err := getVarScreenInfo(file)
	if err != nil {
		return nil, err
	}
	varInfo.bits_per_pixel = 16
	varInfo.xres = displayWidth
	varInfo.yres = displayHeight
	varInfo.xres_virtual = displayWidth
	varInfo.yres_virtual = displayHeight
	varInfo, err = updateVarScreenInfo(file, varInfo)
	if err != nil {
		return nil, err
	}
	fixInfo, err := getFixScreenInfo(file)
	if err != nil {
		return nil, err
	}

	if !(varInfo.red.offset == 11 && varInfo.red.length == 5 && varInfo.red.msb_right == 0 &&
		varInfo.green.offset == 5 && varInfo.green.length == 6 && varInfo.green.msb_right == 0 &&
		varInfo.blue.offset == 0 && varInfo.blue.length == 5 && varInfo.blue.msb_right == 0) {
		file.Close()
		return nil, errors.New("unsupported color model")
	}

	pixels, err := syscall.Mmap(
		int(file.Fd()),
		0, int(fixInfo.smemlen),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED,
	)
	if err != nil {
		file.Close()
		return nil, err
	}

	return &Device{
		file,
		pixels,
		int(fixInfo.line_length),
		image.Rect(0, 0, displayWidth, displayHeight),
		fb.RGB565Model,
	}, nil
}

// Device represents the frame buffer. It implements the draw.Image interface.
type Device struct {
	file       *os.File
	pixels     []byte
	pitch      int
	bounds     image.Rectangle
	colorModel color.Model
}

// Close unmaps the framebuffer memory and closes the device file. Call this
// function when you are done using the frame buffer.
func (d *Device) Close() {
	_ = syscall.Munmap(d.pixels)
	d.file.Close()
}

// Bounds implements the image.Image (and draw.Image) interface.
func (d *Device) Bounds() image.Rectangle {
	return d.bounds
}

// ColorModel implements the image.Image (and draw.Image) interface.
func (d *Device) ColorModel() color.Model {
	return d.colorModel
}

// At implements the image.Image (and draw.Image) interface.
func (d *Device) At(x, y int) color.Color {
	if x < d.bounds.Min.X || x >= d.bounds.Max.X ||
		y < d.bounds.Min.Y || y >= d.bounds.Max.Y {
		return fb.Pixel565(0)
	}
	i := y*d.pitch + 2*x
	return fb.Pixel565(binary.LittleEndian.Uint16(d.pixels[i : i+2]))
}

// Set implements the draw.Image interface.
func (d *Device) Set(x, y int, c color.Color) {
	// the min bounds are at 0,0 (see Open)
	if x >= 0 && x < d.bounds.Max.X &&
		y >= 0 && y < d.bounds.Max.Y {
		_, _, _, a := c.RGBA()
		if a > 0 {
			rgb := uint16(fb.RGB565Model.Convert(c).(fb.Pixel565))
			i := y*d.pitch + 2*x
			binary.LittleEndian.PutUint16(d.pixels[i:i+2], rgb)
		}
	}
}

// Draw implements devices.Display.
func (d *Device) Draw(r image.Rectangle, src image.Image, sp image.Point) error {
	if img, ok := src.(*fb.RGB565); ok &&
		r == d.Bounds() &&
		src.Bounds() == d.bounds &&
		sp.X == 0 &&
		sp.Y == 0 {
		// Fast path
		copy(d.pixels, img.Pix)
		return nil
	}
	draw.Draw(d, r, src, sp, draw.Src)
	return nil
}

// String implements devices.Display.
func (d *Device) String() string {
	return fmt.Sprintf("fb565.Dev{%s}", d.bounds.Max)
}

// Halt turns off the display.
//
// Sending any other command afterward reenables the display.
func (d *Device) Halt() error {
	return nil
}

//var _ display.Drawer = &Device{}
