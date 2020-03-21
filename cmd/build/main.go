package main

import (
	"os"
	"time"

	"github.com/cloudfoundry/packit"
	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/postal"
	"github.com/cloudfoundry/ruby-mri-cnb/ruby"
)

func main() {
	logEmitter := ruby.NewLogEmitter(os.Stdout)
	entryResolver := ruby.NewPlanEntryResolver(logEmitter)
	dependencyManager := postal.NewService(cargo.NewTransport())
	planRefinery := ruby.NewPlanRefinery()
	clock := ruby.NewClock(time.Now)

	packit.Build(ruby.Build(entryResolver, dependencyManager, planRefinery, logEmitter, clock))
}
