package main

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
)

func handleExec(inv *Inventory) {
	fs, _ := parseFilters(*execLimit)

	logrus.Infof("filters: %q", fs)
	logrus.Infof("command: %s", *execCommand)

	runner := &Runner{
		args:   *execCommand,
		dryRun: *execDryrun,
	}
	items := fs.filter(inv.Items)
	formatter := NewFormatter(*execFormat, items)

	err := runner.Run(formatter, items...)
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.Info("DONE")
}

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
func handleGet(inv *Inventory) {
	buf, err := yaml.Marshal(inv.Items[*getItem])
	if err != nil {
		logrus.Error(err)
	}
	fmt.Printf("%s\n", buf)
}
func handleList(inv *Inventory) {
	fs, _ := parseFilters(*listLimits)
	logrus.Infof("filters: %q", fs)
	items := fs.filter(inv.Items)
	result := map[string]map[string]string{}
	for _, item := range items {
		var name = item["name"]
		delete(item, "name")
		result[name] = item
	}
	switch *listFormat {
	case "yaml":
		buf, err := yaml.Marshal(result)
		if err != nil {
			logrus.Error(err)
		}
		fmt.Printf("%s\n", buf)
	case "simple":
		for name := range result {
			fmt.Println(name)
		}
	default:
		logrus.Error("unknown format")
	}

}
