.PHONY: build dockerise lint attack report

build: -imports
	go build -o monitoring-spike .

build-linux: -imports
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o monitoring-spike

-deps:
	go get golang.org/x/tools/cmd/goimports

-imports: -deps
	goimports -w .

lint:
	golint .
	go vet .

dockerise:
	docker build -t jabley/monitoring-spike-builder -f Dockerfile.build .
	docker run --rm jabley/monitoring-spike-builder | docker build -t jabley/monitoring-spike -f Dockerfile.run -

attack:
	vegeta attack -targets=targets.txt -duration=60s -rate=50 > results.bin

report: attack
	cat results.bin | vegeta report
