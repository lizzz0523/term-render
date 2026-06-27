.DEFAULT_GOAL := build

build:
	go build -ldflags="-s -w" -o term-render

run:
	go run . $(ARGS)

clear:
	rm -f term-render
