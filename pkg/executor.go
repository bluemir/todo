package pkg

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
	"text/template"

	"github.com/mgutz/str"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Runner struct {
	args   []string
	dryRun bool
}

func toArgs(tmplStr string, optArgs []string) []string {
	parts := str.ToArgv(tmplStr)
	result := []string{}
	for _, part := range parts {
		if strings.EqualFold(part, "{{args}}") {
			result = append(result, optArgs...)
		} else {
			result = append(result, part)
		}
	}
	return result
}

func (r *Runner) parts(args []string) ([]*template.Template, error) {
	fMap := map[string]interface{}{
		"args": func() string {
			return strings.Join(args, " ")
		},
		"arg": func(index int) string {
			logrus.Tracef("arg index: %d", index)
			if index < 1 || index > len(args) {
				logrus.Warnf("index of bound: {{arg %d}}", index)
				return ""
			}
			return args[index-1]
		},
	}

	tParts := []*template.Template{}
	for n, part := range r.args {
		logrus.Tracef("part: %s", part)
		tmpl, err := template.New("__").Funcs(fMap).Parse(part)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		tParts = append(tParts, tmpl)
		logrus.Debugf("%d %s", n, part)
	}
	return tParts, nil
}
func (r *Runner) Run(formatter Formatter, args []string, items ...Item) error {
	templates, err := r.parts(args)
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(context.Background())
	for _, item := range items {
		result, err := render(templates, item)
		if err != nil {
			return err
		}

		logrus.Debugf("cmd:")
		for n, c := range result {
			logrus.Debugf("%d: %s", n, c)
		}

		if r.dryRun {
			result = append([]string{"echo"}, result...)
		}
		if len(result) < 1 {
			return errors.Errorf("command is blank")
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

		eg.Go(read(item["name"], "stdout", stdout, formatter))
		eg.Go(read(item["name"], "stderr", stderr, formatter))

		eg.Go(c.Start)
	}
	return eg.Wait()
}
func render(t []*template.Template, item Item) ([]string, error) {
	result := []string{}
	for _, tmpl := range t {
		var cmd bytes.Buffer
		err := tmpl.Execute(&cmd, item)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		result = append(result, cmd.String())
	}
	return result, nil
}
func read(name, from string, reader io.Reader, formatter Formatter) func() error {
	return func() error {
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
}
