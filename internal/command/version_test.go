package command

import (
	"runtime/debug"
	"testing"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
)

func TestGitSHAFromGoModuleVersion(t *testing.T) {
	t.Parallel()

	t.Run("tag_has_no_sha", func(t *testing.T) {
		t.Parallel()
		if got := gitSHAFromGoModuleVersion("v0.2.0"); got != "unknown" {
			t.Fatalf("expected unknown, got %q", got)
		}
	})

	t.Run("pseudo_version_extracts_sha", func(t *testing.T) {
		t.Parallel()
		v := "v0.2.0-20260520101010-abcdef123456"
		if got := gitSHAFromGoModuleVersion(v); got != "abcdef123456" {
			t.Fatalf("expected %q, got %q", "abcdef123456", got)
		}
	})

	t.Run("invalid_suffix_returns_unknown", func(t *testing.T) {
		t.Parallel()
		v := "v0.2.0-20260520101010-abcxyz"
		if got := gitSHAFromGoModuleVersion(v); got != "unknown" {
			t.Fatalf("expected unknown, got %q", got)
		}
	})
}

func TestClientGoVersionFromBuildInfo(t *testing.T) {
	t.Parallel()

	t.Run("missing_dep_returns_unknowns", func(t *testing.T) {
		t.Parallel()
		info := &debug.BuildInfo{Deps: []*debug.Module{}}
		version, sha := clientGoVersionFromBuildInfo(info)
		if version != "unknown" || sha != "unknown" {
			t.Fatalf("expected unknown/unknown, got %q/%q", version, sha)
		}
	})

	t.Run("uses_dep_version", func(t *testing.T) {
		t.Parallel()
		info := &debug.BuildInfo{
			Deps: []*debug.Module{
				{Path: constant.ClientGoModulePath, Version: "v0.2.0"},
			},
		}

		version, sha := clientGoVersionFromBuildInfo(info)
		if version != "v0.2.0" {
			t.Fatalf("expected version %q, got %q", "v0.2.0", version)
		}

		if sha != "unknown" {
			t.Fatalf("expected sha %q, got %q", "unknown", sha)
		}
	})

	t.Run("uses_replace_when_present", func(t *testing.T) {
		t.Parallel()
		info := &debug.BuildInfo{
			Deps: []*debug.Module{
				{
					Path:    constant.ClientGoModulePath,
					Version: "v0.2.0",
					Replace: &debug.Module{
						Path:    constant.ClientGoModulePath,
						Version: "v0.2.0-20260520101010-abcdef123456",
					},
				},
			},
		}

		version, sha := clientGoVersionFromBuildInfo(info)
		if version != "v0.2.0-20260520101010-abcdef123456" {
			t.Fatalf("expected version %q, got %q", "v0.2.0-20260520101010-abcdef123456", version)
		}

		if sha != "abcdef123456" {
			t.Fatalf("expected sha %q, got %q", "abcdef123456", sha)
		}
	})
}

func TestVersion_Run(t *testing.T) {
	t.Parallel()

	t.Run("no_build_info", func(t *testing.T) {
		t.Parallel()

		cmd, err := NewVersion()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		cmd.buildInfoReader = func() (*debug.BuildInfo, bool) { return nil, false }
		out, err := cmd.Run("v1.2.3", "deadbeef")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if out.CLI.Version != "v1.2.3" || out.CLI.GitSHA != "deadbeef" {
			t.Fatalf("unexpected cli block: %#v", out.CLI)
		}

		if out.Client.Version != "unknown" || out.Client.GitSHA != "unknown" {
			t.Fatalf("unexpected client block: %#v", out.Client)
		}
	})

	t.Run("build_info_present", func(t *testing.T) {
		t.Parallel()

		cmd, err := NewVersion()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		cmd.buildInfoReader = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Deps: []*debug.Module{
					{
						Path:    constant.ClientGoModulePath,
						Version: "v0.2.0-20260520101010-abcdef123456",
					},
				},
			}, true
		}

		out, err := cmd.Run("v9.9.9", "cafebabe")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if out.CLI.Version != "v9.9.9" || out.CLI.GitSHA != "cafebabe" {
			t.Fatalf("unexpected cli block: %#v", out.CLI)
		}

		if out.Client.Version != "v0.2.0-20260520101010-abcdef123456" || out.Client.GitSHA != "abcdef123456" {
			t.Fatalf("unexpected client block: %#v", out.Client)
		}
	})
}
