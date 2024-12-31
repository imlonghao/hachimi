package rules

import (
	"github.com/expr-lang/expr/vm"
	"sync"
)

type Rule struct {
	Name    string   `toml:"name"`
	Rule    string   `toml:"rule"`
	Tags    []string `toml:"tags"`
	Version int      `toml:"version"`
	program *vm.Program
}

var rulesMutex sync.RWMutex
var rules []Rule
