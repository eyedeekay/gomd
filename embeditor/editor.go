package embeditor

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"text/template"
	"time"

	rice "github.com/GeertJohan/go.rice"
	fcw "github.com/eyedeekay/go-fpw"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"github.com/nochso/gomd/eol"
)

func Runner(args InputArgs) {
	if _, err := os.Stat(*args.File); os.IsNotExist(err) {
		ioutil.WriteFile(*args.File, nil, 0644)
	}
	// Prepare (optionally) embedded resources
	templateBox := rice.MustFindBox("template")
	staticHTTPBox := rice.MustFindBox("static").HTTPBox()
	staticServer := http.StripPrefix("/static/", http.FileServer(staticHTTPBox))

	e := echo.New()

	t := &Template{
		templates: template.Must(template.New("base").Parse(templateBox.MustString("base.html"))),
	}
	e.SetRenderer(t)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/static/*", standard.WrapHandler(staticServer))

	edit := e.Group("/edit")
	edit.Get("/*", EditHandler)
	edit.Post("/*", EditHandlerPost)

	go WaitForServer(args)
	e.Run(standard.New(fmt.Sprintf("127.0.0.1:%d", *args.Port)))
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func EditHandler(c echo.Context) error {
	var ev *EditorView
	ev, ok := c.Get("editorView").(*EditorView)
	if !ok {
		log.Println("reading file")
		filepath := c.P(0)
		content, err := ioutil.ReadFile(filepath)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Unable to read requested file")
		}
		ev = NewEditorView(filepath, string(content))
		ev.CurrentLineEnding = eol.DetectDefault(ev.Content, eol.OSDefault())
		log.Println(ev.CurrentLineEnding.Description())
	}
	return c.Render(http.StatusOK, "base", ev)
}

func EditHandlerPost(c echo.Context) error {
	filepath := c.P(0)
	eolIndex, _ := strconv.Atoi(c.FormValue("eol"))
	content := c.FormValue("content")
	convertedContent, err := eol.LineEnding(eolIndex).Apply(content)
	if err != nil {
		convertedContent = content
		log.Println("Error while converting EOL. Saving without conversion.")
	}
	ioutil.WriteFile(filepath, []byte(convertedContent), 0644)
	c.Set("editorView", NewEditorView(filepath, content))
	return EditHandler(c)
}

func WaitForServer(args InputArgs) {
	log.Printf("Waiting for listener on port %d", *args.Port)
	url := fmt.Sprintf("http://localhost:%d/edit/%s", *args.Port, url.QueryEscape(*args.File))
	for {
		time.Sleep(time.Millisecond * 50)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		resp.Body.Close()
		break
	}
	log.Println("Opening " + url)
	if _, err := fcw.WebAppFirefox("gomd", false, url); err != nil {
		log.Println(err)
	} else {
		//defer ui.Close()
	}
}

type InputArgs struct {
	Port *int
	File *string
}

type EditorView struct {
	File              string
	Content           string
	LineEndings       map[int]string
	CurrentLineEnding eol.LineEnding
}

func NewEditorView(filepath string, content string) *EditorView {
	return &EditorView{
		File:        filepath,
		Content:     content,
		LineEndings: eol.Descriptions,
	}
}
