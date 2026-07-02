package version

import (
	"regexp"
	"testing"
)

func TestVersionNonEmpty(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must not be empty")
	}
}

func TestVersionSemver(t *testing.T) {
	matched, err := regexp.MatchString(`^\d+\.\d+\.\d+$`, Version)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Fatalf("Version %q does not match semver pattern", Version)
	}
}
