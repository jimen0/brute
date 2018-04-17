# brute

[![Go Report Card](https://goreportcard.com/badge/github.com/jimen0/brute)](https://goreportcard.com/report/github.com/jimen0/brute)
[![Documentation](https://godoc.org/github.com/jimen0/brute?status.svg)](https://godoc.org/github.com/jimen0/brute)

Package brute allows concurrently bruteforce subdomains for a domain using a list of DNS servers and querying a desired DNS record.

## Install

```
go get -u github.com/jimen0/brute
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jimen0/brute"
)

func main() {
	f, err := os.Open("/home/jimeno/top100.txt")
	if err != nil {
		log.Fatalf("could not open file: %v", err)
	}
	defer f.Close()

	out := make(chan string)
	done := make(chan struct{})
	go func() {
		for v := range out {
			fmt.Printf("%s\n", v)
		}
		done <- struct{}{}
	}()

	br := brute.Bruter{
		Domain:  "yahoo.com",
		Retries: 1,
		Record:  "A",
		Servers: []string{"1.1.1.1:53", "8.8.8.8:53", "1.0.0.1:53", "8.8.4.4:53"},
		Workers: 10, // increment this value to use more goroutines
	}
	err = br.Brute(context.Background(), f, out)
	if err != nil {
		log.Printf("failed to brute: %v", err)
	}
	<-done
}

```

## Test

Just run `go test -race -v github.com/jimen0/brute/...`


## Improvements

Send a PR or open an issue. Just make sure that your PR passes `gofmt`, `golint` and `govet`.
