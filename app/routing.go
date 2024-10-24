package app

import (
	"RestAPI/core"
	"fmt"
	"regexp"
	"strings"
)

type HandlerFunc func(core.HttpRequest) core.HttpResponse

type funcInfo struct {
	HandlerFunc
	name string
}

var HandlersList = make(map[*regexp.Regexp]funcInfo)

func registerHandler(url string, f HandlerFunc, name ...string) {
	var handlerName string
	if len(name) > 0 {
		handlerName = name[0]
	}
	pattern := regexp.MustCompile(`\{[a-zA-Z0-9:.!,?\-_]+\}`)
	matches := pattern.FindAllString(url, -1)
	if len(matches) > 0 {
		for _, match := range matches {
			if strings.HasPrefix(match, "{int:") {
				url = strings.Replace(url, match, `([0-9]+)`, 1)
			} else {
				url = strings.Replace(url, match, `([a-zA-Z0-9:.!,?\-_]+)`, 1)
			}
		}
	}

	regex := regexp.MustCompile("^" + url + "$")

	HandlersList[regex] = funcInfo{f, handlerName}
}

func router(url string) HandlerFunc {
	for pattern, info := range HandlersList {
		fmt.Println("Checking", pattern)
		if pattern.MatchString(url) {
			return info.HandlerFunc
		}
	}
	return nil
}
