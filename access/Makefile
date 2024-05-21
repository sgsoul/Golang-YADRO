server:
	go build -o server ./cmd/server

bench:
	cd test/ && go test -bench=. -benchmem

clean:
	rm -f server
	rm index.json
	rm cookie.txt

update:
	curl -X POST http://localhost:8080/update --cookie cookie.txt

get:
	curl -X GET 'http://localhost:8080/pics?search=binary+tree'

admin:
	curl -X POST http://localhost:8080/login -d '{"username":"admin","password":"adminpassword"}' \
	-H "Content-Type: application/json" --cookie-jar cookie.txt

user:
	curl -X POST http://localhost:8080/login -d '{"username":"user1","password":"password"}' \
	-H "Content-Type: application/json" --cookie-jar cookie.txt

limiter: #concurrency limiter 
	for i in {1..10}; do curl -s http://localhost:8080/pics?search=binary+tree & done

userlimiter: #user concurrency limiter 
	TOKEN=$(<cookie.txt) && for i in {1..10}; do curl -s "http://localhost:8080/pics?search=binary+tree" --cookie "token=$TOKEN" & sleep 1; done

newuser:
	curl -X POST http://localhost:8080/register -d '{"username":"newUser","password":"password","role":"user"}' -H "Content-Type: application/json"

newadmin:
	curl -X POST http://localhost:8080/register -d '{"username":"newAdmin","password":"adminpassword","role":"admin"}' \
	-H "Content-Type: application/json" --cookie cookie.txt
