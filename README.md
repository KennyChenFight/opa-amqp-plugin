# OPA-AMQP-Plugin
Instead of using OPA embedded bundle to update policy, you can use this plugin to integrate amqp server and opa.<br>
This plugin will consume amqp queue after the opa start.<br>
So assume the amqp publisher publish the new policy then the consumer of opa-amqp-plugin will consume this message and send request to update opa policy.

## How to run
1. build opa-amqp-plugin: `make build-opa`
2. reference config:
```yaml
plugins:
  amqp_policy_consumer:
    amqpUrl: amqp://localhost:5672
    exchangeName: test-exchange
    routerKey: test-key
    queueName: test-queue
```
When running opa-amqp-plugin, we need provide above config file.<br>
This config should provide **amqpUrl**, **exchangeName**, **routerKey**, **queueName**, so it can start amqp consumer successfully.
3. run opa-amqp-plugin: `./build/bin/opa-amqp run --server --config-file config.yaml`
4. build amqp-publisher: `make build-publisher`
5. run amqp-publisher: `./build/bin/amqp-publisher` <br>
After running amqp-publisher, it will publish this rego policy: <br>
```rego
package policy

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
}
```
6. use curl commend to check policy if it updeted.: `curl localhost:8181/v1/data/policy/allow -d @input.json -H 'Content-Type: application/json'`<br>
this input.json file content is:<br>
```json
{
  "input": {
    "method": "GET",
    "path": ["salary", "bob"],
    "subject": {
      "user": "bob"
    }
  }
}
```   
or 
```json
{
  "input": {
    "subject": {
      "user": "bob",
      "groups": ["sales", "admin"]
    }
  }
}
```
These responses will show result: `{"result":true}`<br>
For more details about config file and input.json file should see [example](./example)