#provider #integration

## Schemas

### `github.issue`
Represents a GitHub Issue. See the [JSON Schema]().

### `github.organization`
Represents a GitHub Organization. See the [JSON Schema]().

### `github.pull_request`
Represents a GitHub Pull Request. See the [JSON Schema]().

### `github.release`
Represents a GitHub Release. See the [JSON Schema]().

### `github.repository`
Represents a GitHub Repository. See the [JSON Schema]().

### `github.user`
Represents a GitHub User. See the [JSON Schema]().

## Built-in Functions

### `github.graphql`

Executes an HTTP request against the [GitHub GraphQL API](https://docs.github.com/en/graphql) and returns the status code and response body.

#### Example
```rego
query := `
	query($owner: String!, $name: String!) {
		repository(owner: $owner, name: $name) {
			name
		}
	}
`

vars := {
	"owner": "reposaur",
	"name": "reposaur",
}

response := github.graphql(query, vars)
response.status == 200
response.body.data.repository.name == "reposaur"
```

### `github.request`

Executes an HTTP request against the [GitHub REST API](https://docs.github.com/en/rest) and returns the status code and response body.

Usage is similar to the [Octokit.js](https://github.com/octokit/octokit.js) library and most documentation examples can be translated 1-1 to `github.request`.

#### Example

```rego
response := github.request("GET /repos/{owner}/{repo}", {
	"owner": "reposaur",
	"repo": "reposaur",
})

response.status == 200
response.body.name == "reposaur"
```