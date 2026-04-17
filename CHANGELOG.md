# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0](https://github.com/largeoliu/redmine-cli/compare/v0.1.2...v0.2.0) (2026-04-17)


### Features

* add release-please and commitlint for automated releases ([0806017](https://github.com/largeoliu/redmine-cli/commit/0806017ceed9521592a9da871c6898bdcf477d8c))
* **issues:** add --custom-field flag parsing ([3103fbc](https://github.com/largeoliu/redmine-cli/commit/3103fbc18a4dc04baf83fa7c3513847dc86cfc36))
* **issues:** add CustomFields to create/update requests ([3d7da08](https://github.com/largeoliu/redmine-cli/commit/3d7da082fe141fb0cc12ab9e1891fcc20b1a021b))
* **issues:** add interactive custom field prompting ([e87aa0b](https://github.com/largeoliu/redmine-cli/commit/e87aa0bbef1b6641ddeccaa532121e6816684097))
* **login:** mask API key input with '*' for security ([#31](https://github.com/largeoliu/redmine-cli/issues/31)) ([a633d96](https://github.com/largeoliu/redmine-cli/commit/a633d969de580cf0706aae253f013ee9512d79f4))
* **tracker:** add get command with custom fields support ([69930c2](https://github.com/largeoliu/redmine-cli/commit/69930c2048567cdff628fcf96f47776a89020750))


### Bug Fixes

* add govet to test exclusions, exclude G115 from gosec ([681953f](https://github.com/largeoliu/redmine-cli/commit/681953f4f7ecbaca6bfc43a33178a8e5b846d260))
* add misspell and staticcheck to test exclusions ([adca4bb](https://github.com/largeoliu/redmine-cli/commit/adca4bb0cc611b481058ed28c5d9b2532e9d9384))
* compat with gojq empty query and gosec Go toolchain ([a6c1c97](https://github.com/largeoliu/redmine-cli/commit/a6c1c97754dd9fb30567eaf7c27e810381422814))
* correct golangci.yml v2 format and restore test exclusions ([92ced21](https://github.com/largeoliu/redmine-cli/commit/92ced215ae233e5d9dee34c330779bda766fed26))
* improve test coverage and fix keyring test assertion ([#32](https://github.com/largeoliu/redmine-cli/issues/32)) ([1730d54](https://github.com/largeoliu/redmine-cli/commit/1730d543a2d68d11bf20dcf554eb8e13964a40e3))
* **issues:** handle missing value in id:X:value format and add fuzz test ([7924d34](https://github.com/largeoliu/redmine-cli/commit/7924d3495adb9fa8131c89ed81f04ed10be175a2))
* **issues:** handle parseCustomFieldFlags errors properly ([eec0e4f](https://github.com/largeoliu/redmine-cli/commit/eec0e4f7796db729901bbdd9e8001cededf3f832))
* make equalCustomFields order-independent ([e7b58e5](https://github.com/largeoliu/redmine-cli/commit/e7b58e5e7b0ba4b9e20636b5c44c1943cdc71113))
* pin release-please-action to v4.4.1 (v5 does not exist) ([#34](https://github.com/largeoliu/redmine-cli/issues/34)) ([8517153](https://github.com/largeoliu/redmine-cli/commit/8517153719573ebfcbc650af9e71ff785beb9d10))
* remove empty gosec config and pin trivy-action to valid version ([f070c29](https://github.com/largeoliu/redmine-cli/commit/f070c29381913fba0ee61308f720bf75c68b5203))
* resolve lint errcheck failures and gosec G115 false positive ([e024d13](https://github.com/largeoliu/redmine-cli/commit/e024d139552e6c335ec4bde0cb93b3113a91fa45))
* resolve merge conflicts with master ([512e519](https://github.com/largeoliu/redmine-cli/commit/512e519ae278888a95b1712ef2133d179b4f8c1d))
* update nolint directive and add TestGetCommand tests ([9238925](https://github.com/largeoliu/redmine-cli/commit/923892557ee88a56102a287949207ee035ed1e68))
* update package-lock.json to resolve node-tar CVE vulnerabilities ([ed586ed](https://github.com/largeoliu/redmine-cli/commit/ed586edd21d26a8c9aab7de73f6277cf98f09a67))
* upgrade release-please-action to v5 and fix default branch to master ([#33](https://github.com/largeoliu/redmine-cli/issues/33)) ([037b7aa](https://github.com/largeoliu/redmine-cli/commit/037b7aa245dde3cbf3739b4e475b9f0d893002e4))
* use //nolint:errcheck,gosec to suppress G104 in golangci-lint ([1dc6aab](https://github.com/largeoliu/redmine-cli/commit/1dc6aab252fe45e8db1dea6ea20149f9772551dc))
* use release token for release-please ([#39](https://github.com/largeoliu/redmine-cli/issues/39)) ([c79873c](https://github.com/largeoliu/redmine-cli/commit/c79873cc5a09a5b75a219337d6227fb30a87af45))

## [Unreleased]

### Added
- Initial release
