package main

import (
	"github.com/sirupsen/logrus"
)

func handleRun(inv *Inventory) {
	fs, _ := parseFilters(*runLimits)

	logrus.Infof("filters: %q", fs)
	logrus.Infof("command: %s", *runCommand)

	r := inv.Runner.Exec
	if *runTemplate != "" {
		r = *runTemplate
	}

	runner := &Runner{
		tmpl:     r,
		commands: *runCommand,
		dryRun:   *runDryrun,
	}
	items := fs.filter(inv.Items)
	formatter := NewFormatter(*runFormat, items)

	err := runner.Run(formatter, items...)
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.Info("DONE")
}
