package main

import (
	"os"

	"github.com/kovetskiy/godocs"
	"github.com/kovetskiy/lorg"
	"github.com/zazab/hierr"
)

var (
	version = "2.0"
	docs    = `zabbixctl ` + version + os.ExpandEnv(`

  zabbixctl is tool for working with zabbix server api using command line
interface, it provides effective way for operating on statuses of triggers and
hosts latest data.

  zabbixctl must be configurated before using, configuration file should be
placed in ~/.config/zabbixctl.conf and must be written using following syntax:

    [server]
      address  = "zabbix.local"
      username = "admin"
      password = "password"

    [session]
      path = "~/.cache/zabbixctl.session"

  zabbixctl will authorize in 'zabbix.local' server using given user
credentials and save a zabbix session to a file ~/.cache/zabbixctl.session and
at second run will use saved session instead of new authorization, by the way
zabbix sessions have a ttl that by default equals to 15 minutes, so if saved
zabbix session is outdated, zabbixctl will repeat authorization and rewrite the
session file.

Usage:
  zabbixctl [options] -T [/<pattern>...]
  zabbixctl [options] -L <hostname>... [/<pattern>...]
  zabbixctl -h | --help
  zabbixctl --version

Workflow options:
  -T --triggers         Search on zabbix triggers statuses.
                         Triggers can be filtered using /<pattern> argument,
                         for example, search and acknowledge all triggers in a
                         problem state and match the word 'cache':
                            zabbixctl -Tp /cache
    -k --only-nack      Show only not acknowledged triggers.
    -x --severity       Specify minimum trigger severity.
                         Once for information, twice for warning,
                         three for disaster, four for high, five for disaster.
    -p --problem        Show triggers that have a problem state.
    -r --recent         Show triggers that have recently been in a problem state.
    -s --since <date>   Show triggers that have changed their state
                         after the given time.
                         [default: 7 days ago]
    -u --until <date>   Show triggers that have changed their state
                         before the given time.
    -m --maintenance    Show hosts in maintenance.
    -i --sort <fields>  Show triggers sorted by specified fields.
                         [default: lastchange,priority]
    -o --order <order>  Show triggers in specified order.
                         [default: DESC]
    -n --limit <count>  Show specified amount of triggers.
                         [default: 0]
    -f --noconfirm      Do not prompt acknowledge confirmation dialog.
    -a --acknowledge    Acknowledge all retrieved triggers.

  -L --latest-data      Search and show latest data for specified host(s).
                          Hosts can be searched using wildcard character '*'.
                          Latest data can be filtered using /<pattern> argument,
                          for example retrieve latest data for database nodes
                          and search information about replication:
                              zabbixctl -L dbnode-* /replication
    -g --graph          Show links on graph pages.

Common options:
  -c --config <path>    Use specified configuration file .
                         [default: $HOME/.config/zabbixctl.conf]
  -v --verbosity        Specify program output verbosity.
                         Once for debug, twice for trace.
  -h --help             Show this screen.
  --version             Show version.
`)
	usage = `
  zabbixctl [options] -T [-v]... [-x]... [<pattern>]...
  zabbixctl [options] -L [-v]... <pattern>...
  zabbixctl -h | --help
  zabbixctl --version
`
)

var (
	debugMode bool
	traceMode bool

	logger = getLogger()
)

func main() {
	args, err := godocs.Parse(
		docs, version, godocs.UsePager, godocs.Usage(usage),
	)
	if err != nil {
		fatalln(err)
	}

	switch args["--verbosity"].(int) {
	case 1:
		debugMode = true
		logger.SetLevel(lorg.LevelDebug)
	case 2:
		debugMode = true
		traceMode = true
		logger.SetLevel(lorg.LevelTrace)
	}

	config, err := NewConfig(args["--config"].(string))
	if err != nil {
		fatalln(
			hierr.Errorf(
				err,
				"problem with configuration",
			),
		)
	}

	zabbix, err := NewZabbix(
		config.Server.Address,
		config.Server.Username,
		config.Server.Password,
		config.Session.Path,
	)
	if err != nil {
		fatalln(err)
	}

	switch {
	case args["--triggers"].(bool):
		err = handleTriggers(zabbix, config, args)
	case args["--latest-data"].(bool):
		err = handleLatestData(zabbix, config, args)

	}

	if err != nil {
		fatalln(err)
	}
}
