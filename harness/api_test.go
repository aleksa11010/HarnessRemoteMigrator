package harness

import (
	"testing"
)

func TestGetServiceManifestStoreType_GitHub(t *testing.T) {
	out := GetServiceManifestStoreType("GitHub")
	if out != "GitHub" {
		t.Fatalf("Connector type should be GitHub instead of %s", out)
	}
}

func TestGetServiceManifestStoreType_GitLab(t *testing.T) {
	out := GetServiceManifestStoreType("Gitlab")
	if out != "GitLab" {
		t.Fatalf("Connector type should be GitLab instead of %s", out)
	}
}
