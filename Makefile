.PHONY:

build:
	mkdir -p ./bin
	CGO_ENABLED=0 go build -o ./bin cmd/yafwd/yafwd.go

run_docker:
	cd ./build && docker-compose exec firewall /app/yafwd

test:
	go test -v

test_docker: build
	go test -c -o ./bin/yafw_test
	cd ./build && docker-compose exec firewall /app/yafw_test -test.v
