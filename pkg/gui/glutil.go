package gui

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func createProgram(vertexSource, fragmentSource string) uint32 {
	vertexShader, err := compileShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)
	return program
}

func newVBO() uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	return vbo
}

func fillVBO(vbo uint32, data []float32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(data), gl.Ptr(data), gl.STATIC_DRAW)
}

func newPointsNormalsTcoordsVAO(pointsVBO, normalsVBO, tcoordsVBO uint32) uint32 {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, pointsVBO)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, normalsVBO)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(2)
	gl.BindBuffer(gl.ARRAY_BUFFER, tcoordsVBO)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 0, nil)
	return vao
}

func newPointsNormalsVAO(pointsVBO, normalsVBO uint32) uint32 {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, pointsVBO)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, normalsVBO)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil)
	return vao
}

func newPointsVAO(pointsVBO uint32, size int32) uint32 {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, pointsVBO)
	gl.VertexAttribPointer(0, size, gl.FLOAT, false, 0, nil)
	return vao
}

func bindAttribute(prog, location uint32, name string) {
	s, free := gl.Strs(name + "\x00")
	gl.BindAttribLocation(prog, location, *s)
	free()
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

// uniformLocation retrieves a uniform location by name
func uniformLocation(program uint32, name string) int32 {
	glstr, free := gl.Strs(name + "\x00")
	uniform := gl.GetUniformLocation(program, *glstr)
	free()
	return uniform
}

// InitGlfw initializes the GLFW window
func InitGlfw(windowsizex, windowsizey int, windowname string) *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(windowsizex, windowsizey, windowname, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// InitOpenGL initializes the OpenGL context
func InitOpenGL() {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	gl.Enable(gl.DEPTH_TEST)
	// gl.Enable(gl.POLYGON_OFFSET_FILL)
	// gl.PolygonOffset(2, 0)

	gl.LineWidth(5)
	gl.Enable(gl.LINE_SMOOTH)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func max(num1, num2 int) int {
	if num1 > num2 {
		return num1
	}
	return num2
}
