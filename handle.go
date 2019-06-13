package main

import (
	"github.com/sirupsen/logrus"
)

func handleRun(inv *Inventory) {
	fs, _ := parseFilters(*runLimits)

	logrus.Infof("filters: %q", fs)
	logrus.Infof("command: %s", *runCommand)

	r := inv.Runner.Run
	if *runTemplate != "" {
		r = *runTemplate
	}

	args := toArgs(r, map[string][]string{
		"command": *runCommand,
	})

	runner := &Runner{
		args:   args,
		dryRun: *runDryrun,
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
func handleCp(inv *Inventory) {
	fs, _ := parseFilters(*cpLimits)

	logrus.Infof("filters: %q", fs)

	r := inv.Runner.Copy

	args := toArgs(r, map[string][]string{
		"src":  []string{*cpSrc},
		"dest": []string{*cpDest},
	})

	runner := &Runner{
		args:   args,
		dryRun: *cpDryrun,
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
func handleSet(inv *Inventory) {
	for _, name := range *setItem {
		old, ok := inv.Items[name]
		if !ok {
			old = map[string]string{}
		}

		for k, v := range *setLabels {
			old[k] = v
		}

		inv.Items[name] = old
	}
	err := SaveInventory(*inventory, inv)
	if err != nil {
		logrus.Error(err)
	}
}
