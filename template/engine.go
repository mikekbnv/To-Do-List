package template

import (
	// _ "embed" // go@1.16+
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/labstack/echo"
	"github.com/yosssi/ace"
)

// Engine ...
type Engine struct {
	Options *ace.Options
	mutex   sync.RWMutex
}

// NewEngine ...
func NewEngine(directory string) *Engine {
	return &Engine{Options: &ace.Options{
		Asset:   getAsset,
		BaseDir: directory,
	}}
}

// DynamicReload ...
func (e *Engine) DynamicReload(enabled bool) *Engine {
	e.Options.DynamicReload = enabled
	return e
}

// AddFunc ...
func (e *Engine) AddFunc(name string, fn interface{}) *Engine {
	e.mutex.Lock()
	e.Options.FuncMap[name] = fn
	e.mutex.Unlock()
	return e
}

func getAsset(name string) ([]byte, error) {
	read := ioutil.ReadFile //embed.FS.ReadFile

	if data, err := read(name); err == nil {
		return data, nil
	}

	return nil, fmt.Errorf("Asset %s not found", name)
}

// Render ...
func (e *Engine) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tpl, err := ace.Load(name, "", e.Options)

	if err != nil {
		return err
	}

	return tpl.Execute(w, data)
}
