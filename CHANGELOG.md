# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.4](https://github.com/largeoliu/redmine-cli/compare/v0.4.3...v0.4.4) (2026-04-25)


### Bug Fixes

* remove npm distribution pipeline ([#78](https://github.com/largeoliu/redmine-cli/issues/78)) ([2c4924f](https://github.com/largeoliu/redmine-cli/commit/2c4924f6f642a8c798f9e0bb7f9c501e38c4240d))

## [0.4.3](https://github.com/largeoliu/redmine-cli/compare/v0.4.2...v0.4.3) (2026-04-25)


### Bug Fixes

* **release:** use npm trusted publishing ([#76](https://github.com/largeoliu/redmine-cli/issues/76)) ([032ad3d](https://github.com/largeoliu/redmine-cli/commit/032ad3d76cf682ac981a38b646b9acc300a098c3))

## [0.4.2](https://github.com/largeoliu/redmine-cli/compare/v0.4.1...v0.4.2) (2026-04-25)


### Bug Fixes

* **release:** remove conflicting npm override ([#74](https://github.com/largeoliu/redmine-cli/issues/74)) ([4e0fa7a](https://github.com/largeoliu/redmine-cli/commit/4e0fa7a6f439f2a07e8183484099c713f3ec17a6))

## [0.4.1](https://github.com/largeoliu/redmine-cli/compare/v0.4.0...v0.4.1) (2026-04-25)


### Bug Fixes

* **release:** use correct info command in smoke test ([#72](https://github.com/largeoliu/redmine-cli/issues/72)) ([7c408d6](https://github.com/largeoliu/redmine-cli/commit/7c408d64c30a34bb46e89b5122ab0e8000b0c460))

## [0.4.0](https://github.com/largeoliu/redmine-cli/compare/v0.3.0...v0.4.0) (2026-04-25)


### Features

* add sprint get command to show sprint details via agile_sprints API ([#66](https://github.com/largeoliu/redmine-cli/issues/66)) ([c9f0223](https://github.com/largeoliu/redmine-cli/commit/c9f022307b135ebe806a9e524f7c3d05589e49b6))
* add version-id filtering for issues ([#65](https://github.com/largeoliu/redmine-cli/issues/65)) ([486a671](https://github.com/largeoliu/redmine-cli/commit/486a671350f3b96a47fd41387fc50908d41305f9))
* **agile:** add agile content report ([c7c1265](https://github.com/largeoliu/redmine-cli/commit/c7c1265ad0070ab135246ace26ee0e93980ce811))
* filter issues by sprint ([#68](https://github.com/largeoliu/redmine-cli/issues/68)) ([e266b3d](https://github.com/largeoliu/redmine-cli/commit/e266b3db1fbd7fe5ef53fad6af3e06ee9c940032))
* require --project-id for issue list ([#67](https://github.com/largeoliu/redmine-cli/issues/67)) ([997e76a](https://github.com/largeoliu/redmine-cli/commit/997e76adb46cc3f316a726e520d293bbdc26d8c4))
* **sprint:** 添加 --details 标志展示完整 sprint 详情 ([b423b90](https://github.com/largeoliu/redmine-cli/commit/b423b908df28834437d87cda9e2a73f91625f7cf))
* **sprint:** 添加 sprint 子命令支持 ([9f8c5cc](https://github.com/largeoliu/redmine-cli/commit/9f8c5ccede67f7aa564aaa3903101cfb2b3d7a99))
* support multiple values for issue list filters ([#70](https://github.com/largeoliu/redmine-cli/issues/70)) ([5d81bc9](https://github.com/largeoliu/redmine-cli/commit/5d81bc9fa9a24c2e8e4213feb7a94b262ca38e17))
* 提升测试覆盖率并添加 Sprint 和 Agile 功能 ([223f630](https://github.com/largeoliu/redmine-cli/commit/223f6303ce00d8381111059c831b1efaa5d5a00d))


### Bug Fixes

* override tar dependency to resolve CVE-2026-29786 and CVE-2026-31802 ([#71](https://github.com/largeoliu/redmine-cli/issues/71)) ([cb6d60b](https://github.com/largeoliu/redmine-cli/commit/cb6d60b6a1c2416eac8b2f762914907a1c779208))
* resolve master merge conflict for PR 53 ([ad66bc6](https://github.com/largeoliu/redmine-cli/commit/ad66bc661c033e2b44902952ad0f718a872e858b))

## [0.3.0](https://github.com/largeoliu/redmine-cli/compare/v0.2.5...v0.3.0) (2026-04-18)


### Features

* 添加 upgrade 命令 ([#51](https://github.com/largeoliu/redmine-cli/issues/51)) ([bdc99ce](https://github.com/largeoliu/redmine-cli/commit/bdc99ceebe289ad3014298559076bd045a97f56c))

## [0.2.5](https://github.com/largeoliu/redmine-cli/compare/v0.2.4...v0.2.5) (2026-04-18)


### Bug Fixes

* skip husky when publishing ([#49](https://github.com/largeoliu/redmine-cli/issues/49)) ([3bb7c90](https://github.com/largeoliu/redmine-cli/commit/3bb7c90685c4db15d08c32e554f95e2d014b7a75))

## [0.2.4](https://github.com/largeoliu/redmine-cli/compare/v0.2.3...v0.2.4) (2026-04-18)


### Bug Fixes

* ignore npm lifecycle scripts during release publish ([#47](https://github.com/largeoliu/redmine-cli/issues/47)) ([1458fb5](https://github.com/largeoliu/redmine-cli/commit/1458fb5de1e3bfd1780b5da481a1877fddd4e111))

## [0.2.3](https://github.com/largeoliu/redmine-cli/compare/v0.2.2...v0.2.3) (2026-04-18)


### Bug Fixes

* align release pipeline with supported targets ([#45](https://github.com/largeoliu/redmine-cli/issues/45)) ([314a0e1](https://github.com/largeoliu/redmine-cli/commit/314a0e163829498fcd2df0341f0eaccf1865cb82))

## [0.2.2](https://github.com/largeoliu/redmine-cli/compare/v0.2.1...v0.2.2) (2026-04-18)


### Bug Fixes

* switch release signing to cosign bundles ([#43](https://github.com/largeoliu/redmine-cli/issues/43)) ([def1aca](https://github.com/largeoliu/redmine-cli/commit/def1acaddea771c533d33eb69083a9d205162b28))

## [0.2.1](https://github.com/largeoliu/redmine-cli/compare/v0.2.0...v0.2.1) (2026-04-17)


### Bug Fixes

* install cosign before goreleaser ([#40](https://github.com/largeoliu/redmine-cli/issues/40)) ([3f36172](https://github.com/largeoliu/redmine-cli/commit/3f361721092e755364c677b892a4d23b8f90a195))

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
