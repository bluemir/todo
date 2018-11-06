package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// TODO output format

var (
	app       = kingpin.New("todo", "massive runner for server management")
	debug     = app.Flag("debug", "Enable debug mode.").Bool()
	loglevel  = app.Flag("verbose", "Log level").Short('v').Counter()
	inventory = app.Flag("inventory", "Inventory").Short('i').Default(".inventory.yaml").ExistingFile()

	run         = app.Command("run", "running command")
	runFormat   = run.Flag("format", "display format(json, text, simple, detail or free format").Default("simple").Short('f').String()
	runLimits   = run.Flag("limit", "condition that filter items").Short('l').Strings()
	runTemplate = run.Flag("runner", "").Short('r').String()
	runCommand  = run.Arg("command", "commands to run").Required().Strings()

	set          = app.Command("set", "Put item")
	setLabels    = set.Flag("label", "labels").Short('l').StringMap()
	setInputFile = set.Flag("input-file", "input file").File()
	setItem      = set.Arg("item", "items").Required().Strings()

	get     = app.Command("get", "Get item")
	getItem = get.Arg("item", "item name").Required().String()

	list       = app.Command("list", "list item")
	listLimits = list.Flag("limit", "condition that filter items").Short('l').Strings()
)
var VERSION string

func main() {
	app.Version(VERSION)
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	// adjust loglevel
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(logrus.Level(*loglevel + 3))

	inv, err := ParseInventory(*inventory)
	if err != nil {
		logrus.Error(err)
		return
	}

	switch cmd {
	case run.FullCommand():
		fs, _ := parseFilters(*runLimits)
		logrus.Infof("filters: %q", fs)
		logrus.Infof("command: %s", *runCommand)

		runner := &Runner{
			tmpl:    inv.Runner,
			command: strings.Join(*runCommand, " "),
		}
		items := fs.filter(inv.Items)
		formatter := NewFormatter(*runFormat, items)

		err := runner.Run(&textCollector{formatter}, items...)
		if err != nil {
			logrus.Error(err)
			return
		}

		logrus.Info("DONE")

	case set.FullCommand():
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

	case get.FullCommand():
		buf, err := yaml.Marshal(inv.Items[*getItem])
		if err != nil {
			logrus.Error(err)
		}
		fmt.Printf("%s\n", buf)
	case list.FullCommand():
		fs, _ := parseFilters(*listLimits)
		logrus.Infof("filters: %q", fs)
		items := fs.filter(inv.Items)
		result := map[string]map[string]string{}
		for _, item := range items {
			var name = item["name"]
			delete(item, "name")
			result[name] = item
		}
		buf, err := yaml.Marshal(result)
		if err != nil {
			logrus.Error(err)
		}
		fmt.Printf("%s\n", buf)
	}
}
