package main

import (
	"encoding/json"
	"fmt"

	"github.com/assembla/cony"
	"github.com/streadway/amqp"
)

type Policy struct {
	Package string `json:"package"`
	Content string `json:"content"`
}

const PolicyContent = `package policy

default allow = false

allow {
    some id
    input.method = "GET"
    input.path = ["salary", id]
    input.subject.user = id
}

allow {
    is_admin
}

is_admin {
    input.subject.groups[_] = "admin"
}`

func main() {
	cli := cony.NewClient(
		cony.URL("amqp://localhost:5672"),
		cony.Backoff(cony.DefaultBackoff),
	)

	exc := cony.Exchange{
		Name:       "test-exchange",
		Kind:       "fanout",
		AutoDelete: true,
	}
	cli.Declare([]cony.Declaration{
		cony.DeclareExchange(exc),
	})

	pbl := cony.NewPublisher(exc.Name, "test-key")
	cli.Publish(pbl)
	var policy Policy
	policy.Package = "policy"
	policy.Content = PolicyContent
	body, err := json.Marshal(&policy)
	if err != nil {
		panic(err)
	}
	go func() {
		for cli.Loop() {
			select {
			case err := <-cli.Errors():
				fmt.Printf("Client error: %v\n", err)
			case blocked := <-cli.Blocking():
				fmt.Printf("Client is blocked %v\n", blocked)
			}
		}
	}()
	err = pbl.Publish(amqp.Publishing{
		Body: body,
	})
	if err != nil {
		fmt.Printf("Client publish error: %v\n", err)
	}
}
