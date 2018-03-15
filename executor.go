package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"text/template"

	"github.com/mgutz/str"
	"github.com/sirupsen/logrus"
)

/*

executor := NewExecutor()
for i := range item {
	executor.Exec(name, command, labels)
}
executor.Consume(option);

*/
type Executor interface {
	Exec(name string, command string, labels map[string]string)
	Consume(opt *ConsumeOption) error
}
type ConsumeOption struct {
	ShowName bool
}

func NewExecutor(runner string) Executor {
	logrus.Infof("runner '%s'", runner)
	tmpl, err := template.New("__").Parse(runner)
	if err != nil {
		logrus.Error(err)
	}
	return &simple{
		out:  make(chan Line, 32),
		wg:   new(sync.WaitGroup),
		tmpl: tmpl,
	}
}

type simple struct {
	out        chan Line
	wg         *sync.WaitGroup
	tmpl       *template.Template
	maxNameLen int
}

func (se *simple) Exec(name string, command string, labels map[string]string) {
	// async running command
	se.wg.Add(1)

	if se.maxNameLen < len(name) {
		se.maxNameLen = len(name)
	}

	var cmd bytes.Buffer

	labels["name"] = name
	labels["command"] = command

	err := se.tmpl.Execute(&cmd, labels)
	if err != nil {
		logrus.Error(err)
	}

	logrus.Infof("running command: %s to '%s'", name, cmd.String())
	go func(out chan<- Line, cmd string) {
		defer se.wg.Done()
		parts := str.ToArgv(cmd)
		logrus.Debugf("%+v", parts)

		c := exec.Command(parts[0], parts[1:]...)
		var buf = &bytes.Buffer{}

		r := bufio.NewReader(buf)
		c.Stdout = buf
		c.Stderr = os.Stderr
		err := c.Run()
		if err != nil {
			logrus.Warn(err)
		}

		for {
			line, _, err := r.ReadLine()
			if err == io.EOF {
				logrus.Debugf("exec end.")
				return
			}
			if err != nil {
				logrus.Error(err)
				return
			}
			out <- Line{name: name, text: string(line)}
		}
	}(se.out, cmd.String())
}
func (se *simple) Consume(opt *ConsumeOption) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		logrus.Debugf("ready to consume")
		for line := range se.out {
			if opt.ShowName {
				fmt.Printf("%s | %s\n", str.PadLeft(line.name, " ", se.maxNameLen), line.text)
			} else {
				fmt.Printf("%s\n", line.text)
			}
		}
	}()
	se.wg.Wait()
	close(se.out)
	<-done
	return nil
}
