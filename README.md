# Reposaur

<a href="https://pkg.go.dev/github.com/reposaur/reposaur?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>

> Audit your GitHub data using custom policies written in [Rego][rego].

## Features

* Write custom policies using OPA Rego language
* Simple and easy-to-use CLI
* Extendable using the Go SDK
* Output reports in JSON and SARIF formats
* (TODO) Deploy as a GitHub App or use in GitHub Actions

## Installation

```shell
$ go install github.com/reposaur/reposaur
```

Or

```go
import "github.com/reposaur/reposaur/pkg/sdk"
```

## Gudies

- [Writing & testing your first policy](https://github.com/orgs/reposaur/discussions/1)

## Usage

```bash
$ reposaur --help
Executes a set of Rego policies against the data provided

Usage:
  reposaur [flags]

Aliases:
  reposaur, c

Flags:
  -f, --format string      report output format (one of 'json' and 'sarif') (default "sarif")
  -h, --help               help for reposaur
  -n, --namespace string   use this namespace
  -p, --policy strings     set the path to a policy or directory of policies (default [./policy])
```

### Feeding data from `gh` CLI

```bash
$ gh api /repos/reposaur/reposaur | reposaur
```

### Preparing a SARIF report to send to GitHub

```bash
$ gh api /repos/reposaur/reposaur | reposaur | gzip | base64
```

## Go SDK

### Simple usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/reposaur/reposaur/pkg/sdk"
)

var data = map[string]interface{}{
	"repository":   map[string]interface{}{},
	"pull_request": map[string]interface{}{},
}

func main() {
	ctx := context.Background()

	rs, err := sdk.New(ctx, []string{"./policy"})
	if err != nil {
		panic(err)
	}

	report, err := rs.Check(ctx, data)
	if err != nil {
		panic(err)
	}

	fmt.Println(report)
}
```

## Policies

The policies used by Reposaur use the [Rego][rego] language. There are some outstanding
particularities related with the policies' namespaces, rules, metadata and built-in functions.

### Namespaces

Reposaur can execute multiple policies against different kinds of data. To distinguish
which policies should be executed against a particular set of data we use namespaces.

For example, if you've a policy like following:

```rego
package repository

violation_empty_description {
  not input.description
}
```

and execute the command (notice we're requesting a Pull Request):

```
gh api /repos/reposaur/reposaur/pulls/1 | reposaur
```

the generated report would be empty since there's no policy with the `pull_request` namespace.

By default, the CLI attempts to detect the namespace of some data based on it's properties. If
it's failing to detect a valid namespace, you can always specify it manually using the `--namespace <NAMESPACE>` flag.

## Rules

Reposaur will only query the rules that have the following prefixes (aka "kinds"):

#### `violation_`, `fail_`, `error_`

Cause the CLI to exit with code `1`, the results in the SARIF report will have the `error` level.

#### `warn_` 

Cause the CLI to exit with code `0`, the results in the SARIF report will have the `warning` level.

#### `note_`, `info_`

Cause the CLI to exit with code `0`, the results in the SARIF report will have the `note` level.

## Metadata

Your rules can be enhanced with additional information that will be added in the final report, independently of the output format.
The example below includes all possible metadata fields:

```rego
# METADATA
# title: Forking is enabled
# description: >
#   The repository has forking enabled, which means any member of the organization could
#   fork it to their own account and change it's visibility to be _public_.
#
#   ### Fix
#
#   1. Go to the repository's settings
#
#   3. Uncheck the "Allow forking" option
# custom:
#   tags: [security]
#   security-severity: 9
violation_forking_enabled {
	input.allow_forking
}
```

The above rule would be represented in the SARIF report as follows:

```json
{
  "id": "repository/violation/forking_enabled",
  "name": "Forking is enabled",
  "shortDescription": {
    "text": "Forking is enabled"
  },
  "fullDescription": {
    "text": "The repository has forking enabled, which means any member of the organization could fork it to their own account and change it's visibility to be _public_.\n### Fix\n1. Go to the repository's settings\n3. Uncheck the \"Allow forking\" option\n",
    "markdown": "The repository has forking enabled, which means any member of the organization could fork it to their own account and change it's visibility to be _public_.\n### Fix\n1. Go to the repository's settings\n3. Uncheck the \"Allow forking\" option\n"
  },
  "help": {
    "markdown": "The repository has forking enabled, which means any member of the organization could fork it to their own account and change it's visibility to be _public_.\n### Fix\n1. Go to the repository's settings\n3. Uncheck the \"Allow forking\" option\n"
  },
  "properties": {
    "security-severity": "9",
    "tags": [
      "security"
    ]
  }
}
```

## Built-in Functions

### `github.request`

Does an HTTP request against the GitHub REST API. Usage is similar to the
Octokit.js library, for example:

```rego
resp := github.request("GET /repos/{owner}/{repo}/branches/{branch}/protection", {
	"owner": input.owner.login,
	"repo": input.name,
	"branch": input.default_branch,
})
```

The response will include the following properties:

* `body` - The HTTP Response body
* `error` - Any error that has occurred during the call
* `statusCode` - The HTTP Response status code

### `github.graphql`

Does an HTTP request against the GitHub GraphQL API. For example:

```rego
resp := github.graphql(
	`
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name) { 
				name
			}
		}
	`,
	{
		"owner": input.owner.login,
		"name": input.name,
	},
)
```

The response will include the following properties:

* `body` - The HTTP Response body
* `error` - Any error that has occurred during the call
* `statusCode` - The HTTP Response status code

## Contributing

We appreciate every contribution, thanks for considering it!

- [Open an issue][issues] if you have a problem or found a bug
- [Open a Pull Request][pulls] if you have a suggestion, improvement or bug fix
- [Open a Discussion][discussions] if you have questions or want to discuss ideas

[issues]: https://github.com/reposaur/reposaur/issues
[pulls]: https://github.com/reposaur/reposaur/pulls
[discussions]: https://github.com/orgs/reposaur/discussions

## License

This project is released under the [MIT License](LICENSE).

[rego]: https://www.openpolicyagent.org/docs/latest/policy-language/
