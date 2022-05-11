package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testLogging(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image occam.Image

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("logs useful information for the user", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.MRI.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{
					"BP_MRI_VERSION": "2.7.x",
				}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
				"  Resolving MRI version",
				"    Candidate version sources (in priority order):",
				"      BP_MRI_VERSION -> \"2.7.x\"",
				"      <unknown>      -> \"\"",
			))

			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected MRI version \(using BP_MRI_VERSION\): 2\.7\.\d+`),
			))

			Expect(logs).To(ContainLines(
				"  Executing build process",
				MatchRegexp(`    Installing MRI 2\.7\.\d+`),
				MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
			))

			Expect(logs).To(ContainLines(
				"  Configuring build environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.7\.\d+:/layers/%s/mri/lib/ruby/gems/2\.7\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
			))

			Expect(logs).To(ContainLines(
				"  Configuring launch environment",
				MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.7\.\d+:/layers/%s/mri/lib/ruby/gems/2\.7\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
			))
		})

		context("when the BP_LOG_LEVEL env var is set to DEBUG", func() {
			it("logs denote the logger is set to DEBUG level", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "simple_app"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.MRI.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithEnv(map[string]string{
						"BP_MRI_VERSION": "2.7.x",
						"BP_LOG_LEVEL":   "DEBUG",
					}).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
					"  Resolving MRI version",
					"    Candidate version sources (in priority order):",
					"      BP_MRI_VERSION -> \"2.7.x\"",
					"      <unknown>      -> \"\"",
				))

				Expect(logs).To(ContainLines(
					MatchRegexp(`    Selected MRI version \(using BP_MRI_VERSION\): 2\.7\.\d+`),
				))

				Expect(logs).To(ContainLines(
					"  Getting the layer associated with MRI:",
					"    /layers/paketo-buildpacks_mri/mri",
				))

				Expect(logs).To(ContainLines(
					"  Executing build process",
					MatchRegexp(`    Installing MRI 2\.7\.\d+`),
					"    Installation path: /layers/paketo-buildpacks_mri/mri",
					MatchRegexp(`    Source URI\: https\:\/\/deps\.paketo\.io\/ruby\/ruby_2\.7\.\d+_linux_x64_bionic_.*\.tgz`),
					MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				))

				Expect(logs).To(ContainLines(
					"  Adding /layers/paketo-buildpacks_mri/mri/bin to the $PATH",
				))

				Expect(logs).To(ContainLines(
					"  Configuring build environment",
					MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.7\.\d+:/layers/%s/mri/lib/ruby/gems/2\.7\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				))

				Expect(logs).To(ContainLines(
					"  Configuring launch environment",
					MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.7\.\d+:/layers/%s/mri/lib/ruby/gems/2\.7\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				))
			})
		})

		context("when the app contains a buildpack.yml", func() {
			it("logs that the buildpack.yml is deprecated", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "simple_app"))
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(source, "buildpack.yml"), []byte(`{ "mri": { "version": "2.7.x" } }`), 0600)
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.MRI.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, settings.Buildpack.Name)),
					"  Resolving MRI version",
					"    Candidate version sources (in priority order):",
					"      buildpack.yml -> \"2.7.x\"",
					"      <unknown>     -> \"\"",
				))

				Expect(logs).To(ContainLines(
					MatchRegexp(`    Selected MRI version \(using buildpack\.yml\): 2\.7\.\d+`),
				))

				Expect(logs).To(ContainLines(
					"    WARNING: Setting the MRI version through buildpack.yml will be deprecated soon in MRI Buildpack v2.0.0.",
					"    Please specify the version through the $BP_MRI_VERSION environment variable instead. See README.md for more information.",
				))

				Expect(logs).To(ContainLines(
					"  Executing build process",
					MatchRegexp(`    Installing MRI 2\.7\.\d+`),
					MatchRegexp(`      Completed in ([0-9]*(\.[0-9]*)?[a-z]+)+`),
				))

				Expect(logs).To(ContainLines(
					"  Configuring build environment",
					MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.7\.\d+:/layers/%s/mri/lib/ruby/gems/2\.7\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				))

				Expect(logs).To(ContainLines(
					"  Configuring launch environment",
					MatchRegexp(fmt.Sprintf(`    GEM_PATH -> "/home/cnb/.gem/ruby/2\.7\.\d+:/layers/%s/mri/lib/ruby/gems/2\.7\.\d+"`, strings.ReplaceAll(settings.Buildpack.ID, "/", "_"))),
				))
			})
		})
	})
}
