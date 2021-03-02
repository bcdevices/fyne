// This file is subject to a 1-clause BSD license.
// Its contents can be found in the enclosed LICENSE file.

package fb32bpp

// <linux/fb.h>

type fb_fix_screeninfo struct {
	_           [16]byte  // id; Identification string (e.g.: "TT Builtin")
	_           uintptr   // smemstart; Physical start address of framebuffer memory.
	smemlen     uint32    // Length of framebuffer memory.
	_           uint32    // type; See __TYPE_XXX values.
	_           uint32    // type_aux; Interleave for interleaved planes.
	_           uint32    // visual; See __VISUAL_XXX values.
	_           uint16    // xpanstep; Zero if no hardware panning.
	_           uint16    // ypanstep; Zero if no hardware panning.
	_           uint16    // ywrapstep; Zero if no hardware ywrap.
	line_length uint32    // Length of a line in bytes.
	_           uint64    // mmio_start; Physical start address of mmap'd _IO.
	_           uint32    // mmio_len; Length of mmap'd _IO.
	_           uint32    // accel; Indicate to driver which specific chip/card we have.
	_           uint16    // capabilities; See _CAP_XXXX values.
	_           [2]uint16 // Reserved for future use.
}

// Interpretation of offset for color fields: All offsets are from the right,
// inside a "pixel" value, which is exactly 'bits_per_pixel' wide (means: you
// can use the offset as right argument to <<). A pixel afterwards is a bit
// stream and is written to video memory as-is.
//
// For pseudocolor: offset and length should be the same for all color
// components. Offset specifies the position of the least significant bit
// of the pallette index in a pixel value. Length indicates the number
// of available palette entries (i.e. # of entries = 1 << length).
type fb_bitfield struct {
	offset    uint32 // beginning of bitfield
	length    uint32 // length of bitfield
	msb_right uint32 // != 0 : Most significant bit is right
}

type fb_var_screeninfo struct {
	xres           uint32 // Visible resolution.
	yres           uint32
	xres_virtual   uint32 // Virtual resolution (viewport).
	yres_virtual   uint32
	_              uint32      // xoffset; Offset from virtual to visible resolution.
	_              uint32      // yoffset
	bits_per_pixel uint32      // Bit depth.
	_              uint32      // grayscale; 0 = color, 1 = grayscale, >1 = FOURCC
	red            fb_bitfield // bitfield in FB mem, if true colour. Else only length is significant.
	green          fb_bitfield
	blue           fb_bitfield
	_              fb_bitfield // transparent
	_              uint32      // nonstd; non-zero = non-standard pixel format.
	_              uint32      // activate; See __ACTIVATE_XXXX values.
	_              uint32      // Height of picture in millimetres.
	_              uint32      // Width of picture in millimetres.
	_              uint32      // AccelFlags: obsolete
	_              uint32      // pixclock; Pixel clock in picoseconds.
	_              uint32      // left_margin; Time from sync to picture.
	_              uint32      // right_margin; Time from picture to sync.
	_              uint32      // upper_margin; Time from sync to picture.
	_              uint32      // lower_margin
	_              uint32      // hsync_len; Length of horizontal sync.
	_              uint32      // vsync_len; Length of vertical sync.
	_              uint32      // sync; See _SYNC_XXXX values.
	_              uint32      // vmode; See _VMODE_XXXX values.
	_              uint32      // rotate; Angle of counter-clockwise rotation.
	_              uint32      // colorspace; Colorspace for FOURCC-based modes.
	_              [4]uint32   // Reserved for future use.
}

// _IOCTL values
const (
	_IOGET_VSCREENINFO = 0x4600
	_IOPUT_VSCREENINFO = 0x4601
	_IOGET_FSCREENINFO = 0x4602
)
