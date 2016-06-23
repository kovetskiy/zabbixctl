package main

import (
	"os"

	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
	"github.com/zazab/hierr"
)

var (
	version = "1.0"
	usage   = `zabbixctl ` + version + os.ExpandEnv(`

  zabbixctl is tool for working with zabbix server using command line
interface, it provides effective waay for operating on zabbix triggers and
their statuses, i.e. searching, sorting, showing and acknowledging triggers.

zabbixctl must be configurated before using, configuration file usually locates
  in ~/.config/zabbixctl.conf and must be written with following syntax:

  [server]
    address = "zabbix.hostname"
    username = "e.kovetskiy"
    password = "pa$$word"


Usage:
  zabbixctl [-v]... [options] -T [-x]... [<search>]...
  zabbixctl -h | --help
  zabbixctl --version

Workflow options:
  -T --triggers         Operate on zabbix triggers.
    -k --only-nack      Show unacknowledged triggers only.
    -x --severity       Specify minimum trigger severity.
                         Once for information, twice for warning,
                         three for disaster, four for high, five for disaster.
    -p --problem        Show triggers that have a problem state.
    -r --recent         Show triggers that have recently been in a problem state.
    -s --since <date>   Show triggers that have changed their state after
                         the given time.
                         [default: 7 days ago]
    -u --until <date>   Show triggers that have changed their state before
                         the given time.
    -m --maintenance    Show hosts in maintenance.
    -i --sort <fields>  Show triggers sorted by specified fields.
                         [default: lastchange,priority]
    -o --order <order>  Show triggers in specified order.
                         [default: DESC]
    -n --limit <count>  Show specified amount of triggers.
                         [default: 0]
    -f --noconfirm      Do not prompt acknowledge confirmation dialog.
    -a --acknowledge    Acknowledge triggers.

Common options:
  -c --config <path>    Use specified configuration file .
                         [default: $HOME/.config/zabbixctl.conf]
  -v --verbosity        Specify program output verbosity.
                         Once for debug, twice for trace.
  -h --help             Show this screen.
  --version             Show version.
`)
)

var (
	logger lorg.Logger
)

func main() {
	args, err := godocs.Parse(usage, version, godocs.UsePager)
	if err != nil {
		panic(err)
	}

	var (
		verbosity  = args["--verbosity"].(int)
		configPath = args["--config"].(string)
	)

	logger = getLogger(verbosity)

	config, err := NewConfig(configPath)
	if err != nil {
		fatalln(
			hierr.Errorf(
				err,
				"problem with configuration file using %s",
				configPath,
			),
		)
	}

	zabbix, err := NewZabbix(
		config.Server.Address,
		config.Server.Username,
		config.Server.Password,
	)
	if err != nil {
		fatalln(err)
	}

	switch {
	case args["--triggers"].(bool):
		err = handleModeTriggers(zabbix, config, args)
	}

	if err != nil {
		fatalln(err)
	}
}
