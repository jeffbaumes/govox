package scene

import "github.com/go-gl/glfw/v3.2/glfw"

// FramebufferSize returns the framebuffer size.
// The framebuffer size may be larger than window pixel size for high DPI displays.
func FramebufferSize(w *glfw.Window) (fbw, fbh int) {
	fbw, fbh = w.GetFramebufferSize()
	return
}
