package builtins

type GitHubResponse struct {
	StatusCode int         `json:"status"`
	Body       interface{} `json:"body"`
	Error      string      `json:"error"`
}
