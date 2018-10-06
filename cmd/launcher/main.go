package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jeffbaumes/govox/pkg/client"
	"github.com/jeffbaumes/govox/pkg/gui"
	"github.com/jeffbaumes/govox/pkg/server"
)

type profile struct {
	name, world, host string
	port              int
	normal            bool
}
type ui struct {
	profile int
}

func windowSizeCallback(w *glfw.Window, wd, ht int) {
	fwidth, fheight := gui.FramebufferSize(w)
	gl.Viewport(0, 0, int32(fwidth), int32(fheight))
}

func world() {
	println("world")

}
func getprofiles() []*profile {
	f := readfile()
	profiles := strings.Split(f, "\n")
	prof := []*profile{}

	var err error
	for _, p := range profiles {
		if p != "" {
			pr := profile{}
			s := strings.Split(p, ";")
			for _, h := range s {
				g := strings.Split(h, "=")
				if g[0] == "world" {
					pr.world = g[1]
				} else if g[0] == "port" {
					pr.port, err = strconv.Atoi(g[1])
					if err != nil {
						panic(err)
					}
				} else if g[0] == "host" {
					pr.host = g[1]
				} else if g[0] == "name" {
					pr.name = g[1]
				} else if g[0] == "normal" {
					n, err := strconv.Atoi(g[1])
					if err != nil {
						panic(err)
					}
					if n == 1 {
						pr.normal = true
					} else {
						pr.normal = false
					}
				}
			}
			fmt.Println(pr)
			prof = append(prof, &pr)
		}
	}
	return prof
}
func main() {
	profiles := getprofiles()
	ui := ui{}
	runtime.LockOSThread()
	w := gui.InitGlfw(700, 500, "World Blocks")
	gui.InitOpenGL()
	w.SetSizeCallback(windowSizeCallback)
	screen := gui.NewScreen(w)
	screen.InitGui("textures/font/font.png", "textures/font/font.json", "textures/button.png", "textures/entry.png", 0.3)
	guimousebuttoncallback := screen.MouseButtonCallback()
	guicursorposcallback := screen.CursorPosCallback()
	guikeycallback := screen.KeyCallBack()
	w.SetMouseButtonCallback(guimousebuttoncallback)
	w.SetCursorPosCallback(guicursorposcallback)
	w.SetKeyCallback(guikeycallback)
	play := gui.NewButton(screen, "PLAY", -0.85, 0.6, 0.85, 0.3, 0.1, nil)
	serverb := gui.NewButton(screen, "PLAY ON SERVER", 0.05, 0.6, 0.85, 0.3, 0.1, nil)
	namee := gui.NewEntry(screen, "", -0.55, 0.2, 0.5, 0.2, 0.05, nil)
	gui.NewLabel(screen, "Name", -0.41, 0.45, 0.1)
	worlde := gui.NewEntry(screen, "", 0.05, 0.2, 0.5, 0.2, 0.05, nil)
	gui.NewLabel(screen, "World", 0.18, 0.45, 0.1)
	porte := gui.NewEntry(screen, "", 0.05, -0.17, 0.5, 0.2, 0.05, nil)
	gui.NewLabel(screen, "Port", 0.2, 0.08, 0.1)
	defaulte := gui.NewButton(screen, "", -0.25, -0.4, 0.5, 0.2, 0.05, nil)
	hoste := gui.NewEntry(screen, "", -0.55, -0.17, 0.5, 0.2, 0.05, nil)
	gui.NewLabel(screen, "Host", -0.41, 0.08, 0.1)
	pickpro := gui.NewButton(screen, "Switch profile", -0.95, -0.95, 0.4, 0.2, 0.05, nil)
	message := gui.NewLabel(screen, "", -0.5, -0.95, 0.1)

	loadProfile := func() {
		namee.Text = profiles[ui.profile].name
		worlde.Text = profiles[ui.profile].world
		porte.Text = fmt.Sprintf("%v", profiles[ui.profile].port)
		defaulte.Text = fmt.Sprintf("default = %v", profiles[ui.profile].normal)
		hoste.Text = profiles[ui.profile].host
	}

	setn := func() {
		profiles[ui.profile].normal = !profiles[ui.profile].normal
		if profiles[ui.profile].normal {
			for i, p := range profiles {
				if i == ui.profile {
					continue
				}
				p.normal = false
			}
		}
		loadProfile()
	}
	defaulte.Command = setn

	var err error
	saveProfile := func() {
		profiles[ui.profile].name = namee.Text
		profiles[ui.profile].world = worlde.Text
		profiles[ui.profile].port, err = strconv.Atoi(porte.Text)
		if err != nil {
			profiles[ui.profile].port = 0
		}
		profiles[ui.profile].host = hoste.Text
	}
	connecttoserver := func() {
		client.Start(profiles[ui.profile].name, "friends123.tk", 1234, screen)

	}
	serverb.Command = connecttoserver
	newp := func() {
		message.Text = "Successfully created profile."
		profiles = append(profiles, &profile{})
		ui.profile = len(profiles) - 1
		loadProfile()
	}
	ui.profile = -1
	for i, p := range profiles {
		if p.normal == true {
			ui.profile = i
		}
	}
	if ui.profile == -1 {
		if len(profiles) > 0 {
			ui.profile = 0
		} else {
			newp()
		}
	}
	gui.NewButton(screen, "new profile", -0.95, -0.7, 0.4, 0.2, 0.05, newp)

	removep := func() {
		if len(profiles) > 1 {
			message.Text = "Successfully deleted profile."
			profiles = append(profiles[:ui.profile], profiles[ui.profile+1:]...)
			ui.profile = ui.profile % len(profiles)
			loadProfile()
		} else {
			message.Text = "Cannot delete the only profile."
		}
	}
	gui.NewButton(screen, "Delete profile", -0.5, -0.7, 0.4, 0.2, 0.05, removep)

	pickp := func() {
		saveProfile()
		ui.profile = (ui.profile + 1) % len(profiles)
		loadProfile()
	}
	pickpro.Command = pickp

	saveProfileFile := func() {
		s := []string{}
		for _, p := range profiles {
			ps := []string{}
			ps = append(ps, fmt.Sprintf("name=%v", p.name))
			ps = append(ps, fmt.Sprintf("world=%v", p.world))
			ps = append(ps, fmt.Sprintf("host=%v", p.host))
			normalInt := 0
			if p.normal {
				normalInt = 1
			}
			ps = append(ps, fmt.Sprintf("normal=%v", normalInt))
			ps = append(ps, fmt.Sprintf("port=%v", p.port))
			s = append(s, strings.Join(ps, ";"))
		}
		writefile(strings.Join(s, "\n"))
	}

	run := func() {
		if namee.Text == "" {
			message.Text = "ERROR: Name needs to be something"
			namee.Focus = true
		} else {
			saveProfile()
			saveProfileFile()
			if profiles[ui.profile].world == "" {
				client.Start(profiles[ui.profile].name, profiles[ui.profile].host, profiles[ui.profile].port, screen)
			} else if profiles[ui.profile].world != "" {
				go server.Start(profiles[ui.profile].world, 123, profiles[ui.profile].port, "sun-moon")
				time.Sleep(time.Second)
				client.Start(profiles[ui.profile].name, profiles[ui.profile].host, profiles[ui.profile].port, screen)
			}
		}
	}
	play.Command = run

	loadProfile()
	saveProfile()

	for !w.ShouldClose() {
		gl.ClearColor(0.498, 1.000, 0.831, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		screen.Update()
		glfw.PollEvents()
		w.SwapBuffers()
	}
}

func writefile(t string) {
	err := ioutil.WriteFile("profiles.txt", []byte(t), 0644)
	if err != nil {
		panic(err)
	}
}

func readfile() string {
	b, err := ioutil.ReadFile("profiles.txt")
	if err != nil {
		os.Create("profiles.txt")
		return readfile()
	}
	return string(b)
}
