package main

import (
	"fmt"
	"os"

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

	run        = app.Command("run", "running command")
	runFormat  = run.Flag("format", "display format(json, text, simple, detail or free format").Default("simple").Short('f').String()
	runLimits  = run.Flag("limit", "condition that filter items").Short('l').Strings()
	runCommand = run.Arg("command", "commands to run").Required().String()

	set       = app.Command("set", "Set item")
	setItem   = set.Arg("item", "item name").Required().String()
	setLabels = set.Arg("label", "labels").StringMap()

	get     = app.Command("get", "Get item")
	getItem = get.Arg("item", "item name").Required().String()

	list       = app.Command("list", "list item")
	listLimits = list.Flag("limit", "condition that filter items").Short('l').Strings()
	//listShowLabel = list.GetFlag

	exe        = app.Command("exec", "exec")
	exeLimits  = exe.Flag("limit", "condition that filter items").Short('l').Strings()
	exeFormat  = exe.Flag("format", "display format(json, text, simple, detail or free format").Default("simple").Short('f').String()
	exeCommand = exe.Arg("command", "command").Required().String()
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

		// Load complete
		executor := NewExecutor(inv.Runner)
		for name, item := range inv.Items {
			if !fs.isOk(name, item) {
				logrus.Debugf("next elem")
				continue
			}
			executor.Exec(name, *runCommand, item)
		}
		executor.Consume(&ConsumeOption{
			DisplayFormat: *runFormat,
		})
		logrus.Info("DONE")

	case set.FullCommand():
		old, ok := inv.Items[*setItem]
		if !ok {
			old = map[string]string{}
		}
		for k, v := range *setLabels {
			old[k] = v
		}
		inv.Items[*setItem] = old

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
		result := map[string]map[string]string{}
		for name, item := range inv.Items {

			if !fs.isOk(name, item) {
				logrus.Debugf("next elem")
				continue
			}
			result[name] = item
		}
		buf, err := yaml.Marshal(result)
		if err != nil {
			logrus.Error(err)
		}
		fmt.Printf("%s\n", buf)
	case exe.FullCommand():
		fs, _ := parseFilters(*exeLimits)
		logrus.Infof("filters: %q", fs)
		logrus.Infof("command: %s", *exeCommand)

		// Load complete
		executor := NewExecutor(*exeCommand)
		for name, item := range inv.Items {
			if !fs.isOk(name, item) {
				logrus.Debugf("next elem")
				continue
			}
			executor.Exec(name, "", item)
		}
		executor.Consume(&ConsumeOption{
			DisplayFormat: *exeFormat,
		})
		logrus.Info("DONE")
	}
}
