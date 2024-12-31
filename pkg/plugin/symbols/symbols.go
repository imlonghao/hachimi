package symbols

import "reflect"

//go:generate go install github.com/traefik/yaegi/cmd/yaegi@latest
//go:generate yaegi extract github.com/valyala/fasthttp
//go:generate yaegi extract github.com/fasthttp/router
//go:generate yaegi extract hachimi/pkg/plugin
//go:generate yaegi extract hachimi/pkg/types

var Symbols = map[string]map[string]reflect.Value{}
