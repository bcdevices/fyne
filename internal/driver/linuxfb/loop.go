package linuxfb

import (
	"image"
	"runtime"
	"sync"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/internal/driver/linuxfb/fb565le"
	"fyne.io/fyne/internal/painter"
)

type funcData struct {
	f    func()
	done chan bool
}

type drawData struct {
	f    func()
	win  *fbWindow
	done chan bool
}

// channel for queuing functions on the main thread
var funcQueue = make(chan funcData)
var drawFuncQueue = make(chan drawData)
var runFlag = false
var runMutex = &sync.Mutex{}
var initOnce = &sync.Once{}

// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
}

func running() bool {
	runMutex.Lock()
	defer runMutex.Unlock()
	return runFlag
}

// force a function f to run on the main thread
func runOnMain(f func()) {
	// If we are on main just execute - otherwise add it to the main queue and wait.
	// The "running" variable is normally false when we are on the main thread.
	if !running() {
		f()
	} else {
		done := make(chan bool)

		funcQueue <- funcData{f: f, done: done}
		<-done
	}
}

// force a function f to run on the draw thread
func runOnDraw(w *fbWindow, f func()) {
	done := make(chan bool)

	drawFuncQueue <- drawData{f: f, win: w, done: done}
	<-done
}

func (d *fbDriver) initFB() {
	fbDev, err := fb565le.Open("/dev/fb0")
	if err != nil {
		panic(err)
	}
	d.fbDev = fbDev
	initOnce.Do(func() {
		d.startDrawThread()
	})
}

func (d *fbDriver) runFB() {
	eventTick := time.NewTicker(time.Second / 60)
	runMutex.Lock()
	runFlag = true
	runMutex.Unlock()

	d.initFB()

	for {
		select {
		case <-d.done:
			eventTick.Stop()
			d.drawDone <- nil // wait for draw thread to stop
			return
		case f := <-funcQueue:
			f.f()
			if f.done != nil {
				f.done <- true
			}
		case <-eventTick.C:
			// no events expected from fbdev
		}
	}
}

func (d *fbDriver) repaintWindow(w *fbWindow) {
	canvas := w.canvas
	img := canvas.Capture()
	d.fbDev.Draw(img.Bounds(), img, image.Point{})
	canvas.setDirty(false)
}

func (d *fbDriver) startDrawThread() {
	settingsChange := make(chan fyne.Settings)
	fyne.CurrentApp().Settings().AddChangeListener(settingsChange)
	draw := time.NewTicker(time.Second / 60)

	go func() {
		runtime.LockOSThread()

		for {
			select {
			case <-d.drawDone:
				return
			case f := <-drawFuncQueue:
				f.f()
				if f.done != nil {
					f.done <- true
				}
			case <-settingsChange:
				painter.ClearFontCache()
			case <-draw.C:
				for _, win := range d.windowList() {
					w := win.(*fbWindow)
					w.viewLock.RLock()
					canvas := w.canvas
					visible := w.visible
					w.viewLock.RUnlock()
					if !canvas.isDirty() || !visible {
						continue
					}

					d.repaintWindow(w)
				}
			}
		}
	}()
}

// refreshWindow requests that the specified window be redrawn
func refreshWindow(w *fbWindow) {
	w.canvas.setDirty(true)
}
