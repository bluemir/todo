package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	DisplayFormat string
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

		stdout, err := c.StdoutPipe()
		if err != nil {
			logrus.Warn("pipe stdout", err)
		}
		stderr, err := c.StderrPipe()
		if err != nil {
			logrus.Warn("pipe stderr", err)
		}

		//se.wg.Add(2)

		wg := &sync.WaitGroup{}
		wg.Add(2)
		go pipe(out, stdout, "stdout", name, wg)
		go pipe(out, stderr, "stderr", name, wg)

		err = c.Start()
		if err != nil {
			logrus.Warn(err)
		}

		// read line -> out
		wg.Wait()

		err = c.Wait()
		if err != nil {
			logrus.Warn(err)
		}
		logrus.Debugf("Exec done: %s", cmd)
	}(se.out, cmd.String())
}
func pipe(out chan<- Line, reader io.Reader, from string, name string, wg *sync.WaitGroup) {
	defer wg.Done()
	ln := uint(1)
	r := bufio.NewScanner(reader)
	for r.Scan() {
		logrus.Debugf("read line from %s %s", name, from)
		out <- Line{
			Name: name,
			Text: r.Text(),
			From: from,
			Num:  ln,
		}
		ln++
	}
	if err := r.Err(); err != nil {
		logrus.Errorf("read line error %s %s: %q", name, from, err)
		return
	}
	logrus.Debugf("end of stream")
	// TODO call c.wait
}

func (se *simple) Consume(opt *ConsumeOption) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		logrus.Debugf("ready to consume")
		for line := range se.out {
			logrus.Debugf("get line")
			switch opt.DisplayFormat {
			case "json":
				buf, _ := json.Marshal(line)
				fmt.Printf("%s\n", buf)
			case "text":
				fmt.Printf("%s\n", line.Text)
			case "simple":
				fmt.Printf("%s %05d | %s\n",
					str.PadLeft(line.Name, " ", se.maxNameLen),
					line.Num,
					line.Text,
				)
			case "detail":
				fmt.Printf("%s %05d %s | %s\n",
					str.PadLeft(line.Name, " ", se.maxNameLen),
					line.Num,
					line.From,
					line.Text,
				)
			default:
				for _, c := range opt.DisplayFormat {
					switch c {
					case 'n':
						fmt.Printf(str.PadLeft(line.Name, " ", se.maxNameLen))
					case 'i':
						fmt.Printf("%05d", line.Num)
					case 'f':
						fmt.Printf("%s", line.From)
					case 't':
						fmt.Printf("%s", line.Text)
					default:
						fmt.Printf("%c", c)
					}
					fmt.Printf(" ")
				}
				fmt.Printf("\n")
			}
		}
	}()
	se.wg.Wait()
	// done
	// but there is Lines to read
	// MUST close end of read...
	close(se.out)
	<-done
	return nil
}
