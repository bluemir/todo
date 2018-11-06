package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"text/template"

	"github.com/mgutz/str"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type textCollector struct {
	formatter Formatter
}

func (c *textCollector) Add(name string, cmd *exec.Cmd) (<-chan struct{}, error) {
	done := make(chan struct{})
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logrus.Warn("pipe stdout", err)
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logrus.Warn("pipe stderr", err)
		return nil, err
	}

	go func() {
		defer close(done)

		eg, _ := errgroup.WithContext(context.Background())
		eg.Go(func() error { return c.read(name, "stdout", stdout) })
		eg.Go(func() error { return c.read(name, "stderr", stderr) })
		if err := eg.Wait(); err != nil {
			logrus.Error(err)
		}
	}()
	return done, nil
}
func (c *textCollector) read(name, from string, reader io.Reader) error {
	ln := uint(1)
	r := bufio.NewScanner(reader)
	for r.Scan() {
		logrus.Debugf("read line from %s %s", name, from)
		c.formatter.Out(ln, name, from, r.Text())
		ln++
	}
	if err := r.Err(); err != nil {
		logrus.Errorf("read line error %s %s: %q", name, from, err)
		return err
	}
	logrus.Debugf("end of stream")
	return nil
}

type Runner struct {
	tmpl    string
	command string
}

func (r *Runner) Run(out *textCollector, items ...Item) error {
	tmpl, err := template.New("__").Funcs(map[string]interface{}{
		"command": func() string {
			return r.command
		},
	}).Parse(r.tmpl)
	if err != nil {
		logrus.Error(err)
	}

	eg, _ := errgroup.WithContext(context.Background())
	for _, item := range items {
		eg.Go(rc(out, tmpl, item))
	}
	return eg.Wait()
}
func rc(out *textCollector, tmpl *template.Template, item Item) func() error {
	return func() error {
		var cmd bytes.Buffer

		err := tmpl.Execute(&cmd, item)
		if err != nil {
			return err
		}

		parts := str.ToArgv(cmd.String())
		logrus.Debugf("%+v", parts)

		c := exec.Command(parts[0], parts[1:]...)

		done, err := out.Add(item["name"], c)
		if err != nil {
			return err
		}

		if err := c.Start(); err != nil {
			return err
		}

		<-done

		return c.Wait()
	}
}
