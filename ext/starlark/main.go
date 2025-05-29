// Demo CLI for PE Starlark extension
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	starlarkext "github.com/tmc/pe/ext/starlark"
)

func main() {
	var (
		mode     = flag.String("mode", "eval", "Mode: eval, list, suite, validate, discover")
		testFile = flag.String("file", "", "Starlark test file")
		response = flag.String("response", "", "Response text to test")
		testFunc = flag.String("test", "", "Specific test function to run")
		dir      = flag.String("dir", ".", "Directory to search for .star files")
	)
	flag.Parse()

	switch *mode {
	case "eval":
		if *testFile == "" || *response == "" {
			log.Fatal("eval mode requires -file and -response flags")
		}
		err := starlarkext.RunStarlarkTest(*testFile, *response, *testFunc)
		if err != nil {
			log.Fatal(err)
		}

	case "list":
		if *testFile == "" {
			log.Fatal("list mode requires -file flag")
		}
		err := starlarkext.ListTestFunctions(*testFile)
		if err != nil {
			log.Fatal(err)
		}

	case "suite":
		if *testFile == "" || *response == "" {
			log.Fatal("suite mode requires -file and -response flags")
		}
		err := starlarkext.RunStarlarkTestSuite(*testFile, *response)
		if err != nil {
			log.Fatal(err)
		}

	case "validate":
		if *testFile == "" {
			log.Fatal("validate mode requires -file flag")
		}
		err := starlarkext.ValidateStarlarkFile(*testFile)
		if err != nil {
			log.Fatal(err)
		}

	case "discover":
		err := starlarkext.StarlarkDiscoverCommand([]string{*dir})
		if err != nil {
			log.Fatal(err)
		}

	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		fmt.Println("Available modes: eval, list, suite, validate, discover")
		os.Exit(1)
	}
}