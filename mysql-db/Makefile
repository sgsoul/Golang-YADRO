server:
	go build -o server ./cmd/server

bench:
	cd test/ && go test -bench=. -benchmem

clean:
	rm -f server
	rm index.json

post:
	curl -X POST 'http://localhost:8080/update'

get:
	curl -X GET 'http://localhost:8080/pics?search=binary+tree'