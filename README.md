# Controller

[![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis]

[GoReportCard]: https://goreportcard.com/report/github.com/amalgam8/controller
[GoReportCard Widget]: https://goreportcard.com/badge/github.com/amalgam8/controller
[Travis]: https://travis-ci.org/amalgam8/controller
[Travis Widget]: https://travis-ci.org/amalgam8/controller.svg?branch=master

Amalgam8 Controller

## Running
Locally:

    go run ./main.go

For help on command line arguments, run:

    go run ./main.go -h

In a container:

    ./build.sh
    docker run -p 6379:6379 --name controller controller
