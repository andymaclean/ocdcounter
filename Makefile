
all: test deploy

test: deploy
	./test/e2etest.sh

deploy: bootstrap serverless.yaml serverless/sls_api_handlers.yaml
	sls deploy

%: api/api.yaml %.mustache 
	@echo "Generating $@"
	mustache $^ > $@

bootstrap: go/*.go go/api_handlers.go
	cd go; go test
	env CGO_ENABLED=0 go build -o $@ $^

clean:	
	sls remove
	rm -f bootstrap
	rm -f venom.log
	rm -rf out
	rm -f serverless/sls_api_handlers.yaml
	rm -f go/api_handlers.go

	
