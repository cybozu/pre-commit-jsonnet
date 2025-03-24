# pre-commit-jsonnet
[![lint and tests](https://github.com/cybozu/pre-commit-jsonnet/actions/workflows/lint_and_tests.yml/badge.svg)](https://github.com/cybozu/pre-commit-jsonnet/actions/workflows/lint_and_tests.yml)

[pre-commit](https://pre-commit.com/) hooks for [jsonnet](https://jsonnet.org/).


## Usage

Add the following to your `.pre-commit-config.yaml`:

```yaml
repos:
    - repo: https://github.com/cybozu/pre-commit-jsonnet
      rev: v0.3.1
      hooks:
          - id: jsonnet-fmt
            args: ["--test"]  # you can specify any options of jsonnetfmt command
          - id: jsonnet-lint
            args: []  # you can specify any options of jsonnet-lint command
```
