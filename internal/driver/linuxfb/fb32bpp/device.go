package fb32bpp

// Based on https://github.com/gonutz/framebuffer
//
// Alternatives:
// - https://github.com/NeowayLabs/drm (DRM)
// - https://github.com/gen2brain/framebuffer (fbdev)
//
// FB must be configured prior to use:
//
// ```
// fbset -fb /dev/fb0 -xres 1080 -yres 1920 -match -depth 32
// ```

import (
	//"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"syscall"
	//"github.com/ev3go/ev3dev/fb"
)

// Open expects a framebuffer device as its argument (such as "/dev/fb0"). The
// device will be memory-mapped to a local buffer. Writing to the device changes
// the screen output.
// The returned Device implements the draw.Image interface. This means that you
// can use it to copy to and from other images.
// The only supported color model for the specified frame buffer is NRGBA.
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

	w := varInfo.xres_virtual
	h := varInfo.yres_virtual

	if varInfo.bits_per_pixel != 32 {
		return nil, errors.New("unsupported bit depth")
	}

	fixInfo, err := getFixScreenInfo(file)
	if err != nil {
		return nil, err
	}

	if !(varInfo.red.offset == 16 && varInfo.red.length == 8 && varInfo.red.msb_right == 0 &&
		varInfo.green.offset == 8 && varInfo.green.length == 8 && varInfo.green.msb_right == 0 &&
		varInfo.blue.offset == 0 && varInfo.blue.length == 8 && varInfo.blue.msb_right == 0) {
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

	if h > w {
		return &Device{
			file,
			pixels,
			int(fixInfo.line_length),
			image.Rect(0, 0, int(h), int(w)),
			color.RGBAModel,
			true,
		}, nil
	}

	return &Device{
		file,
		pixels,
		int(fixInfo.line_length),
		image.Rect(0, 0, int(w), int(h)),
		color.RGBAModel,
		false,
	}, nil
}

// Device represents the frame buffer. It implements the draw.Image interface.
type Device struct {
	file       *os.File
	pixels     []byte
	pitch      int
	bounds     image.Rectangle
	colorModel color.Model
	flipped    bool
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
		return color.NRGBA{0, 0, 0, 0}
	}
	var i int = 0
	if d.flipped {
		i = x*d.pitch + 4*y
	} else {
		i = y*d.pitch + 4*x
	}
	return color.NRGBA{
		R: d.pixels[i+2],
		G: d.pixels[i+1],
		B: d.pixels[i],
		A: d.pixels[i+3],
	}
}

// Set implements the draw.Image interface.
func (d *Device) Set(x, y int, c color.Color) {
	// the min bounds are at 0,0 (see Open)
	if x >= 0 && x < d.bounds.Max.X &&
		y >= 0 && y < d.bounds.Max.Y {
		_, _, _, a := c.RGBA()
		if a > 0 {
			rgb := color.NRGBAModel.Convert(c).(color.NRGBA)
			var i int
			if d.flipped {
				i = x*d.pitch + 4*(d.bounds.Max.Y - y)
			} else {
				i = y*d.pitch + 4*x
			}
			d.pixels[i] = rgb.B
			d.pixels[i+1] = rgb.G
			d.pixels[i+2] = rgb.R
			d.pixels[i+3] = rgb.A
		}
	}
}

// Draw implements devices.Display.
func (d *Device) Draw(r image.Rectangle, src image.Image, sp image.Point) error {
	if d.flipped {
		draw.Draw(d, r, src, sp, draw.Src)
		return nil
	}
	if img, ok := src.(*image.NRGBA); ok &&
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
	return fmt.Sprintf("fb32bpp.Dev{%s}", d.bounds.Max)
}

// Halt turns off the display.
//
// Sending any other command afterward reenables the display.
func (d *Device) Halt() error {
	return nil
}

//var _ display.Drawer = &Device{}
