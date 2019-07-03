package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bluemir/todo/pkg"
)

var VERSION string

func main() {
	conf := pkg.NewAppConfig()

	app := kingpin.New("todo", "massive runner for server management")
	app.Flag("debug", "Enable debug mode.").BoolVar(&conf.Debug)
	app.Flag("verbose", "Log level").Short('v').CounterVar(&conf.LogLevel)
	app.Flag("inventory", "Inventory").Short('i').Default(".inventory.yaml").ExistingFileVar(&conf.Inventory)

	exec := app.Command("exec", "running command (alias run)").Alias("run")
	exec.Flag("format", "display format(json, text, simple, detail or free format)").Default("simple").Short('f').StringVar(&conf.Format)
	exec.Flag("limit", "condition that filter items").Short('l').StringsVar(&conf.Limit)
	exec.Flag("dry-run", "Dry Run").Default("false").BoolVar(&conf.DryRun)
	exec.Flag("templates", "running template").Short('t').Default("default").StringVar(&conf.Template)
	exec.Arg("args", "args to run").StringsVar(&conf.Args)

	set := app.Command("set", "Put item")
	set.Flag("label", "labels").Short('l').StringMapVar(&conf.Labels)
	set.Arg("item", "items").Required().StringsVar(&conf.ItemNames)

	get := app.Command("get", "Get item")
	get.Arg("item", "item name").Required().StringsVar(&conf.ItemNames)

	list := app.Command("list", "list item")
	list.Flag("limit", "condition that filter items").Short('l').StringsVar(&conf.Limit)
	list.Flag("format", "display format(simple, yaml)").Short('f').Default("simple").StringVar(&conf.Format)

	app.Version(VERSION)

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	// adjust loglevel
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(logrus.Level(conf.LogLevel + 3))

	var err error
	switch cmd {
	case exec.FullCommand():
		err = conf.Exec()
	case set.FullCommand():
		err = conf.Set()
	case get.FullCommand():
		err = conf.Get()
	case list.FullCommand():
		err = conf.List()
	}

	if err != nil {
		logrus.Error(err)
	}
}
