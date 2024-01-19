deploy: bootstrap
	sls deploy

bootstrap: go/*.go
	env CGO_ENABLED=0 go build -o $@ $^

clean:
	sls remove
	rm -f bootstrap
