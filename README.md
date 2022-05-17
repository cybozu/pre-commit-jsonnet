# pre-commit-jsonnet
[![lint and tests](https://github.com/cybozu-private/pre-commit-jsonnet/actions/workflows/lint_and_tests.yml/badge.svg)](https://github.com/cybozu-private/pre-commit-jsonnet/actions/workflows/lint_and_tests.yml)

`jsonnet-lint`, `jsonnetfmt` を実行する pre-commit hooks


## Usage

### Configurations
pre-commit hooks を適用したいリポジトリの `.pre-commit-config.yaml` に以下のような記載を追加する:

```yaml
repos:
    - repo: https://github.com/cybozu-private/pre-commit-jsonnet
      rev: HEAD
      hooks:
          - id: jsonnet-fmt
            args: ["--test"]  # commit 時に書き換えたい場合は --test を外し -i を追加する
          - id: jsonnet-lint
            args: ["--jpath", "lib/"]  # --jpath に指定するディレクトリは適宜書き換える。不要な場合は削除する。
```

`rev` は `pre-commit autoupdate` を実行することで最新の commit hash に更新出来る。

### Install hooks
`.pre-commit-config.yaml` を配置したリポジトリで以下を実行する:

```
pre-commit install
```
