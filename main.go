package main

import (
	"log"
	"strings"

	"github.com/spf13/pflag"
	"github.com/xperimental/locko2keepass/lckexp"
)

func main() {
	log.SetFlags(0)

	pflag.Parse()
	files := pflag.Args()

	for i, file := range files {
		log.Printf("Processing %d: %s", i, file)

		data, err := lckexp.ReadExport(file)
		if err != nil {
			log.Printf("Error reading export: %s", err)
			continue
		}

		for _, e := range data.RootEntries {
			printEntry(e, 0)
		}
	}
}

func printEntry(e *lckexp.LockoEntry, depth int) {
	log.Printf("%s%s (%s): %s -> %s", strings.Repeat(" ", depth), e.UUID, e.Title, e.Username, e.Password)
	for _, e := range e.Children {
		printEntry(e, depth+1)
	}
}
