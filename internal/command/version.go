package command

import (
	"runtime/debug"
	"strings"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
)

type VersionPayload struct {
	CLI    VersionBlock `json:"cli"`
	Client VersionBlock `json:"client"`
}

type VersionBlock struct {
	Version string `json:"version"`
	GitSHA  string `json:"git_sha"`
}

// Version represents command that builds the version payload for the CLI and its Go client library.
type Version struct {
	buildInfoReader func() (*debug.BuildInfo, bool)
}

func NewVersion() (*Version, error) {
	return &Version{buildInfoReader: debug.ReadBuildInfo}, nil
}

// Run builds the version payload for the CLI and its Go client library.
func (command *Version) Run(cliGitTag string, cliGitSHA string) (VersionPayload, error) {
	info, ok := command.buildInfoReader()
	if !ok || info == nil {
		return VersionPayload{
			CLI: VersionBlock{
				Version: cliGitTag,
				GitSHA:  cliGitSHA,
			},
			Client: VersionBlock{
				Version: "unknown",
				GitSHA:  "unknown",
			},
		}, nil
	}

	clientVersion, clientGitSHA := clientGoVersionFromBuildInfo(info)

	return VersionPayload{
		CLI: VersionBlock{
			Version: cliGitTag,
			GitSHA:  cliGitSHA,
		},
		Client: VersionBlock{
			Version: clientVersion,
			GitSHA:  clientGitSHA,
		},
	}, nil
}

// clientGoVersionFromBuildInfo extracts the version and git SHA from the build info of the Go client library.
func clientGoVersionFromBuildInfo(info *debug.BuildInfo) (version string, gitSHA string) {
	for _, dep := range info.Deps {
		if dep == nil || dep.Path != constant.ClientGoModulePath {
			continue
		}

		mod := dep
		if dep.Replace != nil {
			mod = dep.Replace
		}

		v := mod.Version
		if v == "" {
			v = "unknown"
		}

		return v, gitSHAFromGoModuleVersion(v)
	}

	return "unknown", "unknown"
}

// gitSHAFromGoModuleVersion extracts the commit SHA from a Go pseudo-version:
// vX.Y.Z-yyyymmddhhmmss-<12+ hex>.
// If it's a tag like v0.2.0, there's no SHA embedded and we return "unknown".
func gitSHAFromGoModuleVersion(v string) string {
	parts := strings.Split(v, "-")
	if len(parts) < 3 {
		return "unknown"
	}

	sha := parts[len(parts)-1]
	if sha == "" {
		return "unknown"
	}

	for _, r := range sha {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		default:
			return "unknown"
		}
	}

	return sha
}
