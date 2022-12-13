<div align="center">

[![logo][logo]][website]

# Reposaur

[![go-report][go-report-badge]][go-report]
[![tests-workflow][tests-workflow-badge]][tests-workflow]
[![license][license-badge]]()
[![discussions][discussions-badge]][discussions]
[![discord][discord-badge]][discord-invite]
[![twitter][twitter-badge]][twitter]

**Reposaur is the open source compliance tool for development platforms.**

Audit, verify and report on your data and configurations easily with pre-defined and/or custom policies. <br />
Supports GitHub. GitLab, BitBucket and Gitea support soon.
⚠️ before 1.0.0 expect some bugs and API changes ⚠️
[Getting started](#getting-started) •
[Installation](#installation) •
[Documentation][docs] •
[Guides](#guides) •
[Integrations](#integrations)

</div>

# Getting Started

Have you ever felt like you don't know what's happening in your GitHub/GitLab/BitBucket repositories? Between 100s or 1000s of them it's hard to make sure every single one is compliant to certain security and best practices guidelines.

Reposaur is here to fix that, empowering you to focus on your work instead of hunting for issues and misconfigurations.

## Features

- Custom policies using the [Rego][rego] policy language ([learn more][docs-policy])
- A simple, composable and easy-to-use CLI ([learn more][docs-cli])
- Extendable using a straightforward SDK (written in Go)
- Reports follow the standard SARIF format, enabling easy integrations with different systems
- Policies can be unit tested, guaranteeing they work as expected
- Integration with the major development platforms (see [Integrations](#integrations))
- Easily integrate new platforms using the SDK

## Guides

- [Writing your first policy](https://docs.reposaur.com/guides/writing-your-first-policy)

# Installation

#### Homebrew Tap

```shell
$ brew install reposaur/tap/reposaur
```

#### DEB, RPM and APK Packages

Download the `.deb`, `.rpm` or `.apk` packages from the [releases page][releases]
and install them with the appropriate tools.

#### Go

```shell
$ go install github.com/reposaur/reposaur/cmd/rsr@latest
```

#### Script

The script will download the latest release to a temporary directory and decompress
it to `$HOME/.reposaur`.

```shell
$ curl -sfL https://get.reposaur.com | bash
```

# Integrations

| Platform               | Status      | Details                                                                                   |
|------------------------|-------------|-------------------------------------------------------------------------------------------|
| [GitHub][github]       | In progress | [Provider][github-provider] • [GitHub App][github-app] • [GitHub Actions][github-actions] |
| [GitLab][gitlab]       | Planned     | N/A                                                                                       |
| [Gitea][gitea]         | Planned     | N/A                                                                                       |
| [BitBucket][bitbucket] | Not planned | N/A                                                                                       |

# Contributing

We appreciate every contribution, thanks for considering it!

**TLDR;**

- [Open an issue][issues] if you have a problem or found a bug
- [Open a Pull Request][pulls] if you have a suggestion, improvement or bug fix
- [Open a Discussion][discussions] if you have questions or want to discuss ideas

Check our [Contributing Guide](CONTRIBUTING.md) for more detailed information.

# License

This project is released under the [MIT License](LICENSE).

[website]: https://reposaur.com
[docs]: https://docs.reposaur.com
[docs-policy]: https://docs.reposaur.com/policy
[docs-cli]: https://docs.reposaur.com/cli/exec
[issues]: https://github.com/reposaur/reposaur/issues
[pulls]: https://github.com/reposaur/reposaur/pulls
[logo]: https://user-images.githubusercontent.com/8532541/169531963-bafd3cbf-dadd-486d-83cc-10a4d39c1dbc.png
[rego]: https://www.openpolicyagent.org/docs/latest/policy-language/
[license]: https://github.com/reposaur/reposaur/blob/main/LICENSE
[license-badge]: https://img.shields.io/github/license/reposaur/reposaur?style=flat-square&color=blueviolet
[go-report]: https://goreportcard.com/report/github.com/reposaur/reposaur
[go-report-badge]: https://goreportcard.com/badge/github.com/reposaur/reposaur?style=flat-square&color=blueviolet
[tests-workflow]: https://github.com/reposaur/reposaur/actions/workflows/test.yml
[tests-workflow-badge]: https://img.shields.io/github/workflow/status/reposaur/reposaur/Test?label=tests&style=flat-square
[discussions]: https://github.com/orgs/reposaur/discussions
[discussions-badge]: https://img.shields.io/github/discussions/reposaur/reposaur?style=flat-square&color=blueviolet
[discord-invite]: https://discord.gg/jpx4sqkQYY
[discord-badge]: https://img.shields.io/discord/1021712577132240898?label=discord&style=flat-square&color=blueviolet
[twitter]: https://twitter.com/reposaurhq
[twitter-badge]: https://img.shields.io/badge/twitter-%40reposaurhq-blueviolet?style=flat-square
[github]: https://github.com
[github-app]: https://docs.reposaur.com/integrations/github-app
[github-actions]: https://docs.reposaur.com/integrations/github-actions/setup-reposaur
[github-provider]: https://docs.reposaur.com/
[gitlab]: https://gitlab.com
[gitea]: https://gitea.io
[bitbucket]: https://bitbucket.org
[releases]: https://github.com/reposaur/reposaur/releases/latest
