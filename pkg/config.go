package pkg

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	// common
	Debug     bool
	LogLevel  int
	Inventory string

	// run exec cp
	DryRun bool
	Limit  []string

	// run exec
	Format  string
	Command []string

	Template string

	// cp
	Src  string
	Dest string

	// set
	Labels    map[string]string
	ItemNames []string
	//
}

func NewAppConfig() *AppConfig {
	return &AppConfig{}
}

func (conf *AppConfig) Exec() error {
	inv, err := ParseInventory(conf.Inventory)
	if err != nil {
		return err
	}

	fs, _ := parseFilters(conf.Limit)

	logrus.Infof("filters: %q", fs)
	logrus.Infof("command: %s", conf.Command)

	runner := &Runner{
		args:   conf.Command,
		dryRun: conf.DryRun,
	}
	items := fs.filter(inv.Items)
	formatter := NewFormatter(conf.Format, items)

	if err := runner.Run(formatter, map[string]string{
		"command": strings.Join(conf.Command, " "),
	}, items...); err != nil {
		return err
	}

	logrus.Info("DONE")
	return nil
}
func (conf *AppConfig) Run() error {
	inv, err := ParseInventory(conf.Inventory)
	if err != nil {
		return err
	}

	fs, _ := parseFilters(conf.Limit)

	logrus.Infof("filters: %q", fs)
	logrus.Infof("command: %s", conf.Command)

	r := inv.Runner.Run
	if conf.Template != "" {
		r = conf.Template
	}

	args := toArgs(r, map[string][]string{
		"command": conf.Command,
	})

	runner := &Runner{
		args:   args,
		dryRun: conf.DryRun,
	}
	items := fs.filter(inv.Items)
	formatter := NewFormatter(conf.Format, items)

	if err := runner.Run(formatter, map[string]string{
		"command": strings.Join(conf.Command, " "),
	}, items...); err != nil {
		return err
	}

	logrus.Info("DONE")
	return nil
}
func (conf *AppConfig) Copy() error {
	inv, err := ParseInventory(conf.Inventory)
	if err != nil {
		return err
	}

	fs, _ := parseFilters(conf.Limit)

	logrus.Infof("filters: %q", fs)

	r := inv.Runner.Copy

	args := toArgs(r, map[string][]string{
		"src":  []string{conf.Src},
		"dest": []string{conf.Dest},
	})

	runner := &Runner{
		args:   args,
		dryRun: conf.DryRun,
	}
	items := fs.filter(inv.Items)
	formatter := NewFormatter(conf.Format, items)

	if err := runner.Run(formatter, map[string]string{
		"src":  conf.Src,
		"dest": conf.Dest,
	}, items...); err != nil {
		return err
	}

	logrus.Info("DONE")
	return nil
}
func (conf *AppConfig) Set() error {
	inv, err := ParseInventory(conf.Inventory)
	if err != nil {
		return err
	}
	for _, name := range conf.ItemNames {
		old, ok := inv.Items[name]
		if !ok {
			old = map[string]string{}
		}

		for k, v := range conf.Labels {
			old[k] = v
		}

		inv.Items[name] = old
	}

	if err := SaveInventory(conf.Inventory, inv); err != nil {
		return err
	}
	return nil
}
func (conf *AppConfig) Get() error {
	inv, err := ParseInventory(conf.Inventory)
	if err != nil {
		return err
	}
	for _, name := range conf.ItemNames {
		buf, err := yaml.Marshal(inv.Items[name])
		if err != nil {
			logrus.Error(err)
		}
		fmt.Printf("%s\n", buf)
	}
	return nil
}
func (conf *AppConfig) List() error {
	inv, err := ParseInventory(conf.Inventory)
	if err != nil {
		return err
	}

	fs, _ := parseFilters(conf.Limit)
	logrus.Infof("filters: %q", fs)
	items := fs.filter(inv.Items)
	result := map[string]map[string]string{}
	for _, item := range items {
		var name = item["name"]
		delete(item, "name")
		result[name] = item
	}
	switch conf.Format {
	case "yaml":
		buf, err := yaml.Marshal(result)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", buf)
	case "simple":
		for name := range result {
			fmt.Println(name)
		}
	default:
		logrus.Error("unknown format")
	}
	return nil
}
