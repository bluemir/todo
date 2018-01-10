package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug     = kingpin.Flag("debug", "Enable debug mode.").Bool()
	loglevel  = kingpin.Flag("verbose", "Log level").Short('v').Counter()
	inventory = kingpin.Flag("inventory", "Inventory").Short('i').Required().ExistingFile()
	limits    = kingpin.Flag("limit", "condition that filter items").OverrideDefaultFromEnvar("TODO_LIMIT").Short('l').Strings()
	commands  = kingpin.Arg("command", "commands to run").Required().Strings()
)
var VERSION string

func main() {
	kingpin.Version(VERSION)
	kingpin.Parse()

	// adjust loglevel
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(logrus.Level(*loglevel))

	fs, _ := parseFilters()
	logrus.Debugf("filters: %q", fs)
	logrus.Debugf("command: %s", *commands)

	inv := &Inventory{}

	content, err := ioutil.ReadFile(*inventory)
	if err != nil {
		logrus.Error(err)
	}
	err = yaml.Unmarshal(content, inv)
	if err != nil {
		logrus.Error(err)
	}

	logrus.Debugf("Inventory Path: %s Struct: %+v, ", *inventory, inv)
	// Load complete

	tmpl, err := template.New("test").Parse(inv.Runner)
	if err != nil {
		logrus.Error(err)
	}

	wg := new(sync.WaitGroup)
	lines := make(chan string, 32)
	done := make(chan struct{})

	go output(lines, done)

	for name, elem := range inv.Items {
		if !fs.isOk(elem) {
			logrus.Debugf("next elem")
			continue
		}

		var cmd bytes.Buffer
		elem["name"] = name
		elem["command"] = strings.Join(*commands, " ")
		err = tmpl.Execute(&cmd, elem)
		if err != nil {
			logrus.Error(err)
		}
		logrus.Infof("running command: %s to '%s'", name, cmd.String())

		wg.Add(1)
		go exe_cmd(cmd.String(), wg, lines)
	}

	wg.Wait()
	close(lines)
	<-done
}

type Inventory struct {
	Items  map[string]map[string]string
	Runner string
}

func exe_cmd(cmd string, wg *sync.WaitGroup, out chan<- string) {
	defer wg.Done() // Need to signal to waitgroup that this goroutine is done

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	c := exec.Command(head, parts...)
	var buf = &bytes.Buffer{}

	r := bufio.NewReader(buf)
	c.Stdout = buf
	c.Stderr = os.Stderr
	c.Run()
	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			logrus.Infof("exec end.")
			return
		}
		if err != nil {
			logrus.Error(err)
			return
		}
		out <- string(line)
	}
	// send stdout one-line by one-line
}

type filter map[string]string

func (f filter) isOk(labels map[string]string) bool {
	for key, value := range f {
		val, ok := labels[key]
		if !ok {
			return false
		}

		if val != value {
			return false
		}
	}
	return true
}
func parseFilters() (filters, error) {
	// TODO error handler
	filters := []filter{}

	// AND in same string, OR in other string
	for _, str := range *limits {
		f := filter{}
		arr := strings.Split(str, ",")
		for _, s := range arr {
			a := strings.SplitN(s, "=", 2)
			f[a[0]] = a[1]
		}
		filters = append(filters, f)
	}
	return filters, nil
}

type filters []filter

func (fs filters) isOk(labels map[string]string) bool {
	for _, f := range fs {
		if f.isOk(labels) {
			return true
		}
	}
	return false
}
func output(lines <-chan string, done chan<- struct{}) {
	defer close(done)
	for c := range lines {
		fmt.Println(c)
	}
}
