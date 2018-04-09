package main

import (
	"github.com/cihub/seelog"
	"github.com/urfave/cli"
	"github.com/vrecan/death"

	"fmt"
	"os"
	"syscall"
)

const (
	statsdHostFlag = "host"
	statsdPortFlag = "port"
)

func main() {
	setupLogging()
	defer seelog.Flush()

	seelog.Infof("statsdFeed started")
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: statsdHostFlag, Value: "127.0.0.1", Usage: "statsd host"},
		cli.IntFlag{Name: statsdPortFlag, Value: 8125, Usage: "statsd port"},
	}

	app.Action = statsdFeed

	err := app.Run(os.Args)
	if err != nil {
		seelog.Errorf("cli.Run error: %v", err)
	}
	seelog.Infof("statsdFeed complete")
}

func statsdFeed(c *cli.Context) error {

	addr := fmt.Sprintf(`%s:%d`, c.String(statsdHostFlag), c.Int(statsdPortFlag))
	sender := NewStatsdSender(addr)
	if err := sender.Send(); err != nil {
		return seelog.Errorf("sender.Send error: %v", err)
	}
	death.NewDeath(syscall.SIGINT, syscall.SIGTERM).WaitForDeath(sender)

	return nil
}

func setupLogging() {
	logger, err := seelog.LoggerFromConfigAsString(`
	<?xml version="1.0"?>
	<seelog type="asynctimer" asyncinterval="1000000" minlevel="debug">
	  <outputs formatid="all">
		<console/>
		<rollingfile type="size" filename="statsdsender.log" maxsize="20000000" maxrolls="5"/>
	  </outputs>
	  <formats>
		<format id="all" format="%Date %Time [%LEVEL] [%FuncShort @ %File.%Line] - %Msg%n"/>
	  </formats>
	</seelog>
	`)
	if err != nil {
		panic(err)
	}

	err = seelog.ReplaceLogger(logger)
	if err != nil {
		panic(err)
	}
}
