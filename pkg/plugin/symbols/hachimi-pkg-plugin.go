// Code generated by 'yaegi extract hachimi/pkg/plugin'. DO NOT EDIT.

//go:build go1.22
// +build go1.22

package symbols

import (
	"hachimi/pkg/plugin"
	"reflect"
)

func init() {
	Symbols["hachimi/pkg/plugin/plugin"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"RequestHandler": reflect.ValueOf(plugin.RequestHandler),
		"ServeIndex":     reflect.ValueOf(&plugin.ServeIndex).Elem(),
		"ServeList":      reflect.ValueOf(&plugin.ServeList).Elem(),
		"TitleIndex":     reflect.ValueOf(&plugin.TitleIndex).Elem(),
		"TitleList":      reflect.ValueOf(&plugin.TitleList).Elem(),
	}
}
