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

massive runner for server management

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

  exec [<flags>] <command>...
    running raw command

  run [<flags>] <command>...
    running command

  cp [<flags>] <src-file> <dest-file>
    copy file

  set [<flags>] <item>...
    Put item

  get <item>
    Get item

  list [<flags>]
    list item

```

## EXAMPLE

```
# Restart all worker's docker daemon
todo run -i example.yaml -l cluster=web,role=worker -- 'systemctl restart docker'
# Get docker daemon log
todo run -i example.yaml --show-name -l cluster=web,role=manager -- 'journalctl -u docker --since today'
# Add new item to inventory
todo set newworker cluster=web role=worker
# get item from inventory
todo get swarm-manager01

# Sometimes, run command in local
# get server list with address and label
todo exec -i example.yaml -- echo {{.name}} {{.addr}} {{.cluster}}
```

