package pkg

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"regexp"
	"text/template"

	"github.com/mgutz/str"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Runner struct {
	args   []string
	dryRun bool
}

func toArgs(tmplStr string, opts map[string][]string) []string {
	re := regexp.MustCompile("^{{([a-z]+)}}$")
	parts := str.ToArgv(tmplStr)
	result := []string{}
	for _, part := range parts {
		m := re.FindStringSubmatch(part)
		if len(m) < 2 {
			result = append(result, part)
			continue
		}
		if v, ok := opts[m[1]]; ok {
			result = append(result, v...)
		} else {
			result = append(result, part)
		}
	}
	return result
}

func (r *Runner) parts(opt map[string]string) ([]*template.Template, error) {
	// TODO funcMap
	fMap := map[string]interface{}{}

	for k, v := range opt {
		fMap[k] = func() string {
			return v
		}
	}

	tParts := []*template.Template{}
	for n, part := range r.args {
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
func (r *Runner) Run(formatter Formatter, opts map[string]string, items ...Item) error {
	templates, err := r.parts(opts)
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
