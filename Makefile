
all: test deploy

test: deploy
	./test/e2etest.sh

deploy: bootstrap serverless.yaml
	sls deploy

bootstrap: go/*.go
	cd go; go test
	env CGO_ENABLED=0 go build -o $@ $^

clean:
	sls remove
	rm -f bootstrap
	rm -f venom.log
	rm -rf out
	
