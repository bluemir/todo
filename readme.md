#TODO
massive command runner based on label


## HOW TO BUILD
```
git clone https://github.com/bluemir/todo
cd todo
make
```

## HOW TO USE
```
usage: todo [<flags>] <command> [<args> ...]

massive command runner for server management

Flags:
      --help         Show context-sensitive help (also try --help-long and --help-man).
      --debug        Enable debug mode.
  -v, --verbose ...  Log level
  -i, --inventory=.inventory.yaml
                     Inventory
      --version      Show application version.

Commands:
  help [<command>...]
    Show help.

  exec [<flags>] [<args>...]
    running command (alias run)

  set [<flags>] <item>...
    Put item

  get <item>...
    Get item

  list [<flags>]
    list item

```

### Set items

```
usage: todo set [<flags>] <item>...

Set item with label

Flags:
  -l, --label=LABEL ...  labels

Args:
  <item>  items

```


```
# Register new item
todo set -l cluster=web -l role=worker worker01.web

# Change label
todo set -l role=other worker01.web

# Multiple item
todo set -l cluster=web -l role=worker worker01.web worker02.web

# setting messive item
# If you want register massive amount item, use `seq` and `xargs`
seq -f "worker%02g.web.my-service" 1 70 | xargs todo set -l cluster=web -l role=worker

```

### run command

```
usage: todo exec [<flags>] [<args>...]

running command (alias run)

Flags:
  -f, --format="simple"      display format(json, text, simple, detail or free format)
  -l, --limit=LIMIT ...      condition that filter items
      --dry-run              Dry Run
  -t, --templates="default"  running template

Args:
  [<args>]  args to run

```

#### Set templates in inventory file

```
#.inventory.yaml
...
Templates:
  rsh: 'rsh {{.addr}} {{args}}'
  scp: 'scp {{arg 1}} {{.addr}}:{{arg 2}}'
  ssh: 'ssh -n {{.addr}} -C {{args}}'
  #default: '{{args}}'

```

* `{{args}}` : full arguments
* `{{arg N}}` : n-th argument
* `{{.LABEL}}` : replace items label

#### Running commands

```
# simple ping
todo run -- ping "{{.addr}}"
# filter with label
todo run -l cluster=web -- ping {{.addr}}
# run with template
todo run -t ssh -- uname -a
# AND condition
todo run -l cluster=web,role=worker -- ./run-server.sh
# OR condition
todo run -l cluster=web -l cluster=local -- ping {{.addr}}
# composite
# (cluster=web AND role=worker) OR (cluster=local AND role=test)
todo run -t ssh -l cluster=web,role=worker -l cluster=local,role=test -- cat /var/log/messages

```


## Examples

```
# Register nodes
seq -f "worker%02g.web.my-service" 1 70 | xargs todo set -i example.yaml -l cluster=web -l role=worker

# check ping
todo run -i example.yaml -l cluster=web,role=worker -- ping "{{.addr}}"

# Start all worker's docker daemon
todo run -i example.yaml -l cluster=web,role=worker -t ssh -- sudo systemctl start docker

# Get docker daemon log and tailing (with pipe)
todo run -i example.yaml -l cluster=web,role=manager -t ssh -- 'journalctl -u docker --since today -f | grep -v debug'

# check kernel version
todo run -i example.yaml -t ssh -- uname -a

# copy scripts to nodes
todo run -i example.yaml -l cluster=web,role=worker -t scp -- security-check.sh .

```



