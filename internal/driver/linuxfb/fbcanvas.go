package linuxfb

import (
	"image"
	"image/draw"
	"sync"

	"fyne.io/fyne"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/internal"
	"fyne.io/fyne/internal/app"
	"fyne.io/fyne/theme"
)

var (
	dummyCanvas fyne.Canvas
)

// WindowlessCanvas provides functionality for a canvas to operate without a window
type WindowlessCanvas interface {
	fyne.Canvas

	FocusNext()
	FocusPrevious()
	Padded() bool
	Resize(fyne.Size)
	SetPadded(bool)
}

type fbCanvas struct {
	size  fyne.Size
	scale float32

	content  fyne.CanvasObject
	overlays *internal.OverlayStack
	focusMgr *app.FocusManager
	hovered  desktop.Hoverable
	padded   bool

	onTypedRune func(rune)
	onTypedKey  func(*fyne.KeyEvent)

	fyne.ShortcutHandler
	painter      SoftwarePainter
	propertyLock sync.RWMutex

	dirty      bool
	dirtyMutex *sync.Mutex
}

// Canvas returns a reusable in-memory canvas used for testing
func Canvas() fyne.Canvas {
	if dummyCanvas == nil {
		dummyCanvas = NewCanvas()
	}

	return dummyCanvas
}

// NewCanvas returns a single use in-memory canvas used for testing
func NewCanvas() WindowlessCanvas {
	c := &fbCanvas{
		focusMgr: app.NewFocusManager(nil),
		padded:   true,
		scale:    1.0,
		size:     fyne.NewSize(10, 10),
	}
	c.overlays = &internal.OverlayStack{Canvas: c}
	c.dirtyMutex = &sync.Mutex{}
	return c
}

// NewCanvasWithPainter allows creation of an in-memory canvas with a specific painter.
// The painter will be used to render in the Capture() call.
func NewCanvasWithPainter(painter SoftwarePainter) WindowlessCanvas {
	canvas := NewCanvas().(*fbCanvas)
	canvas.painter = painter

	return canvas
}

func (c *fbCanvas) Capture() image.Image {
	if c.painter != nil {
		return c.painter.Paint(c)
	}
	theme := fyne.CurrentApp().Settings().Theme()

	bounds := image.Rect(0, 0, internal.ScaleInt(c, c.Size().Width), internal.ScaleInt(c, c.Size().Height))
	img := image.NewNRGBA(bounds)
	draw.Draw(img, bounds, image.NewUniform(theme.BackgroundColor()), image.Point{}, draw.Src)

	return img
}

func (c *fbCanvas) Content() fyne.CanvasObject {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.content
}

func (c *fbCanvas) Focus(obj fyne.Focusable) {
	c.focusManager().Focus(obj)
}

func (c *fbCanvas) FocusNext() {
	c.focusManager().FocusNext()
}

func (c *fbCanvas) FocusPrevious() {
	c.focusManager().FocusPrevious()
}

func (c *fbCanvas) Focused() fyne.Focusable {
	return c.focusManager().Focused()
}

func (c *fbCanvas) InteractiveArea() (fyne.Position, fyne.Size) {
	return fyne.Position{}, c.Size()
}

func (c *fbCanvas) OnTypedKey() func(*fyne.KeyEvent) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.onTypedKey
}

func (c *fbCanvas) OnTypedRune() func(rune) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.onTypedRune
}

// Deprecated
func (c *fbCanvas) Overlay() fyne.CanvasObject {
	panic("deprecated method should not be used")
}

func (c *fbCanvas) Overlays() fyne.OverlayStack {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	return c.overlays
}

func (c *fbCanvas) Padded() bool {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.padded
}

func (c *fbCanvas) PixelCoordinateForPosition(pos fyne.Position) (int, int) {
	return int(float32(pos.X) * c.scale), int(float32(pos.Y) * c.scale)
}

func (c *fbCanvas) Refresh(fyne.CanvasObject) {
	c.setDirty(true)
}

func (c *fbCanvas) Resize(size fyne.Size) {
	c.propertyLock.Lock()
	content := c.content
	overlays := c.overlays
	padded := c.padded
	c.size = size
	c.propertyLock.Unlock()

	if content == nil {
		return
	}

	for _, overlay := range overlays.List() {
		overlay.Resize(size)
	}

	if padded {
		theme := fyne.CurrentApp().Settings().Theme()
		content.Resize(size.Subtract(fyne.NewSize(theme.Padding()*2, theme.Padding()*2)))
		content.Move(fyne.NewPos(theme.Padding(), theme.Padding()))
	} else {
		content.Resize(size)
		content.Move(fyne.NewPos(0, 0))
	}
}

func (c *fbCanvas) Scale() float32 {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.scale
}

func (c *fbCanvas) SetContent(content fyne.CanvasObject) {
	c.propertyLock.Lock()
	c.content = content
	c.focusMgr = app.NewFocusManager(c.content)
	c.propertyLock.Unlock()

	if content == nil {
		return
	}

	padding := fyne.NewSize(0, 0)
	if c.padded {
		padding = fyne.NewSize(theme.Padding()*2, theme.Padding()*2)
	}
	c.Resize(content.MinSize().Add(padding))
	c.setDirty(true)
}

func (c *fbCanvas) SetOnTypedKey(handler func(*fyne.KeyEvent)) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.onTypedKey = handler
}

func (c *fbCanvas) SetOnTypedRune(handler func(rune)) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.onTypedRune = handler
}

// Deprecated
func (c *fbCanvas) SetOverlay(_ fyne.CanvasObject) {
	panic("deprecated method should not be used")
}

func (c *fbCanvas) SetPadded(padded bool) {
	c.propertyLock.Lock()
	c.padded = padded
	c.propertyLock.Unlock()

	c.Resize(c.Size())
}

func (c *fbCanvas) SetScale(scale float32) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.scale = scale
	c.setDirty(true)
}

func (c *fbCanvas) Size() fyne.Size {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	return c.size
}

func (c *fbCanvas) Unfocus() {
	c.focusManager().Focus(nil)
}

func (c *fbCanvas) focusManager() *app.FocusManager {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if focusMgr := c.overlays.TopFocusManager(); focusMgr != nil {
		return focusMgr
	}
	return c.focusMgr
}

func (c *fbCanvas) objectTrees() []fyne.CanvasObject {
	trees := make([]fyne.CanvasObject, 0, len(c.Overlays().List())+1)
	if c.content != nil {
		trees = append(trees, c.content)
	}
	trees = append(trees, c.Overlays().List()...)
	return trees
}

func layoutAndCollect(objects []fyne.CanvasObject, o fyne.CanvasObject, size fyne.Size) []fyne.CanvasObject {
	objects = append(objects, o)
	switch c := o.(type) {
	case fyne.Widget:
		r := c.CreateRenderer()
		r.Layout(size)
		for _, child := range r.Objects() {
			objects = layoutAndCollect(objects, child, child.Size())
		}
	case *fyne.Container:
		if c.Layout != nil {
			c.Layout.Layout(c.Objects, size)
		}
		for _, child := range c.Objects {
			objects = layoutAndCollect(objects, child, child.Size())
		}
	}
	return objects
}

func (c *fbCanvas) isDirty() bool {
	c.dirtyMutex.Lock()
	defer c.dirtyMutex.Unlock()

	return c.dirty
}

func (c *fbCanvas) setDirty(dirty bool) {
	c.dirtyMutex.Lock()
	defer c.dirtyMutex.Unlock()

	c.dirty = dirty
}
