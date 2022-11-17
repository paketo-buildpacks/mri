package components_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/mri/dependency/retrieval/components"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"
)

func testFindNewVersions(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("FindNewVersions", func() {
		it("returns versions matching constraints and newer than buildpack.toml entries", func() {
			versions, err := components.FindNewVersions("some-dependency",
				cargo.Config{
					Metadata: cargo.ConfigMetadata{
						Dependencies: []cargo.ConfigMetadataDependency{
							cargo.ConfigMetadataDependency{
								ID:      "some-dependency",
								Version: "1.2.3",
								Stacks:  []string{"stack-1"},
							},
							cargo.ConfigMetadataDependency{
								ID:      "some-dependency",
								Version: "1.2.5",
								Stacks:  []string{"stack-1"},
							},
							cargo.ConfigMetadataDependency{
								ID:      "another-dependency",
								Version: "1.2.6",
								Stacks:  []string{"stack-1"},
							},
						},
						DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
							cargo.ConfigMetadataDependencyConstraint{
								Constraint: "1.2.*",
								ID:         "some-dependency",
								Patches:    3,
							},
							cargo.ConfigMetadataDependencyConstraint{
								Constraint: "2.*",
								ID:         "some-dependency",
								Patches:    1,
							},
						},
					},
				},
				[]string{"0.2.3", "1.0.0", "1.1.2", "1.2.2", "1.2.3", "1.2.4", "1.2.4-rc", "1.2.5", "1.2.5-rc", "1.2.6", "2.3.3", "2.3.3-rc", "2.3.4"},
			)

			Expect(err).To(Not(HaveOccurred()))
			Expect(versions).To(Equal([]string{"1.2.4", "1.2.6", "2.3.4"}))
		})

		context("when there are less new versions than allowed patches", func() {
			it("returns all matching versions that are not in buildpack.toml", func() {
				versions, err := components.FindNewVersions("some-dependency",
					cargo.Config{
						Metadata: cargo.ConfigMetadata{
							Dependencies: []cargo.ConfigMetadataDependency{
								cargo.ConfigMetadataDependency{
									ID:      "some-dependency",
									Version: "1.2.3",
									Stacks:  []string{"stack-1"},
								},
							},
							DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
								cargo.ConfigMetadataDependencyConstraint{
									Constraint: "1.2.*",
									ID:         "some-dependency",
									Patches:    3,
								},
							},
						},
					},
					[]string{"0.2.3", "1.0.0", "1.1.2", "1.2.3", "1.2.4"},
				)

				Expect(err).To(Not(HaveOccurred()))
				Expect(versions).To(Equal([]string{"1.2.4"}))
			})
		})

		context("when no constraints match the dependency ID of interest", func() {
			it("returns nothing", func() {
				versions, err := components.FindNewVersions("another-dependency",
					cargo.Config{
						Metadata: cargo.ConfigMetadata{
							Dependencies: []cargo.ConfigMetadataDependency{
								cargo.ConfigMetadataDependency{
									ID:      "some-dependency",
									Version: "1.2.3",
									Stacks:  []string{"stack-1"},
								},
							},
							DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
								cargo.ConfigMetadataDependencyConstraint{
									Constraint: "1.2.*",
									ID:         "some-dependency",
									Patches:    3,
								},
							},
						},
					},
					[]string{"0.2.3", "1.0.0", "1.1.2", "1.2.3", "1.2.4"},
				)
				Expect(err).To(Not(HaveOccurred()))
				Expect(versions).To(Equal([]string{}))
			})
		})

		context("when the buildpack.toml already has the latest dependencies", func() {
			it("returns nothing", func() {
				versions, err := components.FindNewVersions("some-dependency",
					cargo.Config{
						Metadata: cargo.ConfigMetadata{
							Dependencies: []cargo.ConfigMetadataDependency{
								cargo.ConfigMetadataDependency{
									ID:      "some-dependency",
									Version: "1.2.4",
									Stacks:  []string{"stack-1"},
								},
							},
							DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
								cargo.ConfigMetadataDependencyConstraint{
									Constraint: "1.2.*",
									ID:         "some-dependency",
									Patches:    1,
								},
							},
						},
					},
					[]string{"1.2.0", "1.2.1", "1.2.2", "1.2.3", "1.2.4"},
				)
				Expect(err).To(Not(HaveOccurred()))
				Expect(versions).To(Equal([]string{}))
			})
		})

		context("failure cases", func() {
			context("the constraint cannot be converted into a semver constraint", func() {
				it("returns an error", func() {
					_, err := components.FindNewVersions("some-dependency",
						cargo.Config{
							Metadata: cargo.ConfigMetadata{
								DependencyConstraints: []cargo.ConfigMetadataDependencyConstraint{
									cargo.ConfigMetadataDependencyConstraint{
										Constraint: "bad-constraint",
										ID:         "some-dependency",
										Patches:    1,
									},
								},
							},
						},
						[]string{"1.2.3"},
					)
					Expect(err).To(MatchError(ContainSubstring("improper constraint")))
				})
			})
		})
	})
}
