build:
	@echo "==> Building the CLI..."
	go build ../dice/cmd/dice

clean:
	@rm dice

.PHONY: clean build
