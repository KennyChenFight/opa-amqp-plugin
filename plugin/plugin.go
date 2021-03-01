package plugin

import (
	"encoding/json"

	"github.com/KennyChenFight/opa-amqp-plugin/internal"
	"github.com/open-policy-agent/opa/plugins"
)

type Factory struct{}

func (Factory) New(m *plugins.Manager, config interface{}) plugins.Plugin {

	m.UpdatePluginStatus(internal.PluginName, &plugins.Status{State: plugins.StateNotReady})

	return &internal.PolicyConsumer{
		Manager: m,
		Config:  config.(internal.Config),
	}
}

func (Factory) Validate(_ *plugins.Manager, config []byte) (interface{}, error) {
	parsedConfig := internal.Config{}
	var err error
	err = json.Unmarshal(config, &parsedConfig)
	return parsedConfig, err
}
