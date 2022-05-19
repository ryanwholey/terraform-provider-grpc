.PHONY: acceptance test lint docs rundev

default: test


acceptance:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

test:
	go test ./...

lint:
	golangci-lint run ./...

docs:
	tfplugindocs

rundev:
	go build .
	mkdir -p ~/.terraform.d/plugins/terraform.local/local/grpc/0.0.1/darwin_amd64/
	mv terraform-provider-grpc ~/.terraform.d/plugins/terraform.local/local/grpc/0.0.1/darwin_amd64/terraform-provider-grpc_v0.0.1
	rm -rf terraform/.terraform.lock.hcl
	terraform -chdir=terraform/ init
	terraform -chdir=terraform/ apply
