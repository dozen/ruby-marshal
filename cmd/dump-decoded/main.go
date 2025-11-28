package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	. "github.com/dozen/ruby-marshal"
)

func main() {
	var input *os.File
	var err error

	if len(os.Args) > 1 {
		input, err = os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		defer input.Close()
	} else {
		input = os.Stdin
	}

	bio := bufio.NewReader(input)
	dec := NewDecoder(bio)

	var obj interface{}
	err = dec.Decode(&obj)
	if err != nil {
		fmt.Printf("Error unmarshaling Ruby data: %v", err)
		return
	}

	// dump object hierarchy
	scs := spew.ConfigState{
		Indent:                  "    ",
		DisablePointerAddresses: true,
		DisableCapacities:       true,
	}
	scs.Dump(obj)
}
