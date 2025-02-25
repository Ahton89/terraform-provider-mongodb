package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"terraform-provider-mongodb/internal/provider"
)

var (
	version = "0.0.0+dev"
	address = "registry.terraform.io/Ahton89/mongodb"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: address,
		Debug:   debug,
	}

	p := provider.New(version)

	defer func() {
		if prov, ok := p().(*provider.MongoDBProvider); ok && prov.MongoDB != nil {
			if err := prov.MongoDB.Disconnect(ctx); err != nil {
				log.Fatal(err.Error())
			}
		}
	}()

	err := providerserver.Serve(ctx, p, opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
