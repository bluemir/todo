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

type Runner struct {
	tmpl    string
	command string
	dryRun  bool
}

func (r *Runner) Run(formatter Formatter, items ...Item) error {
	tmpl, err := template.New("__").Funcs(map[string]interface{}{
		"command": func() string {
			return r.command
		},
	}).Parse(r.tmpl)
	if err != nil {
		logrus.Error(err)
	}

	eg, ctx := errgroup.WithContext(context.Background())
	for _, item := range items {
		var cmd bytes.Buffer

		err := tmpl.Execute(&cmd, item)
		if err != nil {
			return err
		}

		parts := str.ToArgv(cmd.String())
		logrus.Debugf("cmd: %+v", cmd.String())
		logrus.Debugf("part: %+v", parts)

		if r.dryRun {
			parts = append([]string{"echo"}, parts...)
		}

		c := exec.CommandContext(ctx, parts[0], parts[1:]...)
		defer c.Wait()

		stdout, err := c.StdoutPipe()
		if err != nil {
			logrus.Warn("pipe stdout", err)
			return err
		}
		stderr, err := c.StderrPipe()
		if err != nil {
			logrus.Warn("pipe stderr", err)
			return err
		}

		eg.Go(func() error { return read(item["name"], "stdout", stdout, formatter) })
		eg.Go(func() error { return read(item["name"], "stderr", stderr, formatter) })

		eg.Go(c.Start)
	}
	return eg.Wait()
}
func read(name, from string, reader io.Reader, formatter Formatter) error {
	ln := uint(1)
	r := bufio.NewScanner(reader)
	for r.Scan() {
		logrus.Debugf("read line from %s %s", name, from)
		formatter.Out(ln, name, from, r.Text())
		ln++
	}
	if err := r.Err(); err != nil {
		logrus.Errorf("read line error %s %s: %q", name, from, err)
		return err
	}
	logrus.Debugf("end of stream")
	return nil
}
