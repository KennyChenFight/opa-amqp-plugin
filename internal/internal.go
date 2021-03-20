package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/KennyChenFight/golib/amqplib"

	"github.com/assembla/cony"
	"github.com/open-policy-agent/opa/plugins"
)

const PluginName = "amqp_policy_consumer"

type Config struct {
	AMQPUrl      string `json:"amqpUrl"`
	ExchangeName string `json:"exchangeName"`
	ExchangeType string `json:"exchangeType"`
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
	connectionConfig := &amqplib.AMQPConnectionConfig{
		URL:          p.Config.AMQPUrl,
		ErrorHandler: nil,
	}
	queueConfig := &amqplib.AMQPQueueConfig{
		ExchangeName:        p.Config.ExchangeName,
		ExchangeType:        amqplib.ExchangeType(p.Config.ExchangeType),
		AutoDeclareExchange: false,
		QueueName:           p.Config.QueueName,
		RoutingKey:          p.Config.RouterKey,
		AutoDelete:          false,
	}
	client := amqplib.NewAMQPClient(connectionConfig)
	defer client.Close()
	consumer, err := client.NewConsumer(queueConfig)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()
	for delivery := range consumer.Consume() {
		log.Printf("Received body: %v\n", delivery.Body)
		if err := p.update(delivery.Body); err != nil {
			log.Printf("update err: %v\n", err)
		}
		delivery.Ack(false)
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
