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

func testSimpleApp(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "simple_app"))
			Expect(err).ToNot(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("pack builds and runs the app successfully", func() {
			var logs fmt.Stringer
			var err error

			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.MRI.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.
				WithCommand("ruby run.rb").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(BeAvailable(), logs.String())
			Eventually(container).Should(Serve(MatchRegexp(`Hello from Ruby 3\.1\.\d+`)).OnPort(8080))
		})

		// This test is not currently applicable on jammy because currently Jammy support
		// only applies to one version of Ruby (version 3.1 and above)
		if builder.LocalInfo.Stack.ID != "io.buildpacks.stacks.jammy" {
			context("using an older version of Ruby", func() {
				it("pack builds and runs the app successfully", func() {
					var err error
					var logs fmt.Stringer

					image, logs, err = pack.WithNoColor().Build.
						WithPullPolicy("never").
						WithBuildpacks(
							settings.Buildpacks.MRI.Online,
							settings.Buildpacks.BuildPlan.Online,
						).
						WithEnv(map[string]string{
							"BP_MRI_VERSION": "2.7.*",
							"BP_LOG_LEVEL":   "DEBUG",
						}).
						Execute(name, source)
					Expect(err).ToNot(HaveOccurred(), logs.String)

					container, err = docker.Container.Run.
						WithCommand("ruby run.rb").
						WithEnv(map[string]string{"PORT": "8080"}).
						WithPublish("8080").
						WithPublishAll().
						Execute(image.ID)
					Expect(err).NotTo(HaveOccurred())

					Eventually(container).Should(BeAvailable(), logs.String())
					Eventually(container).Should(Serve(MatchRegexp(`Hello from Ruby 2\.7\.\d+`)).OnPort(8080))
				})
			})
		}

		context("validating SBOM", func() {
			var (
				container2 occam.Container
				sbomDir    string
			)

			it.Before(func() {
				var err error
				sbomDir, err = os.MkdirTemp("", "sbom")
				Expect(err).NotTo(HaveOccurred())
				Expect(os.Chmod(sbomDir, os.ModePerm)).To(Succeed())
			})

			it.After(func() {
				Expect(docker.Container.Remove.Execute(container2.ID)).To(Succeed())
				Expect(os.RemoveAll(sbomDir)).To(Succeed())
			})

			it("writes SBOM files to the layer and label metadata", func() {
				var err error
				var logs fmt.Stringer

				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.MRI.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithEnv(map[string]string{
						"BP_LOG_LEVEL": "DEBUG",
					}).
					WithSBOMOutputDir(sbomDir).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.
					WithCommand("ruby run.rb").
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable(), logs.String())
				Eventually(container).Should(Serve(MatchRegexp(`Hello from Ruby 3\.1\.\d+`)).OnPort(8080))

				Expect(logs).To(ContainLines(
					fmt.Sprintf("  Generating SBOM for /layers/%s/mri", strings.ReplaceAll(settings.Buildpack.ID, "/", "_")),
					MatchRegexp(`      Completed in \d+(\.?\d+)*`),
				))
				Expect(logs).To(ContainLines(
					"  Writing SBOM in the following format(s):",
					"    application/vnd.cyclonedx+json",
					"    application/spdx+json",
					"    application/vnd.syft+json",
				))

				// check that legacy SBOM is included via metadata
				container2, err = docker.Container.Run.
					WithCommand("cat /layers/sbom/launch/sbom.legacy.json").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container2.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(ContainSubstring(`"name":"MRI"`))

				// check that all required SBOM files are present
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "mri", "sbom.cdx.json")).To(BeARegularFile())
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "mri", "sbom.spdx.json")).To(BeARegularFile())
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "mri", "sbom.syft.json")).To(BeARegularFile())

				// check an SBOM file to make sure it has an entry for mri
				contents, err := os.ReadFile(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(settings.Buildpack.ID, "/", "_"), "mri", "sbom.cdx.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring(`"name": "MRI"`))
			})
		})
	})
}
