package scene

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/govox/pkg/common"
	pb "github.com/jeffbaumes/govox/pkg/govox"
	"github.com/jeffbaumes/govox/pkg/gui"
)

// Text holds the text information
type Text struct {
}

var o int
var tex1 *gui.Label
var texte *gui.Entry
var textl [5]*gui.Label
var oldtext [4]string
var h bool

// Draw draws the overlay text
func (text *Text) Draw(player *common.Player, screen *gui.Screen, u *Universe) {
	r, theta, phi := mgl32.CartesianToSpherical(player.Location())
	if o == 0 {
		tex1 = gui.NewLabel(screen, "", -0.95, -0.95, 0.05)
		for x := range textl {
			textl[x] = gui.NewLabel(screen, "", -0.95, float64(0.85-float64(x)*0.10), 0.08-float64(x)*0.008)
		}

		texte = gui.NewEntry(screen, "", -0.75, -0.85, 1.5, 0.2, 0.04, func() {
			player.Mode = "Play"
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err := u.GRPCClient.SendText(ctx, &pb.SendTextRequest{Text: fmt.Sprintf("%v: %v", player.Name, texte.Text)})
			if err != nil {
				log.Fatalf("sending text failed: %v", err)
			}
			texte.Text = ""
		})
		o = 1
	}
	tex1.Text = fmt.Sprintf("LAT %v, LON %v, ALT %v", int(theta/math.Pi*180-90+0.5), int(phi/math.Pi*180+0.5), int(r+0.5))
	if player.Mode == "Text" {
		texte.Y = -0.85
		texte.Focus = true
		screen.Update()
	} else {
		texte.Y = 10
		texte.Focus = false
	}
	if player.DrawText != oldtext[0] {
		textl[0].Text = player.DrawText
		textl[1].Text = oldtext[0]
		textl[2].Text = oldtext[1]
		textl[3].Text = oldtext[2]
		textl[4].Text = oldtext[3]
		oldtext[3] = oldtext[2]
		oldtext[2] = oldtext[1]
		oldtext[1] = oldtext[0]
		oldtext[0] = player.DrawText
	}
}
