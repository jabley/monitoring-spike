.PHONY: attack clean darwin dockerise linux lint report

DURATION=10m

darwin: -prep
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o monitoring-spike .

linux: -prep
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o monitoring-spike

native: -prep
	go build -o monitoring-spike .

clean:
	rm monitoring-spike ||:

-deps:
	go get golang.org/x/tools/cmd/goimports

-imports: -deps
	goimports -w .

-prep: -imports lint

lint:
	golint .
	go vet .

dockerise:
	docker build -t jabley/monitoring-spike-builder -f Dockerfile.build .
	docker run --rm jabley/monitoring-spike-builder | docker build -t jabley/monitoring-spike -f Dockerfile.run -

attack:
	vegeta attack -targets=targets.txt -duration=$(DURATION) -rate=50 > results.bin

report: attack
	cat results.bin | vegeta report
