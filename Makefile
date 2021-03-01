OPA_AMQP_PLUGIN := opa-amqp
AMQP_PUBLISHER := amqp-publisher

.PHONY: build-opa
build-opa:
	@echo "Building the $(OPA_AMQP_PLUGIN) binary..."
	go build -o build/bin/$(OPA_AMQP_PLUGIN) ./cmd/

.PHONY: build-publisher
build-publisher:
	@echo "Building the $(AMQP_PUBLISHER) binary..."
	go build -o build/bin/$(AMQP_PUBLISHER) ./publisher/