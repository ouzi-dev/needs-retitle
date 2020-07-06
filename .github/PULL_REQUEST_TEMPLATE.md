---
name: Pull Request
about: Suggest a new feature
title: ''
labels: blocked-needs-validation, enhancement
assignees: ''

---

# Pull Request Template

Ensure that the Pull Request title starts with:
  - `major: ` if the pull request has breaking changes with previous versions,
  - `feat: ` if it's a new feature with no breaking changes,
  - `fix: ` if it's a bug fix,
  - `refactor: ` if it's just a refactor,
  - `doc: ` if it's documentation related,
  - `build: ` if it's build related.
Note: if you are still working on the changeset but wish to share with us your progress so far, please prepend `WIP` on your title.

## Description

Include a summary of the change and which issue is fixed. Please also include relevant context to help us understand the motivation behind the change. 

Fixes # (issue if applicable)

## Type of change

Please delete options that are not relevant.

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] This change requires a documentation update

## Checklist:

- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] I have added tests that show my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] I have checked my code and corrected any misspellings