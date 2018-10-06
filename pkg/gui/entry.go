package gui

import (
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

//The Entry struct is a struct with the text, x, y, width (W), height (H), border (C), and focus which is if someone clicked on it
type Entry struct {
	Text           string
	X, Y, W, H, C  float64
	mouseover      bool
	program        uint32
	drawableVAO    uint32
	pointsVBO      uint32
	textureUniform int32
	texture        uint32
	textureUnit    int32
	colorlocation  int32
	Focus          bool
	Command        func()
	Screen         *Screen
	index          int
	Hide           bool
	KeyHolder      bool
	Key            glfw.Key
}

const entryvertexShader = `
#version 410

in vec4 coord;
out vec2 texcoord;

void main(void) {
	gl_Position = vec4(coord.xy, -0.03, 1);
	texcoord = coord.zw;
}
`

const entryfragmentShader = `
#version 410
uniform vec4 color;
in vec2 texcoord;
uniform sampler2D texFont;
out vec4 frag_color;

void main(void) {
	vec4 texel = texture(texFont, texcoord);
	if (texel.a < 0.5) {
		discard;
	}
	frag_color = texel * color;
}
`

// Remove removes the entry from the screen.
func (e *Entry) Remove() {
	e.Screen.entrys[len(e.Screen.entrys)-1], e.Screen.entrys[e.index] = e.Screen.entrys[e.index], e.Screen.entrys[len(e.Screen.entrys)-1]
	e.Screen.entrys = e.Screen.entrys[:len(e.Screen.entrys)-1]
}

func (e *Entry) textSize(screen *Screen) (size float64) {
	wi, ht := FramebufferSize(screen.Window)
	y := e.H / 2
	x := e.W / float64(len(e.Text)) * 19 / 12 * float64(wi) / float64(ht)
	size = math.Min(x, y)
	return
}

func (e *Entry) draw(screen *Screen) {
	if !e.Hide {
		gl.UseProgram(e.program)
		wi, ht := FramebufferSize(screen.Window)
		cy := e.C
		if e.Focus {
			cy *= 1.20
		}
		cx := cy * float64(ht) / float64(wi)
		points := []float32{
			//top left
			float32(e.X), float32(e.Y), 0, 0,
			float32(e.X + cx), float32(e.Y), 0.33, 0,
			float32(e.X), float32(e.Y + cy), 0, 0.33,
			float32(e.X), float32(e.Y + cy), 0, 0.33,
			float32(e.X + cx), float32(e.Y + cy), 0.33, 0.33,
			float32(e.X + cx), float32(e.Y), 0.33, 0,
			//top middle
			float32(e.X + cx), float32(e.Y), 0.33, 0,
			float32(e.X + e.W - cx), float32(e.Y), 0.67, 0,
			float32(e.X + cx), float32(e.Y + cy), 0.33, 0.33,
			float32(e.X + cx), float32(e.Y + cy), 0.33, 0.33,
			float32(e.X + e.W - cx), float32(e.Y), 0.67, 0,
			float32(e.X + e.W - cx), float32(e.Y + cy), 0.67, 0.33,
			//top right
			float32(e.X + e.W - cx), float32(e.Y), 0.67, 0,
			float32(e.X + e.W), float32(e.Y), 1, 0,
			float32(e.X + e.W - cx), float32(e.Y + cy), 0.67, 0.33,
			float32(e.X + e.W - cx), float32(e.Y + cy), 0.67, 0.33,
			float32(e.X + e.W), float32(e.Y), 1, 0,
			float32(e.X + e.W), float32(e.Y + cy), 1, 0.33,
			//side right
			float32(e.X), float32(e.Y + cy), 0, 0.33,
			float32(e.X), float32(e.Y + e.H - cy), 0, 0.67,
			float32(e.X + cx), float32(e.Y + cy), 0.33, 0.33,
			float32(e.X + cx), float32(e.Y + cy), 0.33, 0.33,
			float32(e.X), float32(e.Y + e.H - cy), 0, 0.67,
			float32(e.X + cx), float32(e.Y + e.H - cy), 0.33, 0.67,
			//middle
			float32(e.X + cx), float32(e.Y + cy), 0.33, 0.33,
			float32(e.X + e.W - cx), float32(e.Y + cy), 0.67, 0.33,
			float32(e.X + cx), float32(e.Y + e.H - cy), 0.33, 0.67,
			float32(e.X + e.W - cx), float32(e.Y + cy), 0.67, 0.33,
			float32(e.X + cx), float32(e.Y + e.H - cy), 0.33, 0.67,
			float32(e.X + e.W - cx), float32(e.Y + e.H - cy), 0.67, 0.67,
			//side left
			float32(e.X + e.W - cx), float32(e.Y + cy), 0.67, 0.33,
			float32(e.X + e.W - cx), float32(e.Y + e.H - cy), 0.67, 0.67,
			float32(e.X + e.W), float32(e.Y + cy), 1, 0.33,
			float32(e.X + e.W - cx), float32(e.Y + e.H - cy), 0.67, 0.67,
			float32(e.X + e.W), float32(e.Y + cy), 1, 0.33,
			float32(e.X + e.W), float32(e.Y + e.H - cy), 1, 0.67,
			//bottom left
			float32(e.X), float32(e.Y + e.H - cy), 0, 0.67,
			float32(e.X), float32(e.Y + e.H), 0, 1,
			float32(e.X + cx), float32(e.Y + e.H - cy), 0.33, 0.67,
			float32(e.X), float32(e.Y + e.H), 0, 1,
			float32(e.X + cx), float32(e.Y + e.H - cy), 0.33, 0.67,
			float32(e.X + cx), float32(e.Y + e.H), 0.33, 1,
			//bottom middle
			float32(e.X + cx), float32(e.Y + e.H - cy), 0.33, 0.67,
			float32(e.X + e.W - cx), float32(e.Y + e.H - cy), 0.67, 0.67,
			float32(e.X + cx), float32(e.Y + e.H), 0.33, 1,
			float32(e.X + cx), float32(e.Y + e.H), 0.33, 1,
			float32(e.X + e.W - cx), float32(e.Y + e.H - cy), 0.67, 0.67,
			float32(e.X + e.W - cx), float32(e.Y + e.H), 0.67, 1,
			//bottom left
			float32(e.X + e.W - cx), float32(e.Y + e.H - cy), 0.67, 0.67,
			float32(e.X + e.W), float32(e.Y + e.H - cy), 1, 0.67,
			float32(e.X + e.W - cx), float32(e.Y + e.H), 0.67, 1,
			float32(e.X + e.W - cx), float32(e.Y + e.H), 0.67, 1,
			float32(e.X + e.W), float32(e.Y + e.H - cy), 1, 0.67,
			float32(e.X + e.W), float32(e.Y + e.H), 1, 1,
		}
		fillVBO(e.pointsVBO, points)
		if e.mouseover {
			gl.Uniform4f(e.colorlocation, 0.9, 0.9, 0.9, 1.0)
		} else {
			gl.Uniform4f(e.colorlocation, 1.0, 1.0, 1.0, 1.0)
		}
		gl.Uniform1i(e.textureUniform, e.textureUnit)
		gl.BindVertexArray(e.drawableVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 6*9)
	}
}
func (e *Entry) isInside(x, y float64) bool {
	return x >= e.X && x <= e.X+e.W && y >= e.Y && y <= e.Y+e.H

}

// NewEntry Makes a Entry on the Screen.
//
// Screen is the screen struct you will put it on.
//
// Text is the text that will go in it (if you want starting text).
//
// x, y, width (w), height (h), border, these should be self explanatory.
//
// command is the command you want it to run when you press enter put nil for none
//
// and it return a entry object
func NewEntry(screen *Screen, text string, x, y, w, h, border float64, command func()) *Entry {
	e := Entry{}
	e.Text = text
	e.KeyHolder = false
	e.X = x
	e.Y = y
	e.W = w
	e.H = h
	e.C = border
	e.Command = command
	e.pointsVBO = newVBO()
	e.drawableVAO = newPointsVAO(e.pointsVBO, 4)
	e.program = createProgram(entryvertexShader, entryfragmentShader)
	e.colorlocation = uniformLocation(e.program, "color")
	e.textureUniform = uniformLocation(e.program, "texFont")
	bindAttribute(e.program, 0, "coord")
	existingImageFile, err := os.Open(screen.entrypath)
	if err != nil {
		panic(err)
	}
	defer existingImageFile.Close()
	img, err := png.Decode(existingImageFile)
	if err != nil {
		panic(err)
	}
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Pt(0, 0), draw.Src)

	e.textureUnit = 6
	gl.ActiveTexture(uint32(gl.TEXTURE0 + e.textureUnit))
	gl.GenTextures(1, &e.texture)
	gl.BindTexture(gl.TEXTURE_2D, e.texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		// gl.SRGB_ALPHA,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)

	gl.GenerateMipmap(gl.TEXTURE_2D)
	e.Screen = screen
	e.index = len(screen.entrys) - 1
	screen.entrys = append(screen.entrys, &e)
	return &e
}

// NewKeyEntry Makes a Entry on the Screen which holds a single glfw.Key value.
//
// Screen is the screen struct you will put it on.
//
// key is the key that will go in it.
//
// x, y, width (w), height (h), border, these should be self explanatory.
//
// command is the command you want it to run when you press enter put nil for none
//
// and it return a entry object
func NewKeyEntry(screen *Screen, key glfw.Key, x, y, w, h, border float64, command func()) *Entry {
	e := NewEntry(screen, "", x, y, w, h, border, command)
	e.KeyHolder = true
	e.Key = key
	e.Text = KeyName(key)
	return e
}
