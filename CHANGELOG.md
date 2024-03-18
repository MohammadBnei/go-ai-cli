

# [0.18.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.17.3...0.18.0) (2024-03-18)


### Bug Fixes

* **chat.go:** using a pointer to the updated chat message instead of the value. Updating the whole currentChatMessages (user & assistant) instead of just the content. Fixed a bug on order where I frogot to add +1 to the length of the array ([b8731b2](https://github.com/MohammadBnei/go-ai-cli/commit/b8731b2fab79a258a640053503ccabf727bf7a59))


### Features

* **file:** added an option to add all text files from the current directory. Added a proper list of selected files ([43e9379](https://github.com/MohammadBnei/go-ai-cli/commit/43e9379f23e88ab0afc4c3b2133611b0fa8cb3fc))
* **filepicker:** added an option to disable plain text filtering ([bc8c895](https://github.com/MohammadBnei/go-ai-cli/commit/bc8c895d85941e1afd0544499deb4048d1f61b53))
* **godcontext:** added godcontext to gracefully end all contexts before quitting ([e2903a1](https://github.com/MohammadBnei/go-ai-cli/commit/e2903a1a903820eb104f2f34ed0514e9f8e1acbd))

## [0.17.3](https://github.com/MohammadBnei/go-ai-cli/compare/0.17.2...0.17.3) (2024-03-01)


### Bug Fixes

* **helper.go:** handle case where message is not found for a file to prevent nil pointer error ([7efbcfe](https://github.com/MohammadBnei/go-ai-cli/commit/7efbcfe2658b855c5441d2f16fdcef9cd47bd468))

## [0.17.2](https://github.com/MohammadBnei/go-ai-cli/compare/0.17.1...0.17.2) (2024-03-01)

## [0.17.1](https://github.com/MohammadBnei/go-ai-cli/compare/0.17.0...0.17.1) (2024-03-01)


### Bug Fixes

* **speech_np.go:** change return value of SpeechToText function from nil to empty string to match the error type ([27985fc](https://github.com/MohammadBnei/go-ai-cli/commit/27985fc5c4c25a50b62e2bbd4044bb70750d9289))

# [0.17.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.16.1...0.17.0) (2024-03-01)


### Bug Fixes

* **all:** somewhat stable ([f7bfcef](https://github.com/MohammadBnei/go-ai-cli/commit/f7bfcefde00082a8c0448cd0c31fbcd473f6607f))
* **change-response:** fixed a panic error when changing displayed exchange ([65845ca](https://github.com/MohammadBnei/go-ai-cli/commit/65845ca6bffae655d0564597993083e51cd266db))
* **chat.go:** fix viewport width calculation to correctly wrap markdown renderer ([c622954](https://github.com/MohammadBnei/go-ai-cli/commit/c6229544ac45dba62d3dc02213a5f52415c8ed7d))
* **chat.go:** fix width calculation for title style to match the updated frame size ([e9297fd](https://github.com/MohammadBnei/go-ai-cli/commit/e9297fd694ccdbb6840073117655100e0f6944a0))
* **chat:** remove unused reset function and its dependencies ([5db75bc](https://github.com/MohammadBnei/go-ai-cli/commit/5db75bcf4dd01f83485cac15eff1e43407fe4552))
* **cmd/root.go:** handle case when config file does not exist by creating a new config file if it doesn't exist ([b3ca2ef](https://github.com/MohammadBnei/go-ai-cli/commit/b3ca2ef1f2d0dde87cb2601344e82451a3c3aabd))
* **cmd:** change flag name from --auto-save to --auto-load to improve semantics ([9e09ca0](https://github.com/MohammadBnei/go-ai-cli/commit/9e09ca074b6609f251e3228cabdcea4c77d082ef))
* **config.go:** change Find function to FindLastIndexOf to fix bug in CloseContextById method ([b02600a](https://github.com/MohammadBnei/go-ai-cli/commit/b02600aec569ff04eecd8f9348d7a2dc7600fd47))
* **keys.go:** add condition to only copy assistant's content if there are no chat messages in the stack to prevent copying empty content ([1ab6ede](https://github.com/MohammadBnei/go-ai-cli/commit/1ab6ede85b29365af7c543909c6825f391307093))
* **keys.go:** add key bindings for back and forward to navigate audio player position ([55d0b4a](https://github.com/MohammadBnei/go-ai-cli/commit/55d0b4aefb1567649021900120859192b72df825))
* **keys.go:** remove "esc" key from cancel binding to align with actual behavior ([8712ab2](https://github.com/MohammadBnei/go-ai-cli/commit/8712ab2c8c5631a5456d244a0312956ddacb0d13))
* **player.go:** move assignment of streamer variable before initializing speaker to prevent potential nil pointer error ([1ecf78c](https://github.com/MohammadBnei/go-ai-cli/commit/1ecf78c0e92467d7da2b30d74a80c9ebf693a10f))
* **prompt.go:** remove unused recover() function call to improve code readability and maintainability ([588f488](https://github.com/MohammadBnei/go-ai-cli/commit/588f488994a7436281896d17e250a19453438dd1))
* **test.yml:** update go-version to '^1.22' to match the required version ([ddc59bd](https://github.com/MohammadBnei/go-ai-cli/commit/ddc59bdf8aff4f33426831efb1b7044e1547a0f0))


### Features

* **agent:** added web search agent ([c608a65](https://github.com/MohammadBnei/go-ai-cli/commit/c608a651e5d82cdce695449e587d5bc03bf34717))
* **all:** major update ([a86098f](https://github.com/MohammadBnei/go-ai-cli/commit/a86098ff5d635c77aac115d0af14b28dd574b378))
* **all:** major update ([adf7eb5](https://github.com/MohammadBnei/go-ai-cli/commit/adf7eb52c336b2d23583c14c9812065af8a9dcb5))
* **all:** stable working chat ([3cca3ca](https://github.com/MohammadBnei/go-ai-cli/commit/3cca3ca36bfc700821de28d6a9c802391fc8aaba))
* **audio:** added vfs for audio ([2fcf701](https://github.com/MohammadBnei/go-ai-cli/commit/2fcf701e85a227a68fb3fa87f3ccb864f848458a))
* **chat.go:** add functionality to clear chat messages when CLEAR option is selected ([57817ce](https://github.com/MohammadBnei/go-ai-cli/commit/57817cef2752611187acd70c002dbc3a0a451531))
* **chat:** add support for auto-loading last chat when auto-load flag is set to true ([4004351](https://github.com/MohammadBnei/go-ai-cli/commit/40043519dc9aa574f222aaeef2c44044f168bd43))
* **image:** adding image generation ([a7b8bfa](https://github.com/MohammadBnei/go-ai-cli/commit/a7b8bfa4d719d33b33852003b89d3c2b10729b7e))
* **option-menu:** added an option menu as i ran out of key presses ([a4c9e10](https://github.com/MohammadBnei/go-ai-cli/commit/a4c9e10b35b4209a0d0f572fb8b6441eed56442d))

## [0.16.1](https://github.com/MohammadBnei/go-ai-cli/compare/0.16.0...0.16.1) (2024-02-05)

# [0.16.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.15.0...0.16.0) (2024-02-05)


### Bug Fixes

* **all:** added config editing capacities ([02b1289](https://github.com/MohammadBnei/go-ai-cli/commit/02b1289c15734ed4f335820f2f36f7b9c8dc8e1f))
* **build:** added mock speech cmd to permit build of app without portaudio ([960c930](https://github.com/MohammadBnei/go-ai-cli/commit/960c9303fccb1620555d6474003322a941217f11))
* **chat.go:** change TitleStyle width property to MaxWidth to correctly set the maximum width of the title style ([e5896a0](https://github.com/MohammadBnei/go-ai-cli/commit/e5896a040b825d8e4d4cfc4cf06aac8d8dcdf208))
* **chat.go:** remove unused code that sets the max width of the title style ([889cf4d](https://github.com/MohammadBnei/go-ai-cli/commit/889cf4d18533c7aeba64ef36cfa7517b3b1a1d86))
* **chat:** fixed an error when changing messages ([68595d7](https://github.com/MohammadBnei/go-ai-cli/commit/68595d748719eaeac2b2b464b3c7c2cdcab8e46e))
* **cmd/prompt.go:** remove unused variables and channels to improve code readability and maintainability ([b6673b0](https://github.com/MohammadBnei/go-ai-cli/commit/b6673b0a5b6caa8848571e4a6910dd1cf230bd67))
* **cmd/root.go:** remove unused import of "path/filepath" ([ee4ac4f](https://github.com/MohammadBnei/go-ai-cli/commit/ee4ac4f357791de92ddf06f8ff7c67b123fa95e2))
* **config.go:** change "config-path" to "configfile" to improve semantics ([530f829](https://github.com/MohammadBnei/go-ai-cli/commit/530f8291614992ced67c24913f71d30a6fc358d8))
* **config.go:** remove unused import of api package to improve code cleanliness and remove unused code block ([7efef7a](https://github.com/MohammadBnei/go-ai-cli/commit/7efef7ad55395b702c5077683b92eb21ea600288))
* **keys:** typo ([2979b48](https://github.com/MohammadBnei/go-ai-cli/commit/2979b4801f65d8523f497ad1938b8f8f186cb45a))
* **markdown-mode:** modified prompt config to use viper config ([c6d5bbe](https://github.com/MohammadBnei/go-ai-cli/commit/c6d5bbeaf25a4aafc585e529acaf2c972393e1b0))
* **message.go:** add condition to check if error is not nil when converting string to int in RemoveFn ([9a9d235](https://github.com/MohammadBnei/go-ai-cli/commit/9a9d23582926a5c98c2b79e8f1bccdc69522c2ac))
* **old-commands:** started cleaning old command. Adding a filepicker. Different styles fix ([babcc08](https://github.com/MohammadBnei/go-ai-cli/commit/babcc088be282c9e185a057dd9e296b3be028efe))
* **prompt.go:** replace fmt.Scanln with bufio.Scanner to handle user input with spaces correctly ([9c86caa](https://github.com/MohammadBnei/go-ai-cli/commit/9c86caa49afb2dbd8c6afd6c84d427cd1c107b76))
* **prompt:** reverted to promptui for now ([bba4422](https://github.com/MohammadBnei/go-ai-cli/commit/bba442290a99a23d48c417d66278a42e7a06cebd))
* **server.ts:** change port variable case from lowercase port to uppercase PORT to improve semantics ([aa43b48](https://github.com/MohammadBnei/go-ai-cli/commit/aa43b487edbc427dad27ef9356dde9ce83b8c860))
* **sizing:** fixed sizing issue on the chat ([1ba85c8](https://github.com/MohammadBnei/go-ai-cli/commit/1ba85c8b2f4ecd31ab7bcd9bb108c81c1a5282c9))
* **speech:** fixed pointer error and lang not being saved to speech config ([0e531c9](https://github.com/MohammadBnei/go-ai-cli/commit/0e531c96255fc5f8b8460d049037c2d756a263d0))
* **system.go:** update getDelegateFn function to display correct "Added" status for chat messages ([9479ec3](https://github.com/MohammadBnei/go-ai-cli/commit/9479ec37c34a6353701413fd8e8096e865692f08))
* **ui/file.go:** update file loop condition to include files with .svelte extension in the search results ([e27b5dd](https://github.com/MohammadBnei/go-ai-cli/commit/e27b5dd310267d51d977f13c3d60b5ffa7c70e91))
* **ui/system.go:** handle case when system is not found in systems map by returning an error event ([7ee1eb5](https://github.com/MohammadBnei/go-ai-cli/commit/7ee1eb54cffb5c5f445cd0440b2163f2e3582d01))
* **ui:** replace ioutil.ReadDir with os.ReadDir to use the updated function in Go 1.16 ([901ff10](https://github.com/MohammadBnei/go-ai-cli/commit/901ff10a5ebfc54aefffecec9ef2a5bb7ddfa710))


### Features

* **api:** add config.go file to handle API configuration ([fb4b9d0](https://github.com/MohammadBnei/go-ai-cli/commit/fb4b9d00e70a09fe861977885944308bb1514d33))
* **bubble:** functionnal ([ffaa3ce](https://github.com/MohammadBnei/go-ai-cli/commit/ffaa3cec4df76a805cd106cd6c8cd822d323d9e2))
* **bubble:** stable ([a4c12b9](https://github.com/MohammadBnei/go-ai-cli/commit/a4c12b9f887dc568abb001b75183cdae6ccd0292))
* **chatUpdate.go:** refactor reset function to use getInfoContent function to generate AI response content ([3417b57](https://github.com/MohammadBnei/go-ai-cli/commit/3417b5749d6fb94fad2f4b1b91d2c1b2b96053c9))
* **codebase:** full update of the architecure of the project ([a11e529](https://github.com/MohammadBnei/go-ai-cli/commit/a11e5294e19a4166b847c8479f059e8e80c0cba9))
* **command, markdown, ui:** add new command "responses" to display previous messages, add method "ToMarkdown" to convert text to markdown format, add method "ShowPreviousMessage" to show previous messages with markdown support ([08b8b84](https://github.com/MohammadBnei/go-ai-cli/commit/08b8b84b44a56d18919d29baddd8289d8d4e4898))
* **command, prompt, ui:** add support for 'cli-clear' command to clear the terminal screen ([a006028](https://github.com/MohammadBnei/go-ai-cli/commit/a0060286948399abbdbf07bd645e2117dcee086b))
* **complete:** messages handling ([d0bd742](https://github.com/MohammadBnei/go-ai-cli/commit/d0bd7425c06ac139aceada4cdbc679fb4b7f5c6d))
* **continuous:** working on continuous speech mode ([eeab278](https://github.com/MohammadBnei/go-ai-cli/commit/eeab278383c53776c5cadb2240ae3f3468d7c771))
* **help:** added help for main menu ([3acbb9c](https://github.com/MohammadBnei/go-ai-cli/commit/3acbb9c6a3a5df2f2af08593daa7dc91569f2e30))
* **langchain:** starting to implement langchain function ([f127802](https://github.com/MohammadBnei/go-ai-cli/commit/f127802b2b54c5de3f769440c20cc375a8e624f4))
* **message.go:** add role selection to the message editing form ([8a7404f](https://github.com/MohammadBnei/go-ai-cli/commit/8a7404f0a7b44cd088c467d4eb75778878775df0))
* **name:** changed name from openai to ai ([a845e22](https://github.com/MohammadBnei/go-ai-cli/commit/a845e2290f7a76c4c9895348e2af0b41d84964c5))
* **name:** changed name from openai to ai ([8bce3ad](https://github.com/MohammadBnei/go-ai-cli/commit/8bce3ade7163b9483242c4414db7e82f2cd69182))
* **prompt:** finally found a stable prompt ([c798320](https://github.com/MohammadBnei/go-ai-cli/commit/c798320dbfdd408340f4905431bd8b9d3da07999))
* **styles:** adding style centralized file ([d8fc9a9](https://github.com/MohammadBnei/go-ai-cli/commit/d8fc9a9ded4c2882975eb965618297d9dcf9fb58))

# [0.15.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.14.0...0.15.0) (2023-08-07)


### Features

* **speech.go:** add flag completion for "lang" flag to provide autocomplete options for language selection ([16a2517](https://github.com/MohammadBnei/go-ai-cli/commit/16a2517e9b48543d3cee0054b308516460128e64))

# [0.14.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.13.1...0.14.0) (2023-08-07)


### Bug Fixes

* **speech.go:** change function name from loadContext to LoadContext to follow Go naming conventions ([9382275](https://github.com/MohammadBnei/go-ai-cli/commit/93822754b385df697456ee4cd6444bcbafe48061))


### Features

* **audio.go:** add SendAudio function to send the recorded audio file to OpenAI for transcription ([84cec1c](https://github.com/MohammadBnei/go-ai-cli/commit/84cec1cd06fb9a884bb32ad595b23620e067b7fd))
* **cmd/file.go:** add new 'file' command to convert an audio file to text ([25c1ff1](https://github.com/MohammadBnei/go-ai-cli/commit/25c1ff1cac2cfb38bfbd5fe63e6af210c4605177))
* **cmd/record.go:** add 'record' command to the 'speech' command group ([66c00bf](https://github.com/MohammadBnei/go-ai-cli/commit/66c00bf5b88182acaf69a3f95f8f9f4dfd8861a3))

## [0.13.1](https://github.com/MohammadBnei/go-ai-cli/compare/0.13.0...0.13.1) (2023-08-03)


### Bug Fixes

* **banner:** put banner message in command Run instead of root, because it corrupted the completion output ([8f6361c](https://github.com/MohammadBnei/go-ai-cli/commit/8f6361cede93a0466a8d6fc5884c8f5cd32ccbb6))

# [0.13.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.12.1...0.13.0) (2023-08-03)


### Features

* **speech.go:** add support for auto saving speech to a file ([825f636](https://github.com/MohammadBnei/go-ai-cli/commit/825f636973f66bd0b90e48ad8a92d9172d1b2281))

## [0.12.1](https://github.com/MohammadBnei/go-ai-cli/compare/0.12.0...0.12.1) (2023-08-02)


### Bug Fixes

* **terminal:** removed clear screen command ([11446f3](https://github.com/MohammadBnei/go-ai-cli/commit/11446f3ddb728f59961e2765a5296ffc1c226763))

# [0.12.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.11.0...0.12.0) (2023-08-02)


### Bug Fixes

* **audio:** adding en by default for audio record ([87470e3](https://github.com/MohammadBnei/go-ai-cli/commit/87470e3a8148aa708dabf5268370f49245544681))
* **prompt:** fixed forgotten update of the promptConfig with user prompt ([49b57f1](https://github.com/MohammadBnei/go-ai-cli/commit/49b57f104090e28735dcdbbd7d7d57ddca44addc))
* **speech.go:** add missing import for io package ([a128df0](https://github.com/MohammadBnei/go-ai-cli/commit/a128df061a8a82d4b63634d3844b035611ed1036))
* **speech.go:** change maxMinutes value from 5 to 4 ([82e577d](https://github.com/MohammadBnei/go-ai-cli/commit/82e577dd09ae31400e96bf790ff4776d525b6209))


### Features

* **audio.go:** add language parameter to SpeechToText function ([dba93d0](https://github.com/MohammadBnei/go-ai-cli/commit/dba93d05567d0a48aa88ab0a05fa20b4352fc73a))
* **audio:** audio and normal usage now joint by build args ([8774be0](https://github.com/MohammadBnei/go-ai-cli/commit/8774be0b9d0974f04fe0b875fc6a404bb34652b2))
* **cmd/speech:** add speech command to convert speech to text ([7b92b05](https://github.com/MohammadBnei/go-ai-cli/commit/7b92b051b279247f18d3a042590a39ba7f21a372))
* **mask:** adding hugging face mask abilities ([4f808f4](https://github.com/MohammadBnei/go-ai-cli/commit/4f808f4ef6531c481569718ed51bfe86fef3c648))
* **prompt:** changed way to run cmd ([f40c1dc](https://github.com/MohammadBnei/go-ai-cli/commit/f40c1dc20fd15e97d39f8c9aa15984ad9f9df865))
* **speech:** implementig speech to text ([a4a75bc](https://github.com/MohammadBnei/go-ai-cli/commit/a4a75bc4f4fbcbbef0eeeb9b651812ed16250b87))

# [0.11.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.10.2...0.11.0) (2023-07-24)


### Bug Fixes

* **cmd/config.go:** change variable name 'path' to 'filePath' for clarity ([907ee8b](https://github.com/MohammadBnei/go-ai-cli/commit/907ee8ba7cf35b5dc5e7354b6770adfdf1a70407))
* **markdown:** Fixed backtick error ([4dc3a83](https://github.com/MohammadBnei/go-ai-cli/commit/4dc3a839355e6d6a3e099561728326ba28004054))
* **md-format:** utilizing "md" instead of !md ([0ed38c3](https://github.com/MohammadBnei/go-ai-cli/commit/0ed38c346e70715bf4c2ffc16a1cbc064ad87e02))
* **writer.go:** add support for single backticks in Write method ([71d4caf](https://github.com/MohammadBnei/go-ai-cli/commit/71d4cafa4fd4304ff5befb8eb28e4dab72108f2a))


### Features

* **format:** adding support for markdown format in terminal ([6b4967a](https://github.com/MohammadBnei/go-ai-cli/commit/6b4967af3b01943e2228a9e6f5d867d63584e317))

## [0.10.2](https://github.com/MohammadBnei/go-ai-cli/compare/0.10.1...0.10.2) (2023-06-28)


### Bug Fixes

* **config.go:** change default value of "messages-length" flag from 10 to 20 ([42e166c](https://github.com/MohammadBnei/go-ai-cli/commit/42e166c76ebef285419fbc0b3273470caed00c0d))

## [0.10.1](https://github.com/MohammadBnei/go-ai-cli/compare/0.10.0...0.10.1) (2023-05-26)

# [0.10.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.9.0...0.10.0) (2023-05-26)


### Features

* **main.go:** add support for debugging with pprof ([f29a2e1](https://github.com/MohammadBnei/go-ai-cli/commit/f29a2e1b2bf90f17710baa966a05541d9ad42a5a))

# [0.9.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.8.0...0.9.0) (2023-05-25)


### Bug Fixes

* **openai.go, prompt.go:** add context cancellation to SendPrompt function ([09bc6e6](https://github.com/MohammadBnei/go-ai-cli/commit/09bc6e686e95e22b7589eea4b1955f321ac0a4d3))


### Features

* **config:** adding messages retention numberto config ([7359ac9](https://github.com/MohammadBnei/go-ai-cli/commit/7359ac958b999c30990fb8a79281a2b995859983))
* **config:** adding messages retention numberto config ([5a9af62](https://github.com/MohammadBnei/go-ai-cli/commit/5a9af6231672d9829213c2701e74c5a97d55c198))
* **copy:** added clipboard copy capacities ([35b07c5](https://github.com/MohammadBnei/go-ai-cli/commit/35b07c5eb000b3d43f0c8e86b19fa674151d02bc))
* **image:** adding image generation capabilities ([c71aa6d](https://github.com/MohammadBnei/go-ai-cli/commit/c71aa6d2da60621fdfd1738a27fa8d06fb539b21))

# [0.8.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.7.0...0.8.0) (2023-05-15)


### Features

* **file:** tree view when hovering folder ([97f122c](https://github.com/MohammadBnei/go-ai-cli/commit/97f122cbd99fab8bba85914ddd8a072e638c8872))
* **test.yml:** add paths filter for Go files ([3336ee9](https://github.com/MohammadBnei/go-ai-cli/commit/3336ee95418b045e23481c93c84a03595236fe45))

# [0.7.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.6.0...0.7.0) (2023-05-15)


### Bug Fixes

* **prompt-help:** updated the help section ([55a2cf9](https://github.com/MohammadBnei/go-ai-cli/commit/55a2cf954ef86eb9d5ff93d7066a6505f04ddfc8))


### Features

* **config:** add support for environment variable CONFIG ([48dedf1](https://github.com/MohammadBnei/go-ai-cli/commit/48dedf113387436e3ec6d1f9a1a99eef6106c183))

# [0.6.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.5.0...0.6.0) (2023-05-15)


### Features

* **client-release.yml:** add fail-fast option to matrix strategy ([b53e021](https://github.com/MohammadBnei/go-ai-cli/commit/b53e02147fd201e5068ccca55f1fa936625e782e))
* **prompt.go:** add support for fuzzy file search and multi-file selection ([de9e24a](https://github.com/MohammadBnei/go-ai-cli/commit/de9e24a260c896d5403887a42a8edb22d2b357ce))

# [0.5.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.4.0...0.5.0) (2023-05-15)


### Features

* add Dockerfile and GitHub Actions workflow for building and pushing Docker image ([d58cebc](https://github.com/MohammadBnei/go-ai-cli/commit/d58cebc12c5dddefffd763dfce55ef9a358771fb))

# [0.4.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.3.2...0.4.0) (2023-05-15)


### Bug Fixes

* **config:** create config directory if it does not exist ([61f20b8](https://github.com/MohammadBnei/go-ai-cli/commit/61f20b84013929d01c940a738dcdc48b41c035e5))


### Features

* add zsh completion script ([5895130](https://github.com/MohammadBnei/go-ai-cli/commit/5895130b23d1838f68c89197237c17dfd036f7dd))
* add zsh completion script ([8b2708b](https://github.com/MohammadBnei/go-ai-cli/commit/8b2708ba048c7f910743a6ddfe63c54194fa4efe))

## [0.3.2](https://github.com/MohammadBnei/go-ai-cli/compare/0.3.1...0.3.2) (2023-05-15)

## [0.3.1](https://github.com/MohammadBnei/go-ai-cli/compare/0.3.0...0.3.1) (2023-05-15)


### Bug Fixes

* **node_modules:** removed node modules and pnpm.lock ([c487bd5](https://github.com/MohammadBnei/go-ai-cli/commit/c487bd5b037148572e609e8f2e6ae3a968af70cf))

# [0.3.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.2.0...0.3.0) (2023-05-14)


### Features

* **openai.go:** add ClearMessages function ([3c84b06](https://github.com/MohammadBnei/go-ai-cli/commit/3c84b06cc434d12d51c87bf7df322849bab5f17d))

# [0.2.0](https://github.com/MohammadBnei/go-ai-cli/compare/0.1.1...0.2.0) (2023-05-14)


### Features

* **release:** testing automated release ([a0b300a](https://github.com/MohammadBnei/go-ai-cli/commit/a0b300ad54d93eeb6fbf3b2c2e13eca1b9a17418))

## [0.1.1](https://github.com/MohammadBnei/go-ai-cli/compare/0.1.0...0.1.1) (2023-05-14)

# 0.1.0 (2023-05-14)


### Features

* add license, cmd/config.go, cmd/prompt.go, cmd/root.go, gowatch.yml, main.go, service/openai.go ([efd62fc](https://github.com/MohammadBnei/go-ai-cli/commit/efd62fcf7cb62998e689f8b19f561fddca18fb47))
* **client-release.yml:** add GitHub workflow for releasing Go binary ([2dc52bc](https://github.com/MohammadBnei/go-ai-cli/commit/2dc52bca56203f7e5324cc46243f7088d44f4675))
* **openai.go, prompt.go:** add support for adding file contents to prompt ([9b24376](https://github.com/MohammadBnei/go-ai-cli/commit/9b2437606686bb7a1a35418d5fea2d830fc1e522))
* **project:** adding release it and updating go module ([a031471](https://github.com/MohammadBnei/go-ai-cli/commit/a03147195008f7335daba0415d8a1f37f3e2306a))