package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/assembla/cony"
	"github.com/open-policy-agent/opa/plugins"
)

const PluginName = "amqp_policy_consumer"

type Config struct {
	AMQPUrl      string `json:"amqpUrl"`
	ExchangeName string `json:"exchangeName"`
	RouterKey    string `json:"routerKey"`
	QueueName    string `json:"queueName"`
}

type PolicyConsumer struct {
	Manager *plugins.Manager
	Client  *cony.Client
	Config  Config
}

func (p *PolicyConsumer) Start(ctx context.Context) error {
	p.Manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})
	go p.listen()
	return nil
}

func (p *PolicyConsumer) Stop(ctx context.Context) {
	p.Client.Close()
	p.Manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateNotReady})
}

func (p *PolicyConsumer) Reconfigure(ctx context.Context, config interface{}) {
	p.Config = config.(Config)
}

func (p *PolicyConsumer) listen() {
	cli := cony.NewClient(
		cony.URL(p.Config.AMQPUrl),
		cony.Backoff(cony.DefaultBackoff),
	)
	p.Client = cli

	que := &cony.Queue{
		AutoDelete: true,
		Name:       p.Config.QueueName,
	}
	exc := cony.Exchange{
		Name:       p.Config.ExchangeName,
		Kind:       "fanout",
		AutoDelete: true,
	}
	bnd := cony.Binding{
		Queue:    que,
		Exchange: exc,
		Key:      p.Config.RouterKey,
	}
	p.Client.Declare([]cony.Declaration{
		cony.DeclareQueue(que),
		cony.DeclareExchange(exc),
		cony.DeclareBinding(bnd),
	})

	// Declare and register a consumer
	cns := cony.NewConsumer(
		que,
	)
	p.Client.Consume(cns)
	for cli.Loop() {
		select {
		case msg := <-cns.Deliveries():
			fmt.Printf("Received body: %v\n", msg.Body)
			if err := p.update(msg.Body); err != nil {
				fmt.Printf("update err: %v\n", err)
			}
			msg.Ack(false)
		case err := <-cns.Errors():
			fmt.Printf("Consumer error: %v\n", err)
		case err := <-p.Client.Errors():
			fmt.Printf("Client error: %v\n", err)
		}
	}
}

type Policy struct {
	Package string `json:"package"`
	Content string `json:"content"`
}

func (p *PolicyConsumer) update(body []byte) error {
	var policy Policy
	if err := json.Unmarshal(body, &policy); err != nil {
		return err
	}
	client := http.DefaultClient
	requestBody := bytes.NewBufferString(policy.Content)
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://localhost:8181/v1/policies/%s", policy.Package), requestBody)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("fail update policy: %d", res.StatusCode))
	}
	return nil
}
