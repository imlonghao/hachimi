package rules

import (
	"fmt"
	"github.com/expr-lang/expr"
	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml"
	"log"
	"os"
	"path/filepath"
)

func EvaluateRule(rule *Rule, data map[string]interface{}) (*[]string, error) {
	output, err := expr.Run(rule.program, map[string]any{
		"data":          data,
		"unquote":       unquote,
		"indexOfH":      indexOfH,
		"lastIndexOfH":  lastIndexOfH,
		"hasPrefixH":    hasPrefixH,
		"hasSuffixH":    hasSuffixH,
		"compareBytesH": compareBytesH,
	})
	if err != nil {
		return nil, err
	}
	if output.(bool) {
		return &rule.Tags, nil
	}
	return nil, nil

}

func loadRulesFromFolder(folderPath string) {
	filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error walking through the folder:", err)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".toml" {
			// 解析Toml文件
			rulesFromFile, err := parseTomlFile(path)
			if err != nil {
				fmt.Println("Error parsing Toml file:", err)
			} else {
				// 合并规则到数组
				updateRules(rulesFromFile)
			}
		}
		return nil
	})
}

func parseTomlFile(filePath string) ([]Rule, error) {
	config, err := toml.LoadFile(filePath)
	if err != nil {
		return nil, err
	}
	var rulesFromFile []Rule
	r := config.Get("rule")
	if r != nil {
		ruleArray := r.([]*toml.Tree)
		for _, ruleTree := range ruleArray {
			var rule Rule
			if err := ruleTree.Unmarshal(&rule); err != nil {
				return nil, err
			}
			rulesFromFile = append(rulesFromFile, rule)
		}
	}

	return rulesFromFile, nil
}
func updateRules(newRules []Rule) {
	rulesMutex.Lock()
	defer rulesMutex.Unlock()

	for _, newRule := range newRules {
		found := false
		for i, existingRule := range rules {
			if existingRule.Name == newRule.Name {
				if existingRule.Rule != newRule.Rule {

					program, err := expr.Compile(newRule.Rule)
					if err != nil {
						log.Println(newRule.Name, err)
						found = true
						break
					}
					newRule.program = program
					rules[i] = newRule

				}
				found = true
				break
			}
		}

		if !found {
			program, err := expr.Compile(newRule.Rule)
			if err != nil {
				log.Println(newRule.Name, err)
				continue
			}
			newRule.program = program
			rules = append(rules, newRule)

		}
	}
}

func watchConfigFolder(folderPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}
	defer watcher.Close()

	// 添加文件夹监控
	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error adding folder to watcher:", err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// 文件写入或创建事件，重新加载规则
				rulesFromFile, err := parseTomlFile(event.Name)
				if err != nil {
					fmt.Println("Error parsing Toml file:", err)
				} else {
					// 合并规则到数组
					updateRules(rulesFromFile)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Error in file watcher:", err)
		}
	}
}
