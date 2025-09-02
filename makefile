SHELL=/bin/bash

VERSION := v1.0.7
DATE := $(shell date '+%Y-%m-%d')

TARGET_BIN := ntgp
BUILD_DIR := ./build
MAIN_DIR := .

# env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: lint build

lint:
	golangci-lint fmt ./...
	golangci-lint run ./...

build: lint
	CGO_ENABLED=0 \
	go build -ldflags \
		"-w -s \
		-X main.buildVersion=$(VERSION) \
		-X main.buildDate=$(DATE) \
		-X main.botToken=$(TOKEN) \
		-X main.chatID=$(CHAT) \
		-X main.messagesThreadID=$(TI) " \
		-o $(BUILD_DIR)/$(TARGET_BIN) main.go
