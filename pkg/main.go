package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func main() {

	// Start listening to requests send from Grafana. This call is blocking so
	// it wont finish until Grafana shutsdown the process or the plugin choose
	// to exit close down by itself
	err := datasource.Serve(newDatasource())

	// Log any error if we could start the plugin.
	if err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}

func main2() {

	test := `{ "fred": new Date(1617713540609), "dave": new Date(1617713540609) }`
	var dateLiteralRegex = regexp.MustCompile(`^\${new Date\((\d+)\)}`)
	var dateLiteralRegex2 = regexp.MustCompile(`([^"]+)(new Date\(\d+\))([^"]+)`)

	result := dateLiteralRegex2.ReplaceAllString(test, `$1"${$2}"$3`)
	fmt.Println(result)

	result = dateLiteralRegex.ReplaceAllString("${new Date(1617713540609)}", `$1`)
	fmt.Println(result)
}
