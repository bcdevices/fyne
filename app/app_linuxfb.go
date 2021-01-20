// +build linuxfb

package app

import (
	"fyne.io/fyne"
	"fyne.io/fyne/internal/driver/linuxfb"
)

// NewWithID returns a new app instance using the linuxfb driver.
// The ID string should be globally unique to this app.
func NewWithID(id string) fyne.App {
	return newAppWithDriver(linuxfb.NewFBDriver(), id)
}
