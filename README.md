# needs-retitle <!-- omit in toc -->

- [Overview](#overview)
- [Configuration](#configuration)
- [Build and deploy](#build-and-deploy)

## Overview

`needs-retitle` is an external plugin for [prow](https://github.com/kubernetes/test-infra/tree/master/prow) to avoid merging PRs when the title doesn't match a provided regular expression.

It is based on the [needs-rebase](https://github.com/kubernetes/test-infra/tree/master/prow/external-plugins/needs-rebase) plugin, so the code is more or less the same.

The plugin will check pull requests in the enabled repos and will add a tag `needs-retitle` to the pull requests whose titles don't match the provided regular expression.

The plugin will run every time a pull request is created, edited or new commits are added. It will also run periodically checking open pull requests.

## Configuration

You'll need to add new things to your prow `plugins.yaml` file:

* The plugin configuration: you need to provide a regular expression (**required**), and an optional error message. The message will be added as a comment when the plugin detects a pull request with a title that doesn't match the regular expression. If no error message is provided the plugin will add a default message. Example:

    ```
    needs_retitle:
      regexp: "^(fix:|feat:|major:).*$"
      error_message: |
        Invalid title for the PR, the title needs to be like:

        fix: this is a fix commit
        feat: this is a feature commit
        major: this is a major commit
    ```

* The settings to enable it as external plugin for prow, for example:

  ```
  external_plugins:
    org-foo/repo-bar:
    - name: needs-retitle
      # No endpoint specified implies "http://{{name}}".
      events:
      - pull_request
      # Dispatching issue_comment events to the needs-retitle plugin is optional. If enabled, this may cost up to two token per comment on a PR. If `ghproxy`
      # is in use, these two tokens are only needed if the PR or its mergeability changed.
      - issue_comment
  ```

## Build and deploy

WIP