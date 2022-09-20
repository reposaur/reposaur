package cmdutil

import (
	githubclient "github.com/reposaur/reposaur/provider/github/client"
	"github.com/spf13/pflag"
)

type GitHubClientOptions struct {
	// GitHub API Base URL
	BaseURL string

	// GitHub Personal Access Token
	Token string

	// GitHub App ID
	AppID int64

	// GitHub App Private Key
	AppPrivateKey string

	// GitHub App Installation ID
	InstallationID int64
}

type GitLabClientOptions struct {
	// GitLab API Base URL
	BaseURL string

	// GitLab Personal Access Token
	Token string
}

func AddPolicyPathsFlag(flags *pflag.FlagSet, p *[]string) {
	flags.StringSliceVarP(p, "policy", "p", []string{"."}, "path to policy files or directories")
}

func AddOutputFlag(flags *pflag.FlagSet, p *string) {
	flags.StringVarP(p, "output", "o", "-", "output filename")
}

func AddTraceFlag(flags *pflag.FlagSet, p *bool) {
	flags.BoolVarP(p, "trace", "t", false, "enable tracing")
}

func AddVerboseFlag(flags *pflag.FlagSet, p *bool) {
	flags.BoolVarP(p, "verbose", "v", false, "print debug logs")
}

func AddExperimentalFlag(flags *pflag.FlagSet, p *bool) {
	flags.BoolVar(p, "experimental", false, "accepts the usage of experimental features")
}

func AddGitHubFlags(flags *pflag.FlagSet, p *GitHubClientOptions) {
	var (
		defURL            = getEnv("GH_API_URL", "GITHUB_API_URL")
		defToken          = getEnv("GH_TOKEN", "GITHUB_TOKEN")
		defAppID          = getInt64Env("GH_APP_ID", "GITHUB_APP_ID")
		defAppPrivKey     = getEnv("GH_APP_PRIVATE_KEY", "GITHUB_APP_PRIVATE_KEY")
		defInstallationID = getInt64Env("GH_INSTALLATION_ID", "GITHUB_INSTALLATION_ID")
	)

	if defURL == "" {
		defURL = githubclient.DefaultBaseURL
	}

	flags.StringVar(&p.BaseURL, "github-api-url", defURL, "base url GitHub API")
	flags.StringVar(&p.Token, "github-token", defToken, "token for GitHub")
	flags.Int64Var(&p.AppID, "github-app-id", defAppID, "id for GitHub App")
	flags.StringVar(&p.AppPrivateKey, "github-app-private-key", defAppPrivKey, "base64-encoded private key for GitHub App")
	flags.Int64Var(&p.InstallationID, "github-installation-id", defInstallationID, "installation ID for GitHub App")
}

func AddGitLabFlags(flags *pflag.FlagSet, p *GitLabClientOptions) {
	var (
		defURL   = getEnv("GL_API_URL", "GITLAB_API_URL")
		defToken = getEnv("GL_TOKEN", "GITLAB_TOKEN")
	)

	// if defURL == "" {
	// 	defURL = gitlabclient.DefaultBaseURL
	// }

	flags.StringVar(&p.BaseURL, "gitlab-api-url", defURL, "base url GitLab API")
	flags.StringVar(&p.Token, "gitlab-token", defToken, "token for GitLab")
}
