// +build integration

package integration

import (
	"flag"
)

var (
	// BaseServerAddress The base server address
	argServerAddress = flag.String("server-address", "localhost:8080", "The base server address")
)

func initHelper() {
	flag.Parse()
}
