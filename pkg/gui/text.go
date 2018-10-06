package gui

import (
	"encoding/json"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	vertexShaderSourceText = `
		#version 410

		in vec4 coord;
		out vec2 texcoord;

		void main(void) {
			gl_Position = vec4(coord.xy, -0.1, 1);
			texcoord = coord.zw;
		}
	`

	fragmentShaderSourceText = `
		#version 410

		in vec2 texcoord;
		uniform sampler2D texFont;
		out vec4 frag_color;

		void main(void) {
			vec4 texel = texture(texFont, texcoord);
			if (texel.a < 0.5) {
				discard;
		  }
			frag_color = texel;
		}
	`
)

// Text draws overlay text
type Text struct {
	charInfo       map[string]charInfo
	statusLine     textLine
	charCount      int
	drawableVAO    uint32
	pointsVBO      uint32
	numTriangles   int32
	program        uint32
	texture        uint32
	textureUnit    int32
	textureUniform int32
}

type textLine struct {
	str  string
	x, y int
}

type charInfo struct {
	x, y, width, height, originX, originY, advance int
}

// FramebufferSize returns the GLFW window's framebuffer size
func FramebufferSize(w *glfw.Window) (fbw, fbh int) {
	fbw, fbh = w.GetFramebufferSize()
	return
}

// NewText creates a new Text object
func NewText(screen *Screen) *Text {
	text := Text{}

	text.charInfo = make(map[string]charInfo)
	text.statusLine.x = 1
	text.statusLine.y = 1

	text.program = createProgram(vertexShaderSourceText, fragmentShaderSourceText)
	bindAttribute(text.program, 0, "coord")

	text.textureUniform = uniformLocation(text.program, "texFont")

	text.pointsVBO = newVBO()
	text.drawableVAO = newPointsVAO(text.pointsVBO, 4)

	// Load up texture info
	var textMeta map[string]interface{}
	textMetaBytes, e := ioutil.ReadFile(screen.fontjsonpath)
	if e != nil {
		panic(e)
	}
	json.Unmarshal(textMetaBytes, &textMeta)
	characters := textMeta["characters"].(map[string]interface{})
	for ch, props := range characters {
		propMap := props.(map[string]interface{})
		text.charInfo[ch] = charInfo{
			x:       int(propMap["x"].(float64)),
			y:       int(propMap["y"].(float64)),
			width:   int(propMap["width"].(float64)),
			height:  int(propMap["height"].(float64)),
			originX: int(propMap["originX"].(float64)),
			originY: int(propMap["originY"].(float64)),
			advance: int(propMap["advance"].(float64)),
		}
	}

	// Generated from https://evanw.github.io/font-texture-generator/
	// Inconsolata font (installed on system with Google Web Fonts), size 24
	// Power of 2, white with black stroke, thickness 2
	existingImageFile, err := os.Open(screen.fontpngpath)
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

	text.textureUnit = 0
	gl.ActiveTexture(uint32(gl.TEXTURE0 + text.textureUnit))
	gl.GenTextures(1, &text.texture)
	gl.BindTexture(gl.TEXTURE_2D, text.texture)

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
	return &text
}

func (text *Text) computeGeometry(width, height int, x, y, size float32, cx, cy bool) {
	const ht = 19
	const wd = 12
	sy := size
	sx := sy * float32(height) / float32(width)
	if cx {
		x -= float32(len(text.statusLine.str)) / 2 * sx * wd / ht
	}
	if cy {
		y += sy / 2
	}
	points := []float32{}
	text.charCount = 0
	line := text.statusLine
	for i, ch := range line.str + " " {
		aInfo := text.charInfo[string(ch)]
		ax1 := 1.0 / 512.0 * float32(aInfo.x-1)
		ax2 := 1.0 / 512.0 * float32(aInfo.x-1+aInfo.width)
		ay1 := 1.0 / 128.0 * float32(aInfo.y)
		ay2 := 1.0 / 128.0 * float32(aInfo.y+aInfo.height)
		x1 := x + float32(i)*sx*wd/ht - sx*float32(aInfo.originX)/ht
		x2 := x + float32(i)*sx*wd/ht + sx*float32(aInfo.width)/ht - sx*float32(aInfo.originX)/ht
		y1 := y + -sy*float32(aInfo.originY)/ht
		y2 := y + sy*float32(aInfo.height)/ht - sy*float32(aInfo.originY)/ht
		points = append(points, []float32{
			x1, -y1, ax1, ay1,
			x1, -y2, ax1, ay2,
			x2, -y2, ax2, ay2,

			x2, -y2, ax2, ay2,
			x2, -y1, ax2, ay1,
			x1, -y1, ax1, ay1,
		}...)
		text.charCount++
	}
	fillVBO(text.pointsVBO, points)
}

// Draw draws the overlay text
func (text *Text) draw(drawtext string, x, y, size float64, cx, cy bool, w *glfw.Window) {
	wi, h := FramebufferSize(w)
	text.statusLine.str = drawtext
	text.computeGeometry(wi, h, float32(x), float32(-y), float32(size), cx, cy)
	gl.UseProgram(text.program)
	gl.Uniform1i(text.textureUniform, text.textureUnit)
	gl.BindVertexArray(text.drawableVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6*int32(text.charCount))
}
