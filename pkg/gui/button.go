package gui

import (
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Button is a GUI button with a callback
type Button struct {
	Text           string
	X, Y, W, H, C  float64
	Command        func()
	mouseover      bool
	program        uint32
	drawableVAO    uint32
	pointsVBO      uint32
	textureUniform int32
	texture        uint32
	textureUnit    int32
	colorlocation  int32
	Screen         *Screen
	index          int
	Hide           bool
}

const vertexShader = `
#version 410

in vec4 coord;
out vec2 texcoord;

void main(void) {
	gl_Position = vec4(coord.xy, -0.02, 1);
	texcoord = coord.zw;
}
`

const fragmentShader = `
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

func (b *Button) textSize(screen *Screen) (size float64) {
	wi, ht := FramebufferSize(screen.Window)
	y := b.H / 2
	x := b.W / float64(len(b.Text)) * 19 / 12 * float64(wi) / float64(ht)
	size = math.Min(x, y)
	return
}
func (b *Button) draw(screen *Screen) {
	if !b.Hide {
		gl.UseProgram(b.program)
		wi, ht := FramebufferSize(screen.Window)
		cx := b.C * float64(ht) / float64(wi)
		points := []float32{
			float32(b.X), float32(b.Y), 0, 0,
			float32(b.X + cx), float32(b.Y), 0.33, 0,
			float32(b.X), float32(b.Y + b.C), 0, 0.33,
			float32(b.X), float32(b.Y + b.C), 0, 0.33,
			float32(b.X + cx), float32(b.Y + b.C), 0.33, 0.33,
			float32(b.X + cx), float32(b.Y), 0.33, 0,

			float32(b.X + cx), float32(b.Y), 0.33, 0,
			float32(b.X + b.W - cx), float32(b.Y), 0.67, 0,
			float32(b.X + cx), float32(b.Y + b.C), 0.33, 0.33,
			float32(b.X + cx), float32(b.Y + b.C), 0.33, 0.33,
			float32(b.X + b.W - cx), float32(b.Y), 0.67, 0,
			float32(b.X + b.W - cx), float32(b.Y + b.C), 0.67, 0.33,

			float32(b.X + b.W - cx), float32(b.Y), 0.67, 0,
			float32(b.X + b.W), float32(b.Y), 1, 0,
			float32(b.X + b.W - cx), float32(b.Y + b.C), 0.67, 0.33,
			float32(b.X + b.W - cx), float32(b.Y + b.C), 0.67, 0.33,
			float32(b.X + b.W), float32(b.Y), 1, 0,
			float32(b.X + b.W), float32(b.Y + b.C), 1, 0.33,

			float32(b.X), float32(b.Y + b.C), 0, 0.33,
			float32(b.X), float32(b.Y + b.H - b.C), 0, 0.67,
			float32(b.X + cx), float32(b.Y + b.C), 0.33, 0.33,
			float32(b.X + cx), float32(b.Y + b.C), 0.33, 0.33,
			float32(b.X), float32(b.Y + b.H - b.C), 0, 0.67,
			float32(b.X + cx), float32(b.Y + b.H - b.C), 0.33, 0.67,

			float32(b.X + cx), float32(b.Y + b.C), 0.33, 0.33,
			float32(b.X + b.W - cx), float32(b.Y + b.C), 0.67, 0.33,
			float32(b.X + cx), float32(b.Y + b.H - b.C), 0.33, 0.67,
			float32(b.X + b.W - cx), float32(b.Y + b.C), 0.67, 0.33,
			float32(b.X + cx), float32(b.Y + b.H - b.C), 0.33, 0.67,
			float32(b.X + b.W - cx), float32(b.Y + b.H - b.C), 0.67, 0.67,

			float32(b.X + b.W - cx), float32(b.Y + b.C), 0.67, 0.33,
			float32(b.X + b.W - cx), float32(b.Y + b.H - b.C), 0.67, 0.67,
			float32(b.X + b.W), float32(b.Y + b.C), 1, 0.33,
			float32(b.X + b.W - cx), float32(b.Y + b.H - b.C), 0.67, 0.67,
			float32(b.X + b.W), float32(b.Y + b.C), 1, 0.33,
			float32(b.X + b.W), float32(b.Y + b.H - b.C), 1, 0.67,

			float32(b.X), float32(b.Y + b.H - b.C), 0, 0.67,
			float32(b.X), float32(b.Y + b.H), 0, 1,
			float32(b.X + cx), float32(b.Y + b.H - b.C), 0.33, 0.67,
			float32(b.X), float32(b.Y + b.H), 0, 1,
			float32(b.X + cx), float32(b.Y + b.H - b.C), 0.33, 0.67,
			float32(b.X + cx), float32(b.Y + b.H), 0.33, 1,

			float32(b.X + cx), float32(b.Y + b.H - b.C), 0.33, 0.67,
			float32(b.X + b.W - cx), float32(b.Y + b.H - b.C), 0.67, 0.67,
			float32(b.X + cx), float32(b.Y + b.H), 0.33, 1,
			float32(b.X + cx), float32(b.Y + b.H), 0.33, 1,
			float32(b.X + b.W - cx), float32(b.Y + b.H - b.C), 0.67, 0.67,
			float32(b.X + b.W - cx), float32(b.Y + b.H), 0.67, 1,

			float32(b.X + b.W - cx), float32(b.Y + b.H - b.C), 0.67, 0.67,
			float32(b.X + b.W), float32(b.Y + b.H - b.C), 1, 0.67,
			float32(b.X + b.W - cx), float32(b.Y + b.H), 0.67, 1,
			float32(b.X + b.W - cx), float32(b.Y + b.H), 0.67, 1,
			float32(b.X + b.W), float32(b.Y + b.H - b.C), 1, 0.67,
			float32(b.X + b.W), float32(b.Y + b.H), 1, 1,
		}
		fillVBO(b.pointsVBO, points)
		if b.mouseover {
			gl.Uniform4f(b.colorlocation, 0.9, 0.9, 0.9, 1.0)
		} else {
			gl.Uniform4f(b.colorlocation, 1.0, 1.0, 1.0, 1.0)
		}
		gl.Uniform1i(b.textureUniform, b.textureUnit)
		gl.BindVertexArray(b.drawableVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 6*9)
	}
}

// Remove removes the button from the screen
func (b *Button) Remove() {
	b.Screen.buttons[len(b.Screen.buttons)-1], b.Screen.buttons[b.index] = b.Screen.buttons[b.index], b.Screen.buttons[len(b.Screen.buttons)-1]
	b.Screen.buttons = b.Screen.buttons[:len(b.Screen.buttons)-1]
}

func (b *Button) isInside(x, y float64) bool {
	return x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H
}

// NewButton creates a new button
func NewButton(screen *Screen, text string, x, y, w, h, border float64, command func()) *Button {
	b := Button{}
	b.Text = text
	b.X = x
	b.Y = y
	b.W = w
	b.H = h
	b.C = border
	b.Command = command
	b.pointsVBO = newVBO()
	b.drawableVAO = newPointsVAO(b.pointsVBO, 4)
	b.program = createProgram(vertexShader, fragmentShader)
	b.colorlocation = uniformLocation(b.program, "color")
	b.textureUniform = uniformLocation(b.program, "texFont")
	b.Screen = screen
	bindAttribute(b.program, 0, "coord")
	existingImageFile, err := os.Open(screen.buttonpath)
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

	b.textureUnit = 5
	gl.ActiveTexture(uint32(gl.TEXTURE0 + b.textureUnit))
	gl.GenTextures(1, &b.texture)
	gl.BindTexture(gl.TEXTURE_2D, b.texture)

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
	b.index = len(screen.buttons) + 1
	screen.buttons = append(screen.buttons, &b)
	return &b
}
