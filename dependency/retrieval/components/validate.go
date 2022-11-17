package components

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

func Validate(release RubyRelease) (bool, error) {
	archiveResponse, err := http.Get(release.URL.Gz)
	if err != nil {
		return false, fmt.Errorf("failed to get %s: %w", release.URL.Gz, err)
	}
	defer archiveResponse.Body.Close()

	vr := cargo.NewValidatedReader(archiveResponse.Body, release.SHA256.Gz)
	valid, err := vr.Valid()
	if err != nil {
		return false, err
	}
	if !valid {
		return false, errors.New("failed to validate dependency checksum")
	}
	return true, nil
}
