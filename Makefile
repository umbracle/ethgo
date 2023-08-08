
.PHONY: build-artifacts
build-artifacts: 
	@echo "--> Build Artifacts"
	@sh -c ./scripts/build-artifacts.sh

.PHONY: lint
lint:
	golangci-lint run