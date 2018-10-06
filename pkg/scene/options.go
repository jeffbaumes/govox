package scene

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jeffbaumes/govox/pkg/common"
	"github.com/jeffbaumes/govox/pkg/gui"
)

// Option holds the information for a key binding option
type Option struct {
	Key   glfw.Key
	entry *gui.Entry
	label *gui.Label
}

func newOption(key glfw.Key) *Option {
	opt := Option{}
	opt.Key = key
	return &opt
}

// Options holds all the user-defined options for controlling the game
type Options struct {
	OptionMap map[string]*Option
}

func writefile(t string) {
	err := ioutil.WriteFile("options.txt", []byte(t), 0644)
	if err != nil {
		panic(err)
	}
}

func readfile() string {
	b, err := ioutil.ReadFile("options.txt")
	if err != nil {
		os.Create("options.txt")
		return readfile()
	}
	return string(b)
}

func load(o *Options) {
	f2 := strings.Split(readfile(), "\n")[0]
	s := strings.Split(f2, ";")
	for _, h := range s {
		g := strings.Split(h, "=")
		if len(g) < 2 || g[1] == "" {
			continue
		}
		n, err := strconv.Atoi(g[1])
		if err != nil {
			panic(err)
		}
		f := glfw.Key(n)
		o.OptionMap[g[0]].Key = f
	}
}

func save(o *Options) {
	strs := []string{}
	for name, opt := range o.OptionMap {
		if opt.entry != nil {
			opt.Key = opt.entry.Key
		}
		strs = append(strs, fmt.Sprintf("%v=%v", name, opt.Key))
	}
	writefile(strings.Join(strs, ";"))
}

// NewOptions creates a new options GUI
func NewOptions(screen *gui.Screen) *Options {
	o := Options{}
	o.OptionMap = make(map[string]*Option)
	m := o.OptionMap
	m["Apex"] = newOption(glfw.KeyK)
	m["Mode"] = newOption(glfw.KeyM)
	m["PlanetL"] = newOption(glfw.KeyLeftBracket)
	m["PlanetR"] = newOption(glfw.KeyRightBracket)
	m["Text"] = newOption(glfw.KeyEnter)
	m["Slot1"] = newOption(glfw.Key1)
	m["Slot2"] = newOption(glfw.Key2)
	m["Slot3"] = newOption(glfw.Key3)
	m["Slot4"] = newOption(glfw.Key4)
	m["Slot5"] = newOption(glfw.Key5)
	m["Slot6"] = newOption(glfw.Key6)
	m["Slot7"] = newOption(glfw.Key7)
	m["Slot8"] = newOption(glfw.Key8)
	m["Slot9"] = newOption(glfw.Key9)
	m["Slot10"] = newOption(glfw.Key0)
	m["Slot11"] = newOption(glfw.KeyMinus)
	m["Slot12"] = newOption(glfw.KeyEqual)
	m["Forward"] = newOption(glfw.KeyW)
	m["Backward"] = newOption(glfw.KeyS)
	m["Left"] = newOption(glfw.KeyA)
	m["Right"] = newOption(glfw.KeyD)
	m["Up"] = newOption(glfw.KeySpace)
	m["Down"] = newOption(glfw.KeyLeftShift)
	m["Run"] = newOption(glfw.KeyLeftShift)
	m["FlySprint"] = newOption(glfw.KeyLeftControl)
	m["Inventory"] = newOption(glfw.KeyE)
	m["Build"] = newOption(glfw.Key(1))
	m["Destroy"] = newOption(glfw.Key(0))
	load(&o)
	m["Forward"].entry = gui.NewKeyEntry(screen, m["Forward"].Key, -0.50, 0.75, 0.2, 0.15, 0.05, nil)
	m["Forward"].label = gui.NewLabel(screen, "Forward:", -0.95, 0.80, 0.09)
	m["Backward"].entry = gui.NewKeyEntry(screen, m["Backward"].Key, -0.50, 0.55, 0.2, 0.15, 0.05, nil)
	m["Backward"].label = gui.NewLabel(screen, "Backward:", -0.95, 0.60, 0.09)
	m["Left"].entry = gui.NewKeyEntry(screen, m["Left"].Key, -0.50, 0.35, 0.2, 0.15, 0.05, nil)
	m["Left"].label = gui.NewLabel(screen, "Left:", -0.95, 0.40, 0.09)
	m["Right"].entry = gui.NewKeyEntry(screen, m["Right"].Key, -0.50, 0.15, 0.2, 0.15, 0.05, nil)
	m["Right"].label = gui.NewLabel(screen, "Right:", -0.95, 0.20, 0.09)
	m["Up"].entry = gui.NewKeyEntry(screen, m["Up"].Key, 0, 0.75, 0.2, 0.15, 0.05, nil)
	m["Up"].label = gui.NewLabel(screen, "Up:", -0.25, 0.80, 0.09)
	m["Down"].entry = gui.NewKeyEntry(screen, m["Down"].Key, 0, 0.55, 0.2, 0.15, 0.05, nil)
	m["Down"].label = gui.NewLabel(screen, "Down:", -0.25, 0.60, 0.09)
	m["Run"].entry = gui.NewKeyEntry(screen, m["Run"].Key, 0.75, 0.75, 0.2, 0.15, 0.05, nil)
	m["Run"].label = gui.NewLabel(screen, "Run:", 0.25, 0.80, 0.09)
	m["FlySprint"].entry = gui.NewKeyEntry(screen, m["FlySprint"].Key, 0.75, 0.55, 0.2, 0.15, 0.05, nil)
	m["FlySprint"].label = gui.NewLabel(screen, "FlySprint:", 0.25, 0.60, 0.09)
	m["Inventory"].entry = gui.NewKeyEntry(screen, m["Inventory"].Key, 0.75, 0.35, 0.2, 0.15, 0.05, nil)
	m["Inventory"].label = gui.NewLabel(screen, "Inventory:", 0.25, 0.40, 0.09)
	m["Build"].entry = gui.NewKeyEntry(screen, m["Build"].Key, 0.75, -0.05, 0.2, 0.15, 0.05, nil)
	m["Build"].label = gui.NewLabel(screen, "Build:", 0.25, 0, 0.09)
	m["Destroy"].entry = gui.NewKeyEntry(screen, m["Destroy"].Key, 0.75, 0.15, 0.2, 0.15, 0.05, nil)
	m["Destroy"].label = gui.NewLabel(screen, "Destroy:", 0.25, 0.20, 0.09)
	// fmt.Printf("%+v", o)
	return &o
}

var oldop string

// Draw draws the options screen
func (o *Options) Draw(p *common.Player) {
	if oldop != p.Mode {
		oldop = p.Mode
		save(o)
	}
	hide := p.Mode != "Options"
	for _, opt := range o.OptionMap {
		if opt.entry != nil {
			opt.entry.Hide = hide
		}
		if opt.label != nil {
			opt.label.Hide = hide
		}
	}
}
