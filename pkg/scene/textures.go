package scene

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/govox/pkg/common"
)

// LoadTextures loads textures from the textures directory into a single texture image
func LoadTextures() *image.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for x := 0; x < len(common.Materials); x++ {
		ImageFile, err := os.Open(fmt.Sprintf("textures/%s.png", common.Materials[x]))
		if err != nil {
			panic(err)
		}
		img, err := png.Decode(ImageFile)
		if err != nil {
			panic(err)
		}
		sx := (x % 4) * 16
		sy := (x / 4) * 16
		r, g, b, _ := img.At(8, 8).RGBA()
		common.MaterialColors[x] = mgl32.Vec3{float32(r) / 0xffff, float32(g) / 0xffff, float32(b) / 0xffff}
		draw.Draw(rgba, image.Rect(sx, sy, sx+16, sy+16), img, image.Pt(0, 0), draw.Src)
		ImageFile.Close()
	}
	return rgba
}

// LoadImage loads an image
func LoadImage(path string) *image.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, 16, 16))
	ImageFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	img, err := png.Decode(ImageFile)
	if err != nil {
		panic(err)
	}
	draw.Draw(rgba, image.Rect(0, 0, 16, 16), img, image.Pt(0, 0), draw.Src)
	ImageFile.Close()
	return rgba
}

// LoadImages load images into a larger texture
func LoadImages(paths []string, size int) (rgba *image.RGBA, columns int) {
	columns = int(math.Ceil(math.Sqrt(float64(len(paths)))))
	rgba = image.NewRGBA(image.Rect(0, 0, columns*size, columns*size))
	for x := 0; x < len(paths); x++ {
		ImageFile, err := os.Open(paths[x])
		if err != nil {
			panic(err)
		}
		img, err := png.Decode(ImageFile)
		if err != nil {
			panic(err)
		}
		sx := (x % columns) * size
		sy := (x / columns) * size
		draw.Draw(rgba, image.Rect(sx, sy, sx+size, sy+size), img, image.Pt(0, 0), draw.Src)
		ImageFile.Close()
	}
	return
}
