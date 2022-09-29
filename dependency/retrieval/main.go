package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/paketo-buildpacks/mri/dependency/retrieval/components"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

const versionFeed = "https://raw.githubusercontent.com/ruby/www.ruby-lang.org/master/_data/releases.yml"

// Retrieval for gets newer upstream versions of the ruby dependency from upstream
// and returns a metadata.json for new versions within buildpack.toml constraints
func main() {
	var flags struct {
		buildpackTomlPath string
		output            string
	}

	flag.StringVar(&flags.buildpackTomlPath, "buildpackTomlPath", "", "the path to the buildpack.toml file")
	flag.StringVar(&flags.output, "output", "", "path to file into which an output metadata JSON will be written")
	flag.Parse()
	if flags.buildpackTomlPath == "" {
		fail(errors.New(`missing required input "buildpackTomlPath"`))
	}
	if flags.output == "" {
		fail(errors.New(`missing required input "output"`))
	}

	buildpackConfig, err := cargo.NewBuildpackParser().Parse(flags.buildpackTomlPath)
	if err != nil {
		fail(err)
	}

	// Map where the key is a version and the value is a struct with version metadata
	releaseFetcher := components.NewReleaseFetcher(versionFeed)
	upstreamVersionMap, err := releaseFetcher.GetUpstreamReleases()
	if err != nil {
		fail(err)
	}

	upstreamVersions := []string{}
	for k := range upstreamVersionMap {
		upstreamVersions = append(upstreamVersions, k)
	}

	// Filter down the upstream versions against the buildpack.toml file
	newVersions, err := components.FindNewVersions("ruby", buildpackConfig, upstreamVersions)
	if err != nil {
		fail(err)
	}

	fmt.Printf("New versions: %v\n", newVersions)
	dependencies := []components.Dependency{}
	for _, version := range newVersions {
		// Validate the dependency checksum matches the upstream dependency
		// checkdum before we add it to the list of dependencies
		valid, err := components.Validate(upstreamVersionMap[version])
		if err != nil {
			fail(err)
		}
		if !valid {
			fail(errors.New(fmt.Sprintf("failed to validate dependency checksum for version %s", version)))
		}

		entries, err := components.GenerateMetadata(upstreamVersionMap[version], []string{"jammy", "bionic"}, components.NewLicenseRetriever(), components.NewDeprecationDateRetriever())
		if err != nil {
			fail(err)
		}
		dependencies = append(dependencies, entries...)
	}

	err = components.WriteOutput(flags.output, dependencies)
	if err != nil {
		fail(err)
	}

	fmt.Printf("Succeeded! Metadata written to %s\n", flags.output)
}

func fail(err error) {
	fmt.Printf("Error: %s", err)
	os.Exit(1)
}
