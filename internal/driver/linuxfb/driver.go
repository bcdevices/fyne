package linuxfb

import (
	"fmt"
	"image"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"fyne.io/fyne"
	"fyne.io/fyne/internal/driver"
	"fyne.io/fyne/internal/driver/linuxfb/fb565le"
	"fyne.io/fyne/internal/painter"
	"fyne.io/fyne/internal/painter/software"

	"github.com/goki/freetype/truetype"
	"golang.org/x/image/font"
)

const mainGoroutineID = 1

// SoftwarePainter describes a simple type that can render canvases
type SoftwarePainter interface {
	Paint(fyne.Canvas) image.Image
}

type fbDriver struct {
	fbDev        *fb565le.Device
	device       *device
	painter      SoftwarePainter
	windows      []fyne.Window
	windowsMutex sync.RWMutex
	done         chan interface{}
	drawDone     chan interface{}
}

// Declare conformity with Driver
var _ fyne.Driver = (*fbDriver)(nil)

func NewFBDriver() fyne.Driver {
	drv := new(fbDriver)
	drv.painter = software.NewPainter()
	drv.windowsMutex = sync.RWMutex{}

	return drv
}

func (d *fbDriver) AbsolutePositionForObject(co fyne.CanvasObject) fyne.Position {
	c := d.CanvasForObject(co)
	if c == nil {
		return fyne.NewPos(0, 0)
	}

	tc := c.(*fbCanvas)
	return driver.AbsolutePositionForObject(co, tc.objectTrees())
}

func (d *fbDriver) AllWindows() []fyne.Window {
	d.windowsMutex.RLock()
	defer d.windowsMutex.RUnlock()
	return d.windows
}

func (d *fbDriver) CanvasForObject(fyne.CanvasObject) fyne.Canvas {
	d.windowsMutex.RLock()
	defer d.windowsMutex.RUnlock()
	// cheating: probably the last created window is meant
	return d.windows[len(d.windows)-1].Canvas()
}

func (d *fbDriver) CreateWindow(string) fyne.Window {
	canvas := NewCanvas().(*fbCanvas)
	if d.painter != nil {
		canvas.painter = d.painter
	} else {
		canvas.painter = software.NewPainter()
	}

	window := &fbWindow{canvas: canvas, driver: d}
	window.clipboard = &fbClipboard{}

	d.windowsMutex.Lock()
	d.windows = append(d.windows, window)
	d.windowsMutex.Unlock()
	return window
}

func (d *fbDriver) Device() fyne.Device {
	if d.device == nil {
		d.device = &device{}
	}
	return d.device
}

// RenderedTextSize looks up how bit a string would be if drawn on screen
func (d *fbDriver) RenderedTextSize(text string, size int, style fyne.TextStyle) fyne.Size {
	var opts truetype.Options
	opts.Size = float64(size)
	opts.DPI = painter.TextDPI

	face := painter.CachedFontFace(style, &opts)
	advance := font.MeasureString(face, text)

	sws := fyne.NewSize(advance.Ceil(), face.Metrics().Height.Ceil())
	gls := painter.RenderedTextSize(text, size, style)
	if sws != gls {
		log.Println("SoftwareTextSize:", sws)
		log.Println("GLTextSize:", gls)
	}
	return sws
}

func goroutineID() int {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	// string format expects "goroutine X [running..."
	id := strings.Split(strings.TrimSpace(string(b)), " ")[1]

	num, _ := strconv.Atoi(id)
	return num
}

func (d *fbDriver) Run() {
	if goroutineID() != mainGoroutineID {
		panic("Run() or ShowAndRun() must be called from main goroutine")
	}
	d.runFB()
}

func (d *fbDriver) Quit() {
	fmt.Printf("Quit.\n")
	// no-op
}

func (d *fbDriver) removeWindow(w *fbWindow) {
	d.windowsMutex.Lock()
	i := 0
	for _, window := range d.windows {
		if window == w {
			break
		}
		i++
	}

	d.windows = append(d.windows[:i], d.windows[i+1:]...)
	d.windowsMutex.Unlock()
}

func (d *fbDriver) windowList() []fyne.Window {
	d.windowsMutex.RLock()
	defer d.windowsMutex.RUnlock()
	return d.windows
}
