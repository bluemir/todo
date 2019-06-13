package main

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
)

type Inventory struct {
	Items  map[string]Item
	Runner struct {
		Run  string
		Copy string
	}
}
type Item map[string]string

func ParseInventory(filename string) (*Inventory, error) {
	inv := &Inventory{}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(content, inv)
	if err != nil {
		return nil, err
	}
	if inv.Items == nil {
		inv.Items = map[string]Item{}
	}

	logrus.Infof("Inventory Path: %s Struct: %+v, ", filename, inv)
	return inv, nil
}
func SaveInventory(filename string, inv *Inventory) error {

	buf, err := yaml.Marshal(inv)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, buf, 644)
}
