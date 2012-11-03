package main

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/mandira"
	"io/ioutil"
	"os"
)

var (
	USAGE = "Usage: mandira <template> <context>"
	HELP  = USAGE + `

Options:
  --version         show program's version and exit
  -h, --help        show this help
`
)

func parseArgs() (string, string) {
	args := []string{}
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			fmt.Print(HELP)
			os.Exit(0)
		default:
			args = append(args, arg)
		}
	}
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Error: mandira requires two arguments, template & context.\n")
		fmt.Print(USAGE)
		os.Exit(0)
	}
	return args[0], args[1]
}

func errExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	templatef, contextf := parseArgs()
	template, err := mandira.ParseFile(templatef)
	errExit(err)
	contextdata, err := ioutil.ReadFile(contextf)
	errExit(err)
	var context interface{}
	err = json.Unmarshal(contextdata, &context)
	errExit(err)
	fmt.Print(template.Render(context))
}
