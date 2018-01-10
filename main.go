package main

import (
	"bytes"
	"fmt"
	"html/template"
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
	debug    = kingpin.Flag("debug", "Enable debug mode.").Bool()
	loglevel = kingpin.Flag("verbose", "Log level").Short('v').Counter()
	limits   = kingpin.Flag("limit", "Timeout waiting for ping.").OverrideDefaultFromEnvar("TODO_LIMIT").Short('l').Strings()
	commands = kingpin.Arg("command", "sadsda").Required().Strings()
)

func main() {
	kingpin.Parse()

	// adjust loglevel
	logrus.SetOutput(os.Stderr)
	logrus.SetLevel(logrus.Level(*loglevel))

	fs, _ := parseFilters()
	logrus.Debugf("filters: %q", fs)
	logrus.Debugf("command: %s", *commands)

	config := &Config{}

	content, err := ioutil.ReadFile("todo.yaml")
	if err != nil {
		logrus.Error(err)
	}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		logrus.Error(err)
	}
	logrus.Debugf("config: %+v", config)
	// Load complete

	tmpl, err := template.New("test").Parse(config.Runner)
	if err != nil {
		logrus.Error(err)
	}

	wg := new(sync.WaitGroup)
	lines := make(chan string, 32)
	done := make(chan struct{})

	go output(lines, done)

	for name, elem := range config.List {
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

type Config struct {
	List   map[string]map[string]string
	Runner string
}

func exe_cmd(cmd string, wg *sync.WaitGroup, out chan<- string) {
	defer wg.Done() // Need to signal to waitgroup that this goroutine is done

	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	result, err := exec.Command(head, parts...).Output()
	if err != nil {
		logrus.Error(err)
	}
	//fmt.Printf("%s", result)
	// var buf bytes.Buffer
	// r := bufio.NewReader(buf)
	// cmd.Stdout = buf
	// cmd.Stderr = os.Stderr
	// cmd.Run()
	// for line, err := r.ReadLine(); err == nil {
	// out <-line
	//}
	out <- string(result)
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
