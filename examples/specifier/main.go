package main

import (
	"fmt"
	"log"

	"github.com/aquasecurity/go-pep440-version"
)

func main() {
	v, err := version.Parse("2.1")
	if err != nil {
		log.Fatal(err)
	}

	c, err := version.NewSpecifiers(">= 1.0, < 1.4 || > 2.0")
	if err != nil {
		log.Fatal(err)
	}

	if c.Check(v) {
		fmt.Printf("%s satisfies specifiers '%s'", v, c)
	}
}
