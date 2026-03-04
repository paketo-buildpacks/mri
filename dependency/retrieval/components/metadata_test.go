package components_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/mri/dependency/retrieval/components"
	"github.com/paketo-buildpacks/mri/dependency/retrieval/components/fakes"
	. "github.com/paketo-buildpacks/occam/matchers"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/sclevine/spec"
)

func testMetadataGeneration(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("GenerateMetadata", func() {
		var (
			release                  components.RubyRelease
			licenseRetriever         *fakes.License
			deprecationDateRetriever *fakes.DeprecationDate
		)

		it.Before(func() {
			licenseRetriever = &fakes.License{}
			licenseRetriever.LookupLicensesCall.Returns.InterfaceSlice = []interface{}{"license-1"}

			deprecationDateRetriever = &fakes.DeprecationDate{}
			deprecationDateRetriever.GetDateCall.Returns.String = "2022-11-01"

			release = components.RubyRelease{
				Version: "3.4.5",
				URL:     components.URL{Gz: "ruby-3.4.5-release.tar.gz"},
				SHA256:  components.SHA256{Gz: "some-ruby-sha"},
			}
		})

		it("retrieves all upstream releases", func() {
			time := time.Date(2022, time.Month(11), 01, 00, 00, 00, 00, time.UTC)
			dependencies, err := components.GenerateMetadata(release, []string{"jammy"}, licenseRetriever, deprecationDateRetriever)
			Expect(err).To(Not(HaveOccurred()))
			Expect(dependencies).To(Equal([]components.Dependency{
				components.Dependency{
					cargo.ConfigMetadataDependency{
						CPE:             "cpe:2.3:a:ruby-lang:ruby:3.4.5:*:*:*:*:*:*:*",
						DeprecationDate: &time,
						PURL:            "pkg:generic/ruby@3.4.5?checksum=some-ruby-sha&download_url=ruby-3.4.5-release.tar.gz",
						ID:              "ruby",
						Name:            "Ruby",
						Licenses:        []interface{}{"license-1"},
						Source:          "ruby-3.4.5-release.tar.gz",
						SourceChecksum:  "sha256:some-ruby-sha",
						Stacks:          []string{"io.buildpacks.stacks.jammy"},
						Version:         "3.4.5",
					},
					"jammy",
				},
			}))
		})

		context("failure cases", func() {
			context("the license retriever returns an error", func() {
				it.Before(func() {
					licenseRetriever.LookupLicensesCall.Returns.Error = errors.New("failed to lookup licenses")
				})
				it("returns an error", func() {
					_, err := components.GenerateMetadata(release, []string{"jammy"}, licenseRetriever, deprecationDateRetriever)
					Expect(err).To(MatchError(ContainSubstring("failed to lookup licenses")))
				})
			})

			context("the deprecation date cannot be retrieved ", func() {
				it.Before(func() {
					deprecationDateRetriever.GetDateCall.Returns.Error = errors.New("failed to get deprecationDate")
				})
				it("returns an error", func() {
					_, err := components.GenerateMetadata(release, []string{"jammy"}, licenseRetriever, deprecationDateRetriever)
					Expect(err).To(MatchError(ContainSubstring("failed to get deprecationDate")))
				})
			})

			context("the deprecation date cannot be parsed as a time ", func() {
				it.Before(func() {
					deprecationDateRetriever.GetDateCall.Returns.String = "bad-time"
				})
				it("returns an error", func() {
					_, err := components.GenerateMetadata(release, []string{"jammy"}, licenseRetriever, deprecationDateRetriever)
					Expect(err).To(MatchError(ContainSubstring("invalid EOL date")))
				})
			})

			context("the version cannot be parsed as semver", func() {
				it.Before(func() {
					release = components.RubyRelease{
						Version: "abc",
						URL:     components.URL{Gz: "ruby-1.2.3-release.tar.gz"},
						SHA256:  components.SHA256{Gz: "some-ruby-sha"},
					}
				})
				it("returns an error", func() {
					_, err := components.GenerateMetadata(release, []string{"jammy"}, licenseRetriever, deprecationDateRetriever)
					Expect(err).To(MatchError(ContainSubstring("Invalid Semantic Version")))
				})
			})
		})
	})

	context("WriteOutput", func() {
		var (
			dependencies []components.Dependency
			outputDir    string
		)

		it.Before(func() {
			outputDir = t.TempDir()
			dependencies = []components.Dependency{
				components.Dependency{
					cargo.ConfigMetadataDependency{
						CPE:            "CPE-1",
						PURL:           "PURL-1",
						ID:             "ruby",
						Name:           "Ruby",
						Licenses:       []interface{}{"license-1"},
						Source:         "source-1",
						SourceChecksum: "source-checksum",
						Stacks:         []string{"stack-1"},
						Version:        "version-1",
					},
					"target-1",
				},
				components.Dependency{
					cargo.ConfigMetadataDependency{
						CPE:            "CPE-2",
						PURL:           "PURL-2",
						ID:             "ruby",
						Name:           "Ruby",
						Licenses:       []interface{}{"license-2"},
						Source:         "source-2",
						SourceChecksum: "source-checksum",
						Stacks:         []string{"stack-2"},
						Version:        "version-2",
					},
					"target-2",
				},
			}
		})

		it("writes dependencies to output file", func() {
			err := components.WriteOutput(filepath.Join(outputDir, "metadata.json"), dependencies)
			Expect(err).To(Not(HaveOccurred()))
			Expect(filepath.Join(outputDir, "metadata.json")).To(BeAFileMatching(`[{"cpe":"CPE-1","purl":"PURL-1","id":"ruby","licenses":["license-1"],"name":"Ruby","source":"source-1","source-checksum":"source-checksum","stacks":["stack-1"],"version":"version-1","target":"target-1"},{"cpe":"CPE-2","purl":"PURL-2","id":"ruby","licenses":["license-2"],"name":"Ruby","source":"source-2","source-checksum":"source-checksum","stacks":["stack-2"],"version":"version-2","target":"target-2"}]
`))
		})

		context("failure cases", func() {
			context("the output file cannot be created", func() {
				it.Before(func() {
					Expect(os.Chmod(outputDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(outputDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := components.WriteOutput(filepath.Join(outputDir, "metadata.json"), dependencies)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
