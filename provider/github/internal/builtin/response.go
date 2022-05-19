package builtin

type response struct {
	StatusCode int         `json:"status"`
	Body       interface{} `json:"body"`
}
