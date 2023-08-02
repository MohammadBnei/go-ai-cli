

## [0.12.1](https://github.com/MohammadBnei/go-openai-cli/compare/0.12.0...0.12.1) (2023-08-02)


### Bug Fixes

* **terminal:** removed clear screen command ([11446f3](https://github.com/MohammadBnei/go-openai-cli/commit/11446f3ddb728f59961e2765a5296ffc1c226763))

# [0.12.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.11.0...0.12.0) (2023-08-02)


### Bug Fixes

* **audio:** adding en by default for audio record ([87470e3](https://github.com/MohammadBnei/go-openai-cli/commit/87470e3a8148aa708dabf5268370f49245544681))
* **prompt:** fixed forgotten update of the promptConfig with user prompt ([49b57f1](https://github.com/MohammadBnei/go-openai-cli/commit/49b57f104090e28735dcdbbd7d7d57ddca44addc))
* **speech.go:** add missing import for io package ([a128df0](https://github.com/MohammadBnei/go-openai-cli/commit/a128df061a8a82d4b63634d3844b035611ed1036))
* **speech.go:** change maxMinutes value from 5 to 4 ([82e577d](https://github.com/MohammadBnei/go-openai-cli/commit/82e577dd09ae31400e96bf790ff4776d525b6209))


### Features

* **audio.go:** add language parameter to SpeechToText function ([dba93d0](https://github.com/MohammadBnei/go-openai-cli/commit/dba93d05567d0a48aa88ab0a05fa20b4352fc73a))
* **audio:** audio and normal usage now joint by build args ([8774be0](https://github.com/MohammadBnei/go-openai-cli/commit/8774be0b9d0974f04fe0b875fc6a404bb34652b2))
* **cmd/speech:** add speech command to convert speech to text ([7b92b05](https://github.com/MohammadBnei/go-openai-cli/commit/7b92b051b279247f18d3a042590a39ba7f21a372))
* **mask:** adding hugging face mask abilities ([4f808f4](https://github.com/MohammadBnei/go-openai-cli/commit/4f808f4ef6531c481569718ed51bfe86fef3c648))
* **prompt:** changed way to run cmd ([f40c1dc](https://github.com/MohammadBnei/go-openai-cli/commit/f40c1dc20fd15e97d39f8c9aa15984ad9f9df865))
* **speech:** implementig speech to text ([a4a75bc](https://github.com/MohammadBnei/go-openai-cli/commit/a4a75bc4f4fbcbbef0eeeb9b651812ed16250b87))

# [0.11.0](https://github.com/MohammadBnei/go-openai-cli/compare/0.10.2...0.11.0) (2023-07-24)


### Bug Fixes

* **cmd/config.go:** change variable name 'path' to 'filePath' for clarity ([907ee8b](https://github.com/MohammadBnei/go-openai-cli/commit/907ee8ba7cf35b5dc5e7354b6770adfdf1a70407))
* **markdown:** Fixed backtick error ([4dc3a83](https://github.com/MohammadBnei/go-openai-cli/commit/4dc3a839355e6d6a3e099561728326ba28004054))
* **md-format:** utilizing "md" instead of !md ([0ed38c3](https://github.com/MohammadBnei/go-openai-cli/commit/0ed38c346e70715bf4c2ffc16a1cbc064ad87e02))
* **writer.go:** add support for single backticks in Write method ([71d4caf](https://github.com/MohammadBnei/go-openai-cli/commit/71d4cafa4fd4304ff5befb8eb28e4dab72108f2a))


### Features

* **format:** adding support for markdown format in terminal ([6b4967a](https://github.com/MohammadBnei/go-openai-cli/commit/6b4967af3b01943e2228a9e6f5d867d63584e317))

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