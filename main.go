// Command cq-source-mws is a CloudQuery source plugin for the MWS Cloud Platform.
package main

import (
	"context"
	"log"

	"github.com/cloudquery/plugin-sdk/v4/serve"

	"cq-source-mws/plugin"
)

func main() {
	p := serve.Plugin(plugin.Plugin())

	if err := p.Serve(context.Background()); err != nil {
		log.Fatal(err)
	}
}
