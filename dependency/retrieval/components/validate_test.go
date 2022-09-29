package components_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/mri/dependency/retrieval/components"
	"github.com/sclevine/spec"
)

func testDependencyValidation(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		server *httptest.Server
	)
	it.Before(func() {
		var err error

		// Set up tar files
		buffer := bytes.NewBuffer(nil)
		tw := tar.NewWriter(buffer)

		Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
		_, err = tw.Write(nil)
		Expect(err).NotTo(HaveOccurred())

		licenseFile := filepath.Join("some-dir", "LICENSE")
		licenseContent, err := os.ReadFile(filepath.Join("testdata", "LICENSE"))
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.WriteHeader(&tar.Header{Name: licenseFile, Mode: 0755, Size: int64(len(licenseContent))})).To(Succeed())
		_, err = tw.Write(licenseContent)
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.Close()).To(Succeed())

		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodHead {
				http.Error(w, "NotFound", http.StatusNotFound)

				return
			}

			switch req.URL.Path {
			case "/":
				w.WriteHeader(http.StatusOK)
			case "/file.tgz":
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, buffer.String())
			default:
				t.Fatal(fmt.Sprintf("unknown path: %s", req.URL.Path))
			}
		}))
	})
	it.After(func() {
		server.Close()
	})

	context("Validate", func() {
		it("validates the dependency checksum", func() {
			valid, err := components.Validate(components.RubyRelease{
				URL: components.URL{
					Gz: fmt.Sprintf("%s/file.tgz", server.URL),
				},
				SHA256: components.SHA256{
					Gz: "5556fe4667410329990d436f6d1d11395a204978be1211b44cc60c9624909628",
				},
			})

			Expect(err).To(Not(HaveOccurred()))
			Expect(valid).To(BeTrue())
		})

		context("the checksums do not match", func() {
			it("returns an error", func() {
				valid, err := components.Validate(components.RubyRelease{
					URL: components.URL{
						Gz: fmt.Sprintf("%s/file.tgz", server.URL),
					},
					SHA256: components.SHA256{
						Gz: "another hash",
					},
				})

				Expect(err).To(MatchError("failed to validate dependency checksum"))
				Expect(valid).To(BeFalse())
			})
		})

		context("failure cases", func() {
			context("fails to get artifact", func() {
				it("returns an error", func() {
					_, err := components.Validate(components.RubyRelease{
						URL: components.URL{
							Gz: "nonexistent",
						},
						SHA256: components.SHA256{
							Gz: "another hash",
						},
					})
					Expect(err).To(MatchError(ContainSubstring(`failed to get nonexistent`)))
				})
			})
		})
	})
}
