package mri

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

type BuildPlanMetadata struct {
	VersionSource string `toml:"version-source"`
	Version       string `toml:"version"`
}

func Detect(buildpackYMLParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement
		var err error

		// If versions are provided via BP_MRI_VERSION and/or buildpack.yml:
		// Detection will pass all versions as build plan requirements.
		// The build phase is responsible for using a priority mapping to select correct version.
		// This will allow for greater clarity in log output if the user has set version through multiple configurations.

		// check $BP_MRI_VERSION
		version := os.Getenv("BP_MRI_VERSION")
		versionSource := "BP_MRI_VERSION"

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: MRI,
				Metadata: BuildPlanMetadata{
					VersionSource: versionSource,
					Version:       version,
				},
			})
		}

		// check buildpack.yml
		version, err = buildpackYMLParser.ParseVersion(filepath.Join(context.WorkingDir, BuildpackYMLSource))
		if err != nil {
			return packit.DetectResult{}, err
		}
		versionSource = BuildpackYMLSource

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: MRI,
				Metadata: BuildPlanMetadata{
					VersionSource: versionSource,
					Version:       version,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: MRI},
				},
				Requires: requirements,
			},
		}, nil
	}
}
