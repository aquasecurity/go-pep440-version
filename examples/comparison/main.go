package main

import (
	"fmt"
	"log"

	"github.com/aquasecurity/go-pep440-version"
)

func main() {
	v1, err := version.Parse("1.2a1")
	if err != nil {
		log.Fatal(err)
	}

	v2, err := version.Parse("1.2")
	if err != nil {
		log.Fatal(err)
	}

	// Comparison example. There is also GreaterThan, Equal, and just
	// a simple Compare that returns an int allowing easy >=, <=, etc.
	if v1.LessThan(v2) {
		fmt.Printf("%s is less than %s", v1, v2)
	}
}
