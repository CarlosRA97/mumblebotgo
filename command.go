package main

import "layeh.com/gumble/gumble"

type commandFuncType func(*gumble.TextMessageEvent)

type Commands struct {
	commandKeys []string
	commandMap map[string]commandFuncType
}

func NewCommands() *Commands {
	return &Commands{commandMap: make(map[string]commandFuncType)}
}

func (c *Commands) add(commandType string, commandFunc commandFuncType) {
	c.commandKeys = append(c.commandKeys, commandType)
	c.commandMap[commandType] = commandFunc	
}

func (c *Commands) execute(commandType string, e *gumble.TextMessageEvent) {
	c.commandMap[commandType](e)
}

func (c *Commands) getAll() []string {
	return c.commandKeys
}