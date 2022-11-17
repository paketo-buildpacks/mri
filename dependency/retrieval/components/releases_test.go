package components_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/mri/dependency/retrieval/components"
	"github.com/sclevine/spec"
)

func testReleaseFetcher(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("ReleaseFetcher", func() {
		var (
			releaseFetcher components.ReleaseFetcher
			server         *httptest.Server
		)

		it.Before(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodHead {
					http.Error(w, "NotFound", http.StatusNotFound)
					return
				}

				switch req.URL.Path {
				case "":
					w.WriteHeader(http.StatusOK)
				case "/releases":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `
- version: 1.2.3
  url:
    gz:  ruby-1.2.3.tar.gz
    zip: ruby-1.2.3.zip
    xz:  ruby-1.2.3.tar.xz
  sha256:
    gz:  1.2.3-gz-sha
    zip: 1.2.3-zip-sha
    xz:  1.2.3-xz-sha

- version: 1.3.4
  url:
    gz:  ruby-1.3.4.tar.gz
    zip: ruby-1.3.4.zip
    xz:  ruby-1.3.4.tar.xz
  sha256:
    gz:  1.3.4-gz-sha
    zip: 1.3.4-zip-sha
    xz:  1.3.4-xz-sha

- version: 2.0.0
  url:
    gz:  ruby-2.0.0.tar.gz
    zip: ruby-2.0.0.zip
    xz:  ruby-2.0.0.tar.xz
  sha256:
    gz:  2.0.0-gz-sha
    zip: 2.0.0-zip-sha
    xz:  2.0.0-xz-sha`)

				case "/bad-endpoint":
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintln(w, `bad endpoint`)
				case "/bad-content":
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `{versions: [1.2.3]}`)
				default:
					t.Fatalf("unknown path: %s", req.URL.Path)
				}
			}))

		})

		it.After(func() {
			server.Close()
		})

		context("GetUpstreamReleases", func() {
			context("mock server tests for fetcher parsing logic", func() {
				it.Before(func() {
					releaseFetcher = components.NewReleaseFetcher(fmt.Sprintf("%s/releases", server.URL))
				})
				it("retrieves all upstream releases", func() {
					releases, err := releaseFetcher.GetUpstreamReleases()
					Expect(err).To(Not(HaveOccurred()))
					Expect(releases).To(Equal(map[string]components.RubyRelease{
						"1.2.3": {
							Version: "1.2.3",
							URL: components.URL{
								Gz: "ruby-1.2.3.tar.gz",
							},
							SHA256: components.SHA256{
								Gz: "1.2.3-gz-sha",
							},
						},
						"1.3.4": {
							Version: "1.3.4",
							URL: components.URL{
								Gz: "ruby-1.3.4.tar.gz",
							},
							SHA256: components.SHA256{
								Gz: "1.3.4-gz-sha",
							},
						},
						"2.0.0": {
							Version: "2.0.0",
							URL: components.URL{
								Gz: "ruby-2.0.0.tar.gz",
							},
							SHA256: components.SHA256{
								Gz: "2.0.0-gz-sha",
							},
						},
					}))
				})

				context("failure cases", func() {
					context("version feed endpoint cannot be retrieved", func() {
						it.Before(func() {
							releaseFetcher = components.NewReleaseFetcher("invalid URL")
						})
						it("returns an error", func() {
							_, err := releaseFetcher.GetUpstreamReleases()
							Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
						})
					})

					context("endpoint returns a bad status code", func() {
						it.Before(func() {
							releaseFetcher = components.NewReleaseFetcher(fmt.Sprintf("%s/bad-endpoint", server.URL))
						})
						it("returns an error", func() {
							_, err := releaseFetcher.GetUpstreamReleases()
							Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("failed to query %s/bad-endpoint: 500", server.URL))))
						})
					})

					context("the endpoint cannot be YAML parsed", func() {
						it.Before(func() {
							releaseFetcher = components.NewReleaseFetcher(fmt.Sprintf("%s/bad-content", server.URL))
						})
						it("returns an error", func() {
							_, err := releaseFetcher.GetUpstreamReleases()
							Expect(err).To(MatchError(ContainSubstring("cannot unmarshal")))
						})
					})
				})
			})

			context("real Ruby server test", func() {
				it.Before(func() {
					releaseFetcher = components.NewReleaseFetcher("https://raw.githubusercontent.com/ruby/www.ruby-lang.org/master/_data/releases.yml")
				})
				it("retrieves all upstream releases", func() {
					releases, err := releaseFetcher.GetUpstreamReleases()
					Expect(err).To(Not(HaveOccurred()))
					Expect(releases).To(Not(BeEmpty()))
					Expect(releases["2.7.3"]).To(Equal(
						components.RubyRelease{
							Version: "2.7.3",
							URL: components.URL{
								Gz: "https://cache.ruby-lang.org/pub/ruby/2.7/ruby-2.7.3.tar.gz",
							},
							SHA256: components.SHA256{
								Gz: "8925a95e31d8f2c81749025a52a544ea1d05dad18794e6828709268b92e55338",
							},
						}))
				})

				context("failure cases", func() {
					context("version feed endpoint cannot be retrieved", func() {
						it.Before(func() {
							releaseFetcher = components.NewReleaseFetcher("invalid URL")
						})
						it("returns an error", func() {
							_, err := releaseFetcher.GetUpstreamReleases()
							Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
						})
					})

					context("endpoint returns a bad status code", func() {
						it.Before(func() {
							releaseFetcher = components.NewReleaseFetcher(fmt.Sprintf("%s/bad-endpoint", server.URL))
						})
						it("returns an error", func() {
							_, err := releaseFetcher.GetUpstreamReleases()
							Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("failed to query %s/bad-endpoint: 500", server.URL))))
						})
					})

					context("the endpoint cannot be YAML parsed", func() {
						it.Before(func() {
							releaseFetcher = components.NewReleaseFetcher(fmt.Sprintf("%s/bad-content", server.URL))
						})
						it("returns an error", func() {
							_, err := releaseFetcher.GetUpstreamReleases()
							Expect(err).To(MatchError(ContainSubstring("cannot unmarshal")))
						})
					})
				})
			})
		})
	})
}
