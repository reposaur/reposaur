# Contributing

Beforehand, thank you for considering contributing to Reposaur! We welcome and
appreciate any help we can get.

Contributors are expected to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Feedback and Questions

We appreciate any feedback we can get and are happy to answer any questions you
might have!

Reach out to us in [Discussions][discussions] or in our [Slack Workspace][slack].

## Bug Reporting and Feature Requests

### Bug Reports

If you've found an issue in the CLI, SDK or Policy Engine, please
[open a bug report][bug-report]. We appreciate the feedback and will try to help
as soon as possible. The form will guide you on how to report the bug.

### Feature Requests

If you've an exiting idea, new use case or anything else, we'd love to hear it!
Please, [open a feature request][feature-request] and share it with the community!
The form will guide you on how to request a new feature.

## Development

To get your development environment ready you'll need to have the following
installed:

- [Go][go] >= 1.18
- [Task][task] >= 3.0

### Setting up

Fork the repository to your account and clone it to your machine. Build for the
first time and make sure everything is working as expected:

```console
$ task build
```

If you don't have Task installed you can issue the following commands:

```console
go mod tidy
go build -o ./.bin/rsr ./cmd/rsr/rsr.go \
    -ldflags "-s -w -X github.com/reposaur/reposaur/internal/build.Version=devel"
```

After the command above you should have `rsr` available in `./.bin`:

```console
$ ./.bin/rsr --version
rsr version devel
```

Happy coding!

### Creating a Pull Request

We don't have any rigid structure for Pull Requests right now, although that's
expected to change in the future.

For now, just explain briefly what your changes address and link any related
issues or other Pull Requests.

Creating an [issue][issues] prior to the Pull Request is highly recommended to
avoid multiple people working on the same issue, resulting in duplicated effort.

Visit the [Pull Requests][pulls] page.

## Releasing

To create new releases you'll need to have the following tools installed:

- [GoReleaser][goreleaser] >= 1.9
- [svu][svu] >= 1.9

[discussions]: https://github.com/orgs/reposaur/discussions
[issues]: https://github.com/reposaur/reposaur/issues
[pulls]: https://github.com/reposaur/reposaur/pulls
[bug-report]: https://github.com/reposaur/reposaur/issues/new?assignees=&labels=bug%2Ctriage&template=bug_report.yml&title=%5BBug%5D%3A+
[feature-request]: https://github.com/reposaur/reposaur/issues/new?assignees=&labels=enhancement%2Ctriage&template=feature_request.yml&title=%5BFeature%5D%3A+
[slack]: https://slack.reposaur.com
[go]: https://go.dev/
[task]: https://taskfile.dev/
[goreleaser]: https://goreleaser.com
[svu]: https://github.com/caarlos0/svu
