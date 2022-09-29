package components

import (
	"sort"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"k8s.io/utils/strings/slices"
)

// FindNewVersions will take in a dependency ID, a buildpack.toml content in the form of a
// cargo.Config, and a slice of all upstream versions available. It will filter
// the upstream versions by buildpack.toml constraint and ID, and then return versions
// that conform to the constraint, number of patches, and are not already
// present in the buildpack.toml
func FindNewVersions(id string, buildpackConfig cargo.Config, allVersions []string) ([]string, error) {
	newVersions := []string{}

	for _, c := range buildpackConfig.Metadata.DependencyConstraints {
		if c.ID != id {
			continue
		}
		constraint, err := semver.NewConstraint(c.Constraint)
		if err != nil {
			return nil, err
		}

		// versions in the buildpack.toml that we already have
		existingVersions := []string{}
		for _, dependency := range buildpackConfig.Metadata.Dependencies {
			version := semver.MustParse(dependency.Version)
			if constraint.Check(version) && id == dependency.ID {
				existingVersions = append(existingVersions, dependency.Version)
			}
		}

		matchingVersions := []string{}
		for _, v := range allVersions {
			version := semver.MustParse(v)
			if constraint.Check(version) {
				matchingVersions = append(matchingVersions, v)
			}
		}

		sort.Slice(matchingVersions, func(i, j int) bool {
			iVersion := semver.MustParse(matchingVersions[i])
			jVersion := semver.MustParse(matchingVersions[j])
			return iVersion.LessThan(jVersion)
		})

		// If there are more patches allowed than available, return all
		// If not, return just the # of patches from the list
		// Exclude pre-existing versions from new versions in both cases
		if c.Patches > len(matchingVersions) {
			for _, match := range matchingVersions {
				if !slices.Contains(existingVersions, match) {
					newVersions = append(newVersions, match)
				}
			}
		} else {
			for i := len(matchingVersions) - int(c.Patches); i < len(matchingVersions); i++ {
				if !slices.Contains(existingVersions, matchingVersions[i]) {
					newVersions = append(newVersions, matchingVersions[i])
				}
			}
		}
	}
	return newVersions, nil
}
