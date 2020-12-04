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

		// $BP_MRI_VERSION has the highest priority in setting the MRI version
		version := os.Getenv("BP_MRI_VERSION")
		versionSource := "BP_MRI_VERSION"

		// If $BP_MRI_VERSION is not set, check buildpack.yml for a version
		if version == "" {
			version, err = buildpackYMLParser.ParseVersion(filepath.Join(context.WorkingDir, BuildpackYMLSource))
			if err != nil {
				return packit.DetectResult{}, err
			}
			versionSource = BuildpackYMLSource
		}

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
