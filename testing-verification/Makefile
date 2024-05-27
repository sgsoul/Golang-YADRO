PKG := ./...
COVERFILE := coverage.out
HTMLCOVERAGE := coverage.html
PACKAGES := github.com/sgsoul/internal/server \
			github.com/sgsoul/internal/service \
			github.com/sgsoul/internal/words \
			github.com/sgsoul/internal/service/search \
			github.com/sgsoul/internal/xkcd
PROJECT_PATH := .

server:
	go build -o server ./cmd/server

run: 
	./server &

test:
	CGO_ENABLED=1 go test -race -coverprofile=$(COVERFILE) $(PACKAGES)
	go tool cover -html=$(COVERFILE) -o $(HTMLCOVERAGE)

clean:
	rm -f server
	rm cookie.txt
	rm index.json
	rm $(HTMLCOVERAGE)

cleantest:
	rm $(COVERFILE)
	rm $(HTMLCOVERAGE)

update:
	curl -X POST http://localhost:8080/update --cookie cookie.txt

get:
	curl -X GET 'http://localhost:8080/pics?search=apple+doctor'

admin:
	curl -X POST http://localhost:8080/login -d '{"username":"admin","password":"adminpassword"}' \
	-H "Content-Type: application/json" --cookie-jar cookie.txt

user:
	curl -X POST http://localhost:8080/login -d '{"username":"user1","password":"password"}' \
	-H "Content-Type: application/json" --cookie-jar cookie.txt

newuser:
	curl -X POST http://localhost:8080/register -d '{"username":"newUser","password":"password","role":"user"}' -H "Content-Type: application/json"

newadmin:
	curl -X POST http://localhost:8080/register -d '{"username":"newAdmin","password":"adminpassword","role":"admin"}' \
	-H "Content-Type: application/json" --cookie cookie.txt

lint:
	golangci-lint run

TRIVY_CMD := trivy fs --scanners vuln --exit-code 1 $(PROJECT_PATH)

GOVULNCHECK_CMD := govulncheck 

sec:
	@echo "sec check with Trivy..."
	-$(TRIVY_CMD); TRIVY_EXIT_CODE=$$?; \
	if [ $$TRIVY_EXIT_CODE -eq 1 ]; then \
		echo "Trivy has discovered vulnerabilities. continue checking with Govulncheck..."; \
	elif [ $$TRIVY_EXIT_CODE -ne 0 ]; then \
		exit $$TRIVY_EXIT_CODE; \
	fi
	@echo "sec check with Govulncheck..."
	$(GOVULNCHECK_CMD)
	@echo "all checks have been successfully completed"

check:
	@echo "Checking search results..."
	@curl -X GET 'http://localhost:8080/pics?search=apple+doctor' \
	| grep -q 'apple_a_day' && echo "Found 'apple a day' comic" || (echo "Failed to find 'apple a day' comic" && exit 1)

wait:
	@echo "Waiting..."
	sleep 5

e2e: 
	go test ./test/e2e_test.go

e2etest: admin update get check	