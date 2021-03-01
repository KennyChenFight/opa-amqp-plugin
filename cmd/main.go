package main

import (
	"fmt"
	"os"

	"github.com/KennyChenFight/opa-amqp-plugin/internal"
	"github.com/KennyChenFight/opa-amqp-plugin/plugin"
	"github.com/open-policy-agent/opa/cmd"
	"github.com/open-policy-agent/opa/runtime"
)

func main() {
	runtime.RegisterPlugin(internal.PluginName, plugin.Factory{})

	if err := cmd.RootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
