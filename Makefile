.PHONY: build dockerise

build: -imports
	go build -o monitoring-spike .

-imports:
	goimports -w .

dockerise:
	docker build -t jabley/monitoring-spike-builder -f Dockerfile.build .
	docker run --rm jabley/monitoring-spike-builder | docker build -t jabley/monitoring-spike -f Dockerfile.run -
