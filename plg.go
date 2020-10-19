package main

import (
	"flag"
	"fmt"
)

func main() {

	typeFlag := flag.String("type", "", "record|serve")
	configFile := flag.String("config", "config.json", "configuration file")

	flag.Parse()
	config, err := parseConfig(*configFile)
	if err != nil {
		panic(err)
	}

	switch *typeFlag {
	case "record":
		fmt.Printf("Recording data from %d URLs. It will take at least %d seconds to complete.\n", len(config.Exporters), config.Time)
		Record(config)
		fmt.Printf("...done \n")
	case "serve":
		Serve(config)

	default:
		flag.Usage()
	}

}
