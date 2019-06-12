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
	tmpl     string
	commands []string
	dryRun   bool
}

func (r *Runner) Run(formatter Formatter, items ...Item) error {
	parts := str.ToArgv(r.tmpl)

	tParts := []*template.Template{}

	for n, part := range parts {
		tmpl, err := template.New(part).Parse(part)
		if err != nil {
			logrus.Error(err)
			return err
		}
		tParts = append(tParts, tmpl)
		logrus.Debugf("%d %s", n, part)
	}

	eg, ctx := errgroup.WithContext(context.Background())
	for _, item := range items {
		result := []string{}
		for _, tmpl := range tParts {
			if tmpl.Name() == "{{.command}}" {
				result = append(result, r.commands...)
				continue
			}
			var cmd bytes.Buffer
			err := tmpl.Execute(&cmd, item)
			if err != nil {
				logrus.Error(err)
				return err
			}
			result = append(result, cmd.String())
		}

		logrus.Debugf("cmd:")
		for n, c := range result {
			logrus.Debugf("%d: %s", n, c)
		}

		if r.dryRun {
			result = append([]string{"echo"}, result...)
		}

		c := exec.CommandContext(ctx, result[0], result[1:]...)
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
