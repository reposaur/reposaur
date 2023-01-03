#guide 

```rego
# METADATA
# title: Default branch is not protected
# description: >
#   The repository default branch is not protected, allowing anyone with write
#   access to push to it directly.
# schemas:
#   - input: github.repository
# custom:
#   tags: [security, compliance, best-practices]
warn_default_branch_not_protected {
    resp := github.request("GET /repos/{owner}/{repo}/branches/{branch}/protection", {
        "owner": input.owner.login,
        "repo": input.name,
        "branch": input.default_branch,
    })
    resp.status == 404
} 
```
