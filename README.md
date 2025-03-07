# GitHub Notifications Cleaner

> A CLI tool to automatically clean up GitHub notifications based on configurable rules.

[![CI](https://img.shields.io/github/actions/workflow/status/brpaz/github-notifications-cleaner/ci.yml?style=for-the-badge)](https://github.com/brpaz/github-notifications-cleaner/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/brpaz/github-notifications-cleaner?style=for-the-badge)](https://goreportcard.com/report/github.com/brpaz/github-notifications-cleaner)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

## 📋 Description

GitHub Notifications Cleaner is a command-line tool designed to help developers manage their GitHub notifications more efficiently. The tool automatically marks notifications as "done" based on some rules:

- Mark notifications older than X days as done
- Mark notifications from closed pull requests as done
- Mark notifications from closed issues as done

## 🎯 Motivation

GitHub notifications are a great way to keep track of the tasks that require your attention, like opened issues in your repos or Pull requests that are waiting for your review.

But if you have many active repos, the amount of notifications can grow pretty quickly. I often found myself overwhelmed by the constant stream of GitHub notifications. Most of the times they are already outdated and/or not actionable (Ex: merged PRs). This makes it harder to find the notifications that really matter.

This tool was created specifically to address these pain points by automatically pruning notifications that are no longer relevant or actionable, so that I can focus what the ones that really need my attention.


## 🚀 Getting Started

### Pre-requisites

You must have a GitHub token, with `notifications` and `repo` permissions. You can generate a new one [here](https://github.com/settings/tokens/new).

> [!IMPORTANT] Use classic token instead of fine grained token
> The Github token will need to be able to access all your repos, in order to retrieve any Issue or Pull requests asscociated with a Notification. It´s best to use a classic token instead of a fine grained token for this scenario.

### Installation

#### Using Go

```bash
go install github.com/brpaz/github-notifications-cleaner@latest
```

#### Using docker

```bash
docker pull ghcr.io/brpaz/github-notifications-cleaner:latest
```

#### Binary downloads

You can download pre-built binaries from the [releases page](https://github.com/brpaz/github-notifications-cleaner/releases)

### Usage

```shell
github-notifications-cleaner clean --token YOUR_GITHUB_TOKEN
```

or with docker:

```shell
docker run --rm -e GITHUB_TOKEN=<token> ghcr.io/brpaz/github-notifications-cleaner:latest clean
```

#### Command Arguments

The `github-notifications-cleaner clean` command accepts the following arguments:

| Argument           | Short | Required | Default | Description                                                                                                      |
| ------------------ | ----- | -------- | ------- | ---------------------------------------------------------------------------------------------------------------- |
| `--token`          | `-t`  | Yes      | -       | GitHub Personal Access Token with notifications access. Can also be set via `GITHUB_TOKEN` environment variable. |
| `--days-threshold` | `-d`  | No       | 30      | Mark notifications older than this number of days as done.                                                       |
| `--dry-run`        | `-n`  | No       | `false` | Run in dry-run mode, which shows what would be cleaned without actually marking notifications as done.           |

> [!TIP]
> The GitHub token should have `notifictation` and `repo` permissions.

#### Examples

```bash
# Basic usage
github-notifications-cleaner clean --token YOUR_GITHUB_TOKEN

# Mark notifications older than 30 days as done
github-notifications-cleaner clean --token YOUR_GITHUB_TOKEN --days-threshold 30

# Run in dry-run mode to preview what would be cleaned
github-notifications-cleaner clean --token YOUR_GITHUB_TOKEN --dry-run
```

## 🤝 Contributing

Check [CONTRIBUTING.md](CONTRIBUTING.md) files for details.

## 🫶 Support

If you find this project helpful and would like to support its development, there are a few ways you can contribute:

[![Sponsor me on GitHub](https://img.shields.io/badge/Sponsor-%E2%9D%A4-%23db61a2.svg?&logo=github&logoColor=red&&style=for-the-badge&labelColor=white)](https://github.com/sponsors/brpaz)

<a href="https://www.buymeacoffee.com/Z1Bu6asGV" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: auto !important;width: auto !important;" ></a>


## 📃 License

Distributed under the MIT License. See [LICENSE](LICENSE) file for details.

## 📩 Contact

✉️ **Email** - [oss@brunopaz.dev](oss@brunopaz.dev)

🖇️ **Source code**: [https://github.com/brpaz/github-notifications-cleaner](https://github.com/brpaz/github-notifications-cleaner)




