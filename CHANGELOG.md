

## [0.10.2](https://github.com/MohammadBnei/go-openai-cli/compare/0.10.1...0.10.2) (2023-06-28)


### Bug Fixes

* **config.go:** change default value of "messages-length" flag from 10 to 20 ([42e166c](https://github.com/MohammadBnei/go-openai-cli/commit/42e166c76ebef285419fbc0b3273470caed00c0d))

## [0.10.1](https://github.com/MohammadBnei/go-openai-cli/compare/0.10.0...0.10.1) (2023-05-26)

# [0.10.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.9.0...0.10.0) (2023-05-26)


### Features

* **main.go:** add support for debugging with pprof ([f29a2e1](https://github.com/MohammadBnei/go-openai-cli/commit/f29a2e1b2bf90f17710baa966a05541d9ad42a5a))

# [0.9.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.8.0...0.9.0) (2023-05-25)


### Bug Fixes

* **openai.go, prompt.go:** add context cancellation to SendPrompt function ([09bc6e6](https://github.com/MohammadBnei/go-openai-cli/commit/09bc6e686e95e22b7589eea4b1955f321ac0a4d3))


### Features

* **config:** adding messages retention numberto config ([7359ac9](https://github.com/MohammadBnei/go-openai-cli/commit/7359ac958b999c30990fb8a79281a2b995859983))
* **config:** adding messages retention numberto config ([5a9af62](https://github.com/MohammadBnei/go-openai-cli/commit/5a9af6231672d9829213c2701e74c5a97d55c198))
* **copy:** added clipboard copy capacities ([35b07c5](https://github.com/MohammadBnei/go-openai-cli/commit/35b07c5eb000b3d43f0c8e86b19fa674151d02bc))
* **image:** adding image generation capabilities ([c71aa6d](https://github.com/MohammadBnei/go-openai-cli/commit/c71aa6d2da60621fdfd1738a27fa8d06fb539b21))

# [0.8.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.7.0...0.8.0) (2023-05-15)


### Features

* **file:** tree view when hovering folder ([97f122c](https://github.com/MohammadBnei/go-openai-cli/commit/97f122cbd99fab8bba85914ddd8a072e638c8872))
* **test.yml:** add paths filter for Go files ([3336ee9](https://github.com/MohammadBnei/go-openai-cli/commit/3336ee95418b045e23481c93c84a03595236fe45))

# [0.7.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.6.0...0.7.0) (2023-05-15)


### Bug Fixes

* **prompt-help:** updated the help section ([55a2cf9](https://github.com/MohammadBnei/go-openai-cli/commit/55a2cf954ef86eb9d5ff93d7066a6505f04ddfc8))


### Features

* **config:** add support for environment variable CONFIG ([48dedf1](https://github.com/MohammadBnei/go-openai-cli/commit/48dedf113387436e3ec6d1f9a1a99eef6106c183))

# [0.6.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.5.0...0.6.0) (2023-05-15)


### Features

* **client-release.yml:** add fail-fast option to matrix strategy ([b53e021](https://github.com/MohammadBnei/go-openai-cli/commit/b53e02147fd201e5068ccca55f1fa936625e782e))
* **prompt.go:** add support for fuzzy file search and multi-file selection ([de9e24a](https://github.com/MohammadBnei/go-openai-cli/commit/de9e24a260c896d5403887a42a8edb22d2b357ce))

# [0.5.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.4.0...0.5.0) (2023-05-15)


### Features

* add Dockerfile and GitHub Actions workflow for building and pushing Docker image ([d58cebc](https://github.com/MohammadBnei/go-openai-cli/commit/d58cebc12c5dddefffd763dfce55ef9a358771fb))

# [0.4.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.3.2...0.4.0) (2023-05-15)


### Bug Fixes

* **config:** create config directory if it does not exist ([61f20b8](https://github.com/MohammadBnei/go-openai-cli/commit/61f20b84013929d01c940a738dcdc48b41c035e5))


### Features

* add zsh completion script ([5895130](https://github.com/MohammadBnei/go-openai-cli/commit/5895130b23d1838f68c89197237c17dfd036f7dd))
* add zsh completion script ([8b2708b](https://github.com/MohammadBnei/go-openai-cli/commit/8b2708ba048c7f910743a6ddfe63c54194fa4efe))

## [0.3.2](https://github.com/MohammadBnei/go-openai-cli/compare/0.3.1...0.3.2) (2023-05-15)

## [0.3.1](https://github.com/MohammadBnei/go-openai-cli/compare/0.3.0...0.3.1) (2023-05-15)


### Bug Fixes

* **node_modules:** removed node modules and pnpm.lock ([c487bd5](https://github.com/MohammadBnei/go-openai-cli/commit/c487bd5b037148572e609e8f2e6ae3a968af70cf))

# [0.3.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.2.0...0.3.0) (2023-05-14)


### Features

* **openai.go:** add ClearMessages function ([3c84b06](https://github.com/MohammadBnei/go-openai-cli/commit/3c84b06cc434d12d51c87bf7df322849bab5f17d))

# [0.2.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.1.1...0.2.0) (2023-05-14)


### Features

* **release:** testing automated release ([a0b300a](https://github.com/MohammadBnei/go-openai-cli/commit/a0b300ad54d93eeb6fbf3b2c2e13eca1b9a17418))

## [0.1.1](https://github.com/MohammadBnei/go-openai-cli/compare/0.1.0...0.1.1) (2023-05-14)

# 0.1.0 (2023-05-14)


### Features

* add license, cmd/config.go, cmd/prompt.go, cmd/root.go, gowatch.yml, main.go, service/openai.go ([efd62fc](https://github.com/MohammadBnei/go-openai-cli/commit/efd62fcf7cb62998e689f8b19f561fddca18fb47))
* **client-release.yml:** add GitHub workflow for releasing Go binary ([2dc52bc](https://github.com/MohammadBnei/go-openai-cli/commit/2dc52bca56203f7e5324cc46243f7088d44f4675))
* **openai.go, prompt.go:** add support for adding file contents to prompt ([9b24376](https://github.com/MohammadBnei/go-openai-cli/commit/9b2437606686bb7a1a35418d5fea2d830fc1e522))
* **project:** adding release it and updating go module ([a031471](https://github.com/MohammadBnei/go-openai-cli/commit/a03147195008f7335daba0415d8a1f37f3e2306a))