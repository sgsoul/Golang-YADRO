build:
	go build -o xkcd ./cmd/xkcd

test:
	./xkcd -c pkg/config/config.yaml

bench:
	cd test/ && go test -bench=. -benchmem

clean:
	rm -f xkcd
