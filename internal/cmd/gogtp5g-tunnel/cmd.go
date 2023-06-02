package main

import (
	"strings"
)

type CmdNode interface {
	Find([]string) (CmdFunc, []string)
}

type CmdNodeList []CmdNode

func (nl CmdNodeList) Find(args []string) (CmdFunc, []string) {
	for _, n := range nl {
		f, params := n.Find(args)
		if f != nil {
			return f, params
		}
	}
	return nil, args
}

type CmdToken struct {
	Name string
	Desc string
	Next CmdNode
}

func (t CmdToken) Find(args []string) (CmdFunc, []string) {
	if len(args) < 1 {
		return nil, args
	}
	if !strings.HasPrefix(t.Name, args[0]) {
		return nil, args
	}
	if t.Next == nil {
		return nil, args
	}
	return t.Next.Find(args[1:])
}

type CmdFunc func([]string) error

func (f CmdFunc) Find(args []string) (CmdFunc, []string) {
	return f, args
}

type CmdParser struct {
	args []string
	pos  int
}

func NewCmdParser(args []string) *CmdParser {
	c := new(CmdParser)
	c.args = args
	c.pos = 0
	return c
}

func (c *CmdParser) GetToken() (string, bool) {
	if c.pos >= len(c.args) {
		return "", false
	}
	tok := c.args[c.pos]
	c.pos++
	return tok, true
}
