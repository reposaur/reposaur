# Reposaur

[![Go Reference](https://pkg.go.dev/badge/github.com/reposaur/reposaur.svg)](https://pkg.go.dev/github.com/reposaur/reposaur)
[![Go Report Card](https://goreportcard.com/badge/github.com/reposaur/reposaur)](https://goreportcard.com/report/github.com/reposaur/reposaur)
[![Tests](https://img.shields.io/github/workflow/status/reposaur/reposaur/Test?longCache=true&label=Test&logo=github%20actions&logoColor=fff)](https://github.com/reposaur/reposaur/actions/workflows/test.yml)

> Audit your GitHub data using custom policies written in [Rego][rego].

Reposaur allows users and organizations to execute policies against GitHub data to
generate reports, perform auditing and more. The end-goal is to make it easy to
perform such tasks in an automated way, enabling people to focus their time in solving
any issues instead of looking for them ([see the examples](#examples)).

:warning: While version `v1.0.0` is not released, consider the API to be _unstable_. The CLI and its output
will most probably not change much, except for the fact that it might get a `test` command before the first
stable release.

# Features

* [x] Write custom policies using [Rego][rego] policy language ([see more](#policies))
* [x] Simple, composable and easy-to-use CLI ([see more](#examples))
* [x] Extendable using the Go SDK
* [x] Output reports in JSON and SARIF formats
* [x] Use in GitHub Actions ([see more](#use-in-github-actions))
* [ ] Policies unit testing (possible with `opa test` if not using built-in functions) (see reposaur/reposaur#1)
* [ ] Deploy as a GitHub App (possible but no official guide yet) (see reposaur/reposaur#2)

# Installation

#### Using Script

```shell
$ curl -o- https://raw.githubusercontent.com/reposaur/reposaur/main/install.sh | bash
```

#### Using Go

```shell
$ go install github.com/reposaur/reposaur
```

# Guides

- [Writing & testing your first policy](https://github.com/orgs/reposaur/discussions/1)

# Usage

```bash
$ reposaur --help
Executes a set of Rego policies against the data provided

Usage:
  reposaur [flags]

Flags:
  -f, --format string      report output format (one of 'json' and 'sarif') (default "sarif")
  -h, --help               help for reposaur
  -n, --namespace string   use this namespace
  -p, --policy strings     set the path to a policy or directory of policies (default [./policy])
```

# Examples

The following examples assume we're running Reposaur in a directory
with the following policies inside `./policy` directory. If you're looking 
for more examples of policies check the [reposaur/policy](https://github.com/reposaur/policy) repository.

_./policy/repository.go_
```rego
package repository

# METADATA
# title: Repository description is empty
# description: >
#   It's important that repositoryies have a short but
#   meaningful description. A description helps other people
#   finding the repository more easily and understanding what
#   the repository is all about.
warn_description_empty {
	input.description == null
}
```

_./policy/pull_request.go_
```rego
package pull_request

# METADATA
# title: Pull Request title is malformed
# description: Pull Request titles must follow [Conventional Commits](https://www.conventionalcommits.org) guidelines.
warn_title_malformed {
	not regex.match("(?i)^(\\w+)(\\(.*\\))?:.*", input.title)
}
```

_./policy/organization.go_
```rego
package organization

# METADATA
# title: Organization has 2FA requirement disabled
# description: Organization doesn't require members to have 2FA enabled
warn_two_factor_requirement_disabled {
	input.two_factor_requirement_enabled == false
}
```

## Executing the policies against a single repository

```shell
$ gh api /repos/reposaur/reposaur | reposaur
# { ... }
```

## Executing the policies against every repository in an organization

```shell
$ gh api /orgs/reposaur/repos --paginate | reposaur
# [{ ... }, ...]
```

## Executing the policies against an organization

```shell
$ gh api /orgs/reposaur | reposaur
# { ... }
```

## Uploading a SARIF report to GitHub

```shell
$ report=$(gh api /repos/reposaur/reposaur | reposaur | gzip | base64)

$ gh api /repos/reposaur/reposaur/code-scanning/sarifs \
    -f sarif="$report" \
    -f commit_sha="..." \
    -f ref="..."
```

## Uploading a SARIF report for multiple repositories

```shell
$ gh api /orgs/reposaur/repos --paginate \
  | reposaur \
  | jq -r '.[] | @base64' \
  | {
    while read r; do
      _r() {
        echo ${r} | base64 -d | jq -r ${1}
      }

      owner=$(_r '.runs[0].properties.owner')
      repo=$(_r '.runs[0].properties.repo')
      branch=$(_r '.runs[0].properties.default_branch')
      commit_sha=$(gh api "/repos/$owner/$repo/branches/$branch" | jq -r '.commit.sha')

      gh api /repos/reposaur/reposaur/code-scanning/sarifs \
        -f sarif="$(_r '.' | gzip | base64)" \
        -f commit_sha="$commit_sha" \
        -f ref="refs/heads/$branch"
    done
  }
```

# Policies

Policies are written in [Rego][rego]. There are some particularities that
Reposaur takes into consideration, detailed below.

## Namespaces

Reposaur can execute multiple policies against different kinds of data. To distinguish
which policies should be executed against a particular set of data we use namespaces.

A namespace is defined using the `package` keyword:

```rego
# This policy has the "repository" namespace
package repository
```

By default, the CLI attempts to detect the namespace based on the data. If
it's failing to detect a valid namespace, you can specify it manually using the `--namespace <NAMESPACE>` flag.

## Rules

Reposaur will only query the rules that have the following prefixes (aka "kinds"):

### `violation_`, `fail_`, `error_`

Cause the CLI to exit with code `1`, the results in the SARIF report will have the `error` level.

### `warn_` 

Cause the CLI to exit with code `0`, the results in the SARIF report will have the `warning` level.

### `note_`, `info_`

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
* `statusCode` - The HTTP Response status code

Forbidden errors are treated in a special manner and will cause
policy execution to halt. Usually these errors happen when authentication is
required, a token is invalid or doesn't have sufficient permissions or rate limit
has been exceeded.

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
* `statusCode` - The HTTP Response status code

Forbidden errors are treated in a special manner and will cause
policy execution to halt. Usually these errors happen when authentication is
required, a token is invalid or doesn't have sufficient permissions or rate limit
has been exceeded.

# Use in GitHub Actions

```yaml
steps:
  - name: Setup Reposaur
    uses: reposaur/reposaur@main

  - run: reposaur --help
```

# Contributing

We appreciate every contribution, thanks for considering it!

- [Open an issue][issues] if you have a problem or found a bug
- [Open a Pull Request][pulls] if you have a suggestion, improvement or bug fix
- [Open a Discussion][discussions] if you have questions or want to discuss ideas

[issues]: https://github.com/reposaur/reposaur/issues
[pulls]: https://github.com/reposaur/reposaur/pulls
[discussions]: https://github.com/orgs/reposaur/discussions

# License

This project is released under the [MIT License](LICENSE).

[rego]: https://www.openpolicyagent.org/docs/latest/policy-language/
