package cmdutil

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
