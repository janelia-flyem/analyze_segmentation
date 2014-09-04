package main

import (
	"analyze_segmentation/server"
	"fmt"
        "flag"
        "os"
)
	
var (
        showHelp = flag.Bool("help", false, "")
)

const helpMessage = `
Launches service that uses NeuroProof to assess segmentation quality.

Usage: analyze_segmentation <data-directory> <neuroproof-data-directory>
        -h, -help     (flag)          Show help message
`

func main() {
	flag.BoolVar(showHelp, "h", false, "Show help message")
	flag.Parse()

	if *showHelp {
		fmt.Printf(helpMessage)
		os.Exit(0)
	}

	if flag.NArg() != 2 {
		fmt.Println("Must provide two arguments")
		fmt.Println(helpMessage)
		os.Exit(0)
	}

        server := server.NewServer(flag.Arg(0), flag.Arg(1))
	server.Serve()
}
