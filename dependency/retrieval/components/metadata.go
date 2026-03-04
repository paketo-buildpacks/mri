package components

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type Dependency struct {
	cargo.ConfigMetadataDependency
	Target string `json:"target,omitempty"`
}
type PlatformTarget struct {
	Stacks []string
	Target string
	OS     string
	Arch   string
}

//go:generate faux --interface License --output fakes/license.go
type License interface {
	LookupLicenses(dependencyName, sourceURL string) ([]interface{}, error)
}

//go:generate faux --interface DeprecationDate --output fakes/deprecation_date.go
type DeprecationDate interface {
	GetDate(feed, version string) (string, error)
}

// GenerateMetadata will generate Ruby dependency-specific metadata for each given platform target
func GenerateMetadata(release RubyRelease, platformTargets []PlatformTarget, licenseRetriever License, deprecationDate DeprecationDate) ([]Dependency, error) {
	dependencies := []Dependency{}
	licenses, err := licenseRetriever.LookupLicenses("ruby", release.URL.Gz)
	if err != nil {
		return dependencies, fmt.Errorf("could not get retrieve licenses: %w", err)
	}

	purl := GeneratePurl("ruby", release.Version, release.SHA256.Gz, release.URL.Gz)
	cpe := fmt.Sprintf("cpe:2.3:a:ruby-lang:ruby:%s:*:*:*:*:*:*:*", release.Version)
	srcChecksum := release.SHA256.Gz
	if algorithm, _, found := strings.Cut(release.SHA256.Gz, ":"); !found {
		srcChecksum = "sha256:" + algorithm
	}

	date, err := deprecationDate.GetDate("https://raw.githubusercontent.com/ruby/www.ruby-lang.org/master/_data/branches.yml", release.Version)
	if err != nil {
		return dependencies, err
	}

	for _, platformTarget := range platformTargets {
		// Validate Ruby version compatibility with target
		// Ruby 4.x requires GLIBC 2.38+, which is not available in jammy (GLIBC 2.35)
		if platformTarget.Target == "jammy" {
			version, err := semver.NewVersion(release.Version)
			if err != nil {
				return dependencies, err
			}
			constraint, err := semver.NewConstraint(">= 4.0")
			if err != nil {
				return dependencies, err
			}
			if constraint.Check(version) {
				// Skip Ruby 4.x for jammy
				continue
			}
		}

		dependency := Dependency{
			Target: platformTarget.Target,
		}

		stacks := platformTarget.Stacks

		dependency.ConfigMetadataDependency = cargo.ConfigMetadataDependency{
			Version:        release.Version,
			Source:         release.URL.Gz,
			SourceChecksum: srcChecksum,
			ID:             "ruby",
			Name:           "Ruby",
			CPE:            cpe,
			PURL:           purl,
			Stacks:         stacks,
			OS:             platformTarget.OS,
			Arch:           platformTarget.Arch,
			Licenses:       licenses,
		}

		if date != "" {
			dateFormatted, err := time.Parse("2006-01-02", date)
			if err != nil {
				return dependencies, fmt.Errorf("invalid EOL date: %w", err)
			}
			dependency.ConfigMetadataDependency.DeprecationDate = &dateFormatted
		}

		dependencies = append(dependencies, dependency)
	}

	return dependencies, nil
}

func WriteOutput(outputPath string, dependencies []Dependency) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(dependencies)
	if err != nil {
		//untested
		return err
	}
	return nil
}
