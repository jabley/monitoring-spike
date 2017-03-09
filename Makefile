.PHONY: build dockerise lint attack report

build: -imports
	go build -o monitoring-spike .

-imports:
	goimports -w .

lint:
	golint .
	go vet .

dockerise:
	docker build -t jabley/monitoring-spike-builder -f Dockerfile.build .
	docker run --rm jabley/monitoring-spike-builder | docker build -t jabley/monitoring-spike -f Dockerfile.run -

attack:
	vegeta attack -targets=targets.txt -duration=60s -rate=200 > results.bin

report: attack
	cat results.bin | vegeta report
