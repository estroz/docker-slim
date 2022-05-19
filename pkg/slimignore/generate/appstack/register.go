package appstack

import "fmt"

type AppStackFunc func() AppStack

var allAppStacks = map[string]AppStackFunc{}

func Register(sf AppStackFunc) {
	name := sf().Name()
	if _, added := allAppStacks[name]; added {
		panic(fmt.Sprintf("app stack %s already added", name))
	}
	allAppStacks[name] = sf
}

func GetAll() map[string]AppStackFunc {
	return allAppStacks
}
