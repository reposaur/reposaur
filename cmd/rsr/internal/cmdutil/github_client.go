package cmdutil

import (
	"context"
	"net/http"

	"github.com/reposaur/reposaur/pkg/util"
	"github.com/rs/zerolog"
)

type GitHubClientOptions struct {
	// GitHub Personal Access Token
	Token string

	// GitHub App ID
	AppID int64

	// Base64-encoded GitHub App Private Key
	AppPrivateKey string

	// GitHub App Installation ID
	InstallationID int64
}

// NewGithubClient returns a http.Client authenticated to use
// to call the GitHub API. If no authentication information is
// found in environment variables returns a http.DefaultClient.
func NewGitHubClient(ctx context.Context, opts GitHubClientOptions) (*http.Client, error) {
	logger := zerolog.Ctx(ctx)

	if opts.Token != "" {
		return util.NewTokenHTTPClient(ctx, *logger, opts.Token), nil
	}

	if opts.AppID != 0 && opts.InstallationID != 0 && opts.AppPrivateKey != "" {
		logger.Debug().Msg("found environment variables for GitHub App authentication")
		return util.NewInstallationHTTPClient(ctx, *logger, opts.AppID, opts.InstallationID, opts.AppPrivateKey)
	}

	logger.Debug().Msg("using an unauthenticated GitHub client")

	return http.DefaultClient, nil
}
