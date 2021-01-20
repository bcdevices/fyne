package linuxfb

import (
	"fyne.io/fyne"
	"sync"
)

type fbWindow struct {
	title              string
	fullScreen         bool
	fixedSize          bool
	focused            bool
	onClosed           func()
	onCloseIntercepted func()

	canvas    *fbCanvas
	clipboard fyne.Clipboard
	driver    *fbDriver
	menu      *fyne.MainMenu

	viewLock   sync.RWMutex
	createLock sync.Once
	decorate   bool
	master     bool
	centered   bool
	visible    bool
}

// NewWindow creates and registers a new window for test purposes
//func NewWindow(content fyne.CanvasObject) fyne.Window {
//	window := fyne.CurrentApp().NewWindow("")
//	window.SetContent(content)
//	return window
//}

func (w *fbWindow) Canvas() fyne.Canvas {
	return w.canvas
}

func (w *fbWindow) CenterOnScreen() {
	// no-op
}

func (w *fbWindow) Clipboard() fyne.Clipboard {
	return w.clipboard
}

func (w *fbWindow) Close() {
	if w.onClosed != nil {
		w.onClosed()
	}
	w.focused = false
	w.driver.removeWindow(w)
}

func (w *fbWindow) Content() fyne.CanvasObject {
	return w.Canvas().Content()
}

func (w *fbWindow) FixedSize() bool {
	return w.fixedSize
}

func (w *fbWindow) FullScreen() bool {
	return w.fullScreen
}

func (w *fbWindow) Hide() {
	w.focused = false
	w.viewLock.Lock()
	w.visible = false
	w.viewLock.Unlock()

	// hide top canvas element
	if w.canvas.Content() != nil {
		w.canvas.Content().Hide()
	}

}

func (w *fbWindow) Icon() fyne.Resource {
	return fyne.CurrentApp().Icon()
}

func (w *fbWindow) MainMenu() *fyne.MainMenu {
	return w.menu
}

func (w *fbWindow) Padded() bool {
	return w.canvas.Padded()
}

func (w *fbWindow) RequestFocus() {
	for _, win := range w.driver.AllWindows() {
		win.(*fbWindow).focused = false
	}

	w.focused = true
}

func (w *fbWindow) Resize(size fyne.Size) {
	w.canvas.Resize(size)
}

func (w *fbWindow) SetContent(obj fyne.CanvasObject) {
	w.Canvas().SetContent(obj)
}

func (w *fbWindow) SetFixedSize(fixed bool) {
	w.fixedSize = fixed
}

func (w *fbWindow) SetIcon(_ fyne.Resource) {
	// no-op
}

func (w *fbWindow) SetFullScreen(fullScreen bool) {
	w.fullScreen = fullScreen
}

func (w *fbWindow) SetMainMenu(menu *fyne.MainMenu) {
	w.menu = menu
}

func (w *fbWindow) SetMaster() {
	// no-op
}

func (w *fbWindow) SetOnClosed(closed func()) {
	w.onClosed = closed
}

func (w *fbWindow) SetCloseIntercept(callback func()) {
	w.onCloseIntercepted = callback
}

func (w *fbWindow) SetPadded(padded bool) {
	w.canvas.SetPadded(padded)
}

func (w *fbWindow) SetTitle(title string) {
	w.title = title
}

func (w *fbWindow) Show() {
	w.RequestFocus()
	w.viewLock.Lock()
	w.visible = true
	w.viewLock.Unlock()
	if w.canvas.Content() != nil {
		w.canvas.Content().Show()
	}
}

func (w *fbWindow) ShowAndRun() {
	w.Show()
	w.driver.Run()
	//fyne.CurrentApp().Driver().Run()
}

func (w *fbWindow) Title() string {
	return w.title
}
