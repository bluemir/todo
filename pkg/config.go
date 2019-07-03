package pkg

import (
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	// common
	Debug     bool
	LogLevel  int
	Inventory string

	// exec
	DryRun   bool
	Limit    []string
	Format   string
	Args     []string
	Template string

	// set get list
	Labels    map[string]string
	ItemNames []string
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Labels: map[string]string{},
	}
}

func (conf *AppConfig) Exec() error {
	inv, err := ParseInventory(conf.Inventory)
	if err != nil {
		return err
	}

	fs, _ := parseFilters(conf.Limit)

	logrus.Infof("filters: %q", fs)
	logrus.Infof("args: %s", conf.Args)
	logrus.Infof("templates: %s", conf.Template)

	// TODO check default

	args := toArgs(inv.Templates[conf.Template], conf.Args)

	runner := &Runner{
		args:   args,
		dryRun: conf.DryRun,
	}
	items := fs.filter(inv.Items)
	formatter := NewFormatter(conf.Format, items)

	if err := runner.Run(formatter, conf.Args, items...); err != nil {
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
