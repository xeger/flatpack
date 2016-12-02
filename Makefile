#! /usr/bin/make
#
# Makefile for flatpack
#
# Interesting targets:
#   cover:    run all tests and display detailed coverage report
#   test:     run all tests and produce coverage summary

.PHONY: cover test

SHELL=/bin/bash

test: $(GOPATH)/bin/ginkgo
	ginkgo --randomizeAllSpecs --randomizeSuites --failOnPending -cover

cover: test
	go tool cover -html=flatpack.coverprofile;

$(GOPATH)/bin/ginkgo:
	go install github.com/onsi/ginkgo/ginkgo
