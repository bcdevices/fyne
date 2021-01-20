package linuxfb

import "fyne.io/fyne"

type fbClipboard struct {
	content string
}

func (c *fbClipboard) Content() string {
	return c.content
}

func (c *fbClipboard) SetContent(content string) {
	c.content = content
}

// NewClipboard returns a single use in-memory clipboard used for testing
func NewClipboard() fyne.Clipboard {
	return &fbClipboard{}
}
