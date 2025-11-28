package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is set via ldflags during build
var version = "dev"

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name outpost

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/outpost/outpost",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}

