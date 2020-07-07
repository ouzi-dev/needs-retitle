# needs-retitle <!-- omit in toc -->

- [Overview](#overview)
- [Configuration](#configuration)
  - [Enable only in some repositories of an organization](#enable-only-in-some-repositories-of-an-organization)
- [Build](#build)
- [Deploy](#deploy)

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

* Add the new label to `missingLabels` in the tide settings (in prow usually in `config.yaml`), that way the label `needs-retitle` will stop tide from merging the pull requests, example:

  ```
  tide:
    sync_period: 1m
    merge_method:
      my-org: squash
    pr_status_base_urls:
      "*": https://prow.my-host.com/pr
    blocker_label: tide/merge-blocker
    squash_label: tide/merge-method-squash
    rebase_label: tide/merge-method-rebase
    merge_label: tide/merge-method-merge
    context_options:
      from-branch-protection: true
      skip-unknown-contexts: false
    queries:
      - orgs:
          - my-org
        labels:
          - lgtm
          - approved
        missingLabels:
          - do-not-merge
          - do-not-merge/hold
          - do-not-merge/work-in-progress
          - needs-rebase
          - do-not-merge/invalid-owners-file
          - needs-retitle
  ```

### Enable only in some repositories of an organization

To enable the plugin only in some repositories of the same organization you have two options:

* You can add to the `external_plugins` options all the repositories you want like:

  ```
  external_plugins:
    org-foo/repo-bar:
    - name: needs-retitle
      events:
      - pull_request
    org-foo/repo-bar-2:
    - name: needs-retitle
      events:
      - pull_request
    org-foo/repo-bar-3:
    - name: needs-retitle
      events:
      - pull_request
    ...
  ```

* You can use `require_enable_as_not_external`, this way you can enable the plugin for the whole organization at the `external_plugins` config level, and then just add the plugin to each repo like any other plugin, example:
  
  * First we enable the `require_enable_as_not_external` setting for the plugin:
  
    ```
    needs_retitle:
      regexp: "^(fix:|feat:|major:).*$"
      require_enable_as_not_external: true
      error_message: |
        Invalid title for the PR, the title needs to be like:

        fix: this is a fix commit
        feat: this is a feature commit
        major: this is a major commit    
    ```

  * Then we enable the plugin for the whole organization in `external_plugins`:
  
  ```
  external_plugins:
    org-foo:
    - name: needs-retitle
      events:
      - pull_request
  ```

  * And now we enable it like any other plugin for the repositories:

  ```
  plugins:
    org-foo/test-infra:
      - approve 
      - assign 
      - config-updater
      - needs-retitle
    org-foo/my-repo:
      - approve 
      - assign 
      - lgtm
      - needs-retitle
    org-foo/another-repo:
      - trigger
      - needs-retitle
  ```

## Build 

* To make just the binary run: `make build`
* To build an specific version run: `VERSION=v0.1.0 make build`
* To run the tests run: `make test`
* To build the docker image run: `make docker-build`

## Deploy

You can find manifest to use as example to deploy the plugin in [deploy](./deploy)

__Remember to change the version, the manifests are using as version `canary`__