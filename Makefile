OUTPUT ?= go-sqlite

.DEFAULT_GOAL := all

build:
	CGO_ENABLED=0 go build -o $(OUTPUT)

all: build
