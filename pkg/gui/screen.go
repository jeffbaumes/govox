package gui

import (
	"math"

	"github.com/go-gl/glfw/v3.2/glfw"
)

// Screen is a GUI manager associated with a GLFW window
type Screen struct {
	Window                                           *glfw.Window
	labels                                           []*Label
	buttons                                          []*Button
	entrys                                           []*Entry
	text                                             *Text
	Xpos, Ypos, MaxTextSize                          float64
	fontpngpath, fontjsonpath, buttonpath, entrypath string
	shift                                            bool
	keys                                             string
	keysShift                                        string
	keysShiftMap                                     map[byte]byte
	Key                                              string
}

// InitGui initializes configuration parameters
func (screen *Screen) InitGui(fontpngpath, fontjsonpath, buttonpath, entrypath string, maxtextsize float64) {
	screen.fontpngpath = fontpngpath
	screen.fontjsonpath = fontjsonpath
	screen.buttonpath = buttonpath
	screen.entrypath = entrypath
	screen.MaxTextSize = maxtextsize
	screen.text = NewText(screen)
	screen.keys = "`1234567890-=qwertyuiop[]\\asdfghjkl;'zxcvbnm,./ "
	screen.keysShift = "~!@#$%^&*()_+QWERTYUIOP{}|ASDFGHJKL:\"ZXCVBNM<>? "
	screen.keysShiftMap = make(map[byte]byte)
	for i := 0; i < len(screen.keys); i++ {
		screen.keysShiftMap[screen.keys[i]] = screen.keysShift[i]
	}
	println(len(screen.keys), len(screen.keysShift))
}

// Clear removes all elements from the screen
func (screen *Screen) Clear() {
	screen.entrys = nil
	screen.buttons = nil
	screen.labels = nil
}

// NewScreen creates a new GUI screen associated with a GLFW window
func NewScreen(window *glfw.Window) *Screen {
	s := Screen{}
	s.Window = window
	return &s
}

// Update updates the GUI
func (screen *Screen) Update() {
	for _, label := range screen.labels {
		if !label.Hide {
			screen.text.draw(label.Text, label.X, label.Y, label.Size, false, false, screen.Window)
		}
	}
	for _, button := range screen.buttons {
		button.draw(screen)
		if !button.Hide {
			screen.text.draw(button.Text, button.X+button.W/2, button.Y+button.H/2, button.textSize(screen), true, true, screen.Window)
		}
	}
	for _, entry := range screen.entrys {
		if !entry.Hide {
			entry.draw(screen)
			screen.text.draw(entry.Text, entry.X, entry.Y+entry.H/2, math.Min(entry.textSize(screen), screen.MaxTextSize), false, true, screen.Window)
		}
	}
}

// KeyName returns a human-readable name for a GLFW key
func KeyName(key glfw.Key) string {
	name := glfw.GetKeyName(key, 0)
	if name != "" && name != " " {
		return name
	}
	switch key {
	case glfw.KeySpace:
		return "Space"
	case glfw.KeyLeftShift:
		return "LShift"
	case glfw.KeyRightShift:
		return "RShift"
	case glfw.KeyLeftControl:
		return "LControl"
	case glfw.KeyRightControl:
		return "RControl"
	case glfw.KeyLeftAlt:
		return "LAlt"
	case glfw.KeyRightAlt:
		return "RAlt"
	case glfw.KeyLeftSuper:
		return "LSuper"
	case glfw.KeyRightSuper:
		return "RSuper"
	case glfw.Key(0):
		return "LButton"
	case glfw.Key(1):
		return "RButton"
	case glfw.Key(2):
		return "MButton"
	}
	return "Unknown"
}

// MouseButtonCallback is the GLFW mouse button callback for the GUI
func (screen *Screen) MouseButtonCallback() func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	return func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		wd, ht := w.GetSize()
		x, y := screen.Xpos/float64(wd)*2-1, -(screen.Ypos/float64(ht)*2 - 1)
		for _, b := range screen.buttons {
			if b.isInside(x, y) && button == glfw.MouseButtonLeft && action == glfw.Press {
				b.Command()
			}
		}
		for _, e := range screen.entrys {
			if e.KeyHolder && e.Focus {
				if action == glfw.Press {
					if button == glfw.MouseButtonRight {
						e.Key = glfw.Key(button)
						e.Text = KeyName(e.Key)
						e.Focus = false
					} else if button == glfw.MouseButtonLeft {
						e.Key = glfw.Key(button)
						e.Text = KeyName(e.Key)
						e.Focus = false
					}
				}
			}
			if e.isInside(x, y) && button == glfw.MouseButtonLeft && action == glfw.Press {
				e.Focus = true
			} else if button == glfw.MouseButtonLeft && action == glfw.Press {
				e.Focus = false
			}
		}
	}
}

// CursorPosCallback is the GLFW cursor position callback for the GUI
func (screen *Screen) CursorPosCallback() func(w *glfw.Window, xpos, ypos float64) {
	return func(w *glfw.Window, xpos, ypos float64) {
		screen.Xpos = xpos
		screen.Ypos = ypos
		wd, ht := w.GetSize()
		// wd, ht := FramebufferSize(w)
		x, y := screen.Xpos/float64(wd)*2-1, -(screen.Ypos/float64(ht)*2 - 1)
		for _, b := range screen.buttons {
			b.mouseover = b.isInside(x, y)
		}
		for _, e := range screen.entrys {
			e.mouseover = e.isInside(x, y)
		}
	}
}

// KeyCallBack is the GLFW key callback for the GUI
func (screen *Screen) KeyCallBack() func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	return func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action != glfw.Press {
			return
		}
		for _, e := range screen.entrys {
			if !e.Focus {
				continue
			}
			if e.KeyHolder {
				keyName := KeyName(key)
				e.Key = key
				e.Text = keyName
				e.Focus = false
				if e.Command != nil {
					e.Command()
				}
			} else {
				switch {
				case key == glfw.KeyBackspace:
					e.Text = e.Text[0:max(len(e.Text)-1, 0)]
				case key == glfw.KeyEnter && e.Command != nil:
					e.Command()
				case key == glfw.KeyEscape:
					e.Focus = false
				default:
					keyName := glfw.GetKeyName(key, 0)
					if keyName != "" {
						if mods&glfw.ModShift != 0 {
							keyName = string(screen.keysShiftMap[keyName[0]])
						}
						e.Text += keyName
					}
				}
			}
		}
	}
}
