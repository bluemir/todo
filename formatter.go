package main

import (
	"encoding/json"
	"fmt"

	"github.com/mgutz/str"
)

type Formatter interface {
	Out(line uint, name, from, text string) error
}

func NewFormatter(format string, items []Item) Formatter {
	max := 0
	for _, item := range items {
		if max < len(item["name"]) {
			max = len(item["name"])
		}
	}
	switch format {
	case "text":
		return &TextFormatter{}
	case "simple":
		return &SimpleFormatter{max}
	case "detail":
		return &DetailFormatter{max}
	case "json":
		return &JsonFormatter{}
	default:
		return &DefaultFormatter{format, max}
	}
}

type TextFormatter struct {
}

func (f *TextFormatter) Out(line uint, name, from, text string) error {
	fmt.Printf("%s\n", text)
	return nil
}

type SimpleFormatter struct {
	max int
}

func (f *SimpleFormatter) Out(line uint, name, from, text string) error {
	fmt.Printf("%s | %s\n",
		str.PadLeft(name, " ", f.max),
		text,
	)
	return nil
}

type DetailFormatter struct {
	max int
}

func (f *DetailFormatter) Out(line uint, name, from, text string) error {
	fmt.Printf("%s %05d %s | %s\n",
		str.PadLeft(name, " ", f.max),
		line,
		from,
		text,
	)
	return nil
}

type JsonFormatter struct {
}

func (f *JsonFormatter) Out(line uint, name, from, text string) error {
	buf, _ := json.Marshal(struct {
		Line uint
		Name string
		From string
		Text string
	}{line, name, from, text})
	fmt.Printf("%s\n", buf)
	return nil
}

type DefaultFormatter struct {
	format string
	max    int
}

func (f *DefaultFormatter) Out(line uint, name, from, text string) error {
	for _, c := range f.format {
		switch c {
		case 'n':
			fmt.Printf(str.PadLeft(name, " ", f.max))
			fmt.Printf(" ")
		case 'i':
			fmt.Printf("%05d", line)
			fmt.Printf(" ")
		case 'f':
			fmt.Printf("%s", from)
			fmt.Printf(" ")
		case 't':
			fmt.Printf("%s", text)
			fmt.Printf(" ")
		default:
			fmt.Printf("%c", c)
		}
	}
	fmt.Printf("\n")
	return nil
}
