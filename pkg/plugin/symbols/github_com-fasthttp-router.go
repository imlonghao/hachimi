// Code generated by 'yaegi extract github.com/fasthttp/router'. DO NOT EDIT.

package symbols

import (
	"github.com/fasthttp/router"
	"go/constant"
	"go/token"
	"reflect"
)

func init() {
	Symbols["github.com/fasthttp/router/router"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"MatchedRoutePathParam": reflect.ValueOf(&router.MatchedRoutePathParam).Elem(),
		"MethodWild":            reflect.ValueOf(constant.MakeFromLiteral("\"*\"", token.STRING, 0)),
		"New":                   reflect.ValueOf(router.New),

		// type definitions
		"Group":  reflect.ValueOf((*router.Group)(nil)),
		"Router": reflect.ValueOf((*router.Router)(nil)),
	}
}
