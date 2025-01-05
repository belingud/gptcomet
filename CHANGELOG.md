
---
## [0.1.7](https://github.com/belingud/gptcomet/compare/v0.1.6..v0.1.7) - 2025-01-05

### ⛰️  Features

- add manual provider input, improve config input - ([3cbb9ec](https://github.com/belingud/gptcomet/commit/3cbb9ec8ce920c2fd63cf41baf6fa4ce76610227)) - belingud

### 📚 Documentation

- update README with install script and gif - ([bbf5941](https://github.com/belingud/gptcomet/commit/bbf5941eb4d35658701eca6c2bf04bd02c4d41ba)) - belingud
- update changelog for release 0.1.6 - ([7cb49ad](https://github.com/belingud/gptcomet/commit/7cb49ad3c90800f2eeffac3444bc9d5c086d5aa8)) - belingud

### ⚙️ Miscellaneous Tasks

- add pygmsg command to pyproject.toml - ([68f4366](https://github.com/belingud/gptcomet/commit/68f4366f902113095cf9863d129d772e380a0efe)) - belingud
- update gitignore to exclude test binaries - ([3fb4da6](https://github.com/belingud/gptcomet/commit/3fb4da6cf5cb491fd7d67182fede4d4e6a19e55c)) - belingud


---
## [0.1.6](https://github.com/belingud/gptcomet/compare/v0.1.6-dev..v0.1.6) - 2025-01-05

### ⚙️ Miscellaneous Tasks

- Update release workflow - ([f3778be](https://github.com/belingud/gptcomet/commit/f3778befbd6332cd5264843801b5e092c39f2df9)) - belingud


---
## [0.1.5](https://github.com/belingud/gptcomet/compare/v0.1.4..v0.1.5) - 2024-12-28

### ⛰️  Features

- update default model to gpt-4o - ([57b1bd8](https://github.com/belingud/gptcomet/commit/57b1bd8c9e72159c868093e5a0051505a3214016)) - belingud

### 🚜 Refactor

- enhance logging and update config handling - ([dac110f](https://github.com/belingud/gptcomet/commit/dac110fa73614ad20acd0c0aa95c54539acbb591)) - belingud
- simplify log level setting - ([c9bead4](https://github.com/belingud/gptcomet/commit/c9bead47cfb5434c4f50fc7d97661742585aed18)) - belingud

### 📚 Documentation

- update answer path in README - ([ee5a9e4](https://github.com/belingud/gptcomet/commit/ee5a9e4bb291624cc5e46f1fe667e622b11b6422)) - belingud
- update install guide and supported languages - ([d479c30](https://github.com/belingud/gptcomet/commit/d479c30d111beb8ac3aad8c36f37560c551c6002)) - belingud
- update README with uv install and license - ([423a78a](https://github.com/belingud/gptcomet/commit/423a78a326d8c754c5c2b3169b9989c005c6c967)) - belingud


---
## [0.1.4](https://github.com/belingud/gptcomet/compare/v0.1.3..v0.1.4) - 2024-12-22

### 🐛 Bug Fixes

- handle NoSuchProvider error in config - ([50a6f8d](https://github.com/belingud/gptcomet/commit/50a6f8d5aeaf93b66932b438c6e14e93bd8a91fa)) - belingud

### 🚜 Refactor

- remove try/except block in create_provider_config - ([73476a8](https://github.com/belingud/gptcomet/commit/73476a81769e48041837d2b9356b520b2e7c42d3)) - belingud

### 🎨 Styling

- fix spacing in commit message template - ([10221bd](https://github.com/belingud/gptcomet/commit/10221bdc1db7486f88b600164474d78212ad326f)) - belingud

### 🧪 Testing

- use requests and fix test assertions - ([bc2eba9](https://github.com/belingud/gptcomet/commit/bc2eba9628dea5c1e4f5281f5458697cae2d0719)) - belingud

### ⚙️ Miscellaneous Tasks

- comment out windows build in github action - ([6898774](https://github.com/belingud/gptcomet/commit/68987743de25af7362f5488fd4f522d0b83914ee)) - belingud
- simplify windows build workflow - ([651273e](https://github.com/belingud/gptcomet/commit/651273ef79795410908b544f4998e659280a0551)) - belingud
- improve build workflow with Nuitka - ([012d4ff](https://github.com/belingud/gptcomet/commit/012d4ffa31dca7d8a68f55cafac1e712d0f2a521)) - belingud
- Improve build and release workflow - ([ee1122e](https://github.com/belingud/gptcomet/commit/ee1122e238d32daaf38359be397a6c9893a236b1)) - belingud
- update release file pattern - ([06be8e1](https://github.com/belingud/gptcomet/commit/06be8e19ba70b76b0b5f2071d976d97935965b73)) - belingud
- use github.ref_name for build output - ([b237421](https://github.com/belingud/gptcomet/commit/b237421a25b69fceab6fed9345847bd5ec763b4a)) - belingud
- add tag to build output and test executable - ([d19f078](https://github.com/belingud/gptcomet/commit/d19f078ae2f1bf0798e333180e0c7e6f1237d885)) - belingud
- use git-cliff to generate release notes - ([06837fc](https://github.com/belingud/gptcomet/commit/06837fcd9bf4fc57a31dab3242ad247e7eed4627)) - belingud
- add release notes generation and notification - ([f04074e](https://github.com/belingud/gptcomet/commit/f04074e7781524442109fce5a820ca5dca2a3051)) - belingud
- format nuitka command - ([4f38133](https://github.com/belingud/gptcomet/commit/4f381331a80ea846fe37d549889b48c3e72242ab)) - belingud
- remove unused dev and build dependencies - ([3815e98](https://github.com/belingud/gptcomet/commit/3815e988032a9d5744134e02d78de53af5a77508)) - belingud
- add build workflow for executables - ([c9ed91a](https://github.com/belingud/gptcomet/commit/c9ed91a706799f04ff38680feb215e30cda02f71)) - belingud
- update imports and build system, replace httpx with requests, modify config and LLM handling - ([a88124d](https://github.com/belingud/gptcomet/commit/a88124dc1e3afc695bd31a3158f6ad1c14b2eea2)) - belingud
- bump version to 0.1.3 - ([123ba6a](https://github.com/belingud/gptcomet/commit/123ba6accdd6e3e4804f9eb67d2844948d9ea87a)) - belingud

### Build

- Add nuitka build, update publish, uv install docs - ([9d1d2b0](https://github.com/belingud/gptcomet/commit/9d1d2b098c430e9ea47ea2b6486ffe0f2b36ba81)) - belingud


---
## [0.1.3](https://github.com/belingud/gptcomet/compare/v0.1.1..v0.1.3) - 2024-12-19

> !Important
> Add support for Deepseek, Kimi, Silicon and other LLM providers

### ⛰️  Features

- add presence_penalty and other input validation - ([f3cff3f](https://github.com/belingud/gptcomet/commit/f3cff3f08b31db1e7626525fae2875e5415d0235)) - belingud
- add debug logging for proxy and requests - ([217f933](https://github.com/belingud/gptcomet/commit/217f9338c66e02be7e1169243d906868941813ba)) - belingud
- add Deepseek, Kimi and Silicon LLM providers - ([00e6229](https://github.com/belingud/gptcomet/commit/00e62297a00dfca1bd5d7c43a4f5d3f9f2b132fa)) - belingud
- add CLI options and provider config overrides - ([8613f18](https://github.com/belingud/gptcomet/commit/8613f1878951ad1c7618cf396094b7f6b6904c65)) - belingud

### 🐛 Bug Fixes

- remove traceback print from commit cli - ([a52ae4a](https://github.com/belingud/gptcomet/commit/a52ae4aa104cd89075a36cd14483d9bde81936f7)) - belingud

### 🚜 Refactor

- replace logger with console print in config clis - ([f121ca5](https://github.com/belingud/gptcomet/commit/f121ca5af03df1a539a0412d43e818fa439331d8)) - belingud
- improve config handling and error management - ([c05f8a4](https://github.com/belingud/gptcomet/commit/c05f8a4a79c4cbc7e1eddda2a662d7a55ce9d08d)) - belingud
- improve provider selection and config management - ([aa1d28a](https://github.com/belingud/gptcomet/commit/aa1d28a2f17e2fce392abe802333c7109b3c5778)) - belingud
- support gemini and other provider - ([03407ae](https://github.com/belingud/gptcomet/commit/03407ae2d968eb4ff35b0ff9270de04d8194c0de)) - belingud

### 📚 Documentation

- update README for new provider setup - ([6aa77fe](https://github.com/belingud/gptcomet/commit/6aa77fefc2518aa2e6e1f104f3734cbfbbaa2eb1)) - belingud

### 🎨 Styling

- fix minor whitespace issues in tests - ([41bd91a](https://github.com/belingud/gptcomet/commit/41bd91abea6e6f7408a9239fd9713042e70932fa)) - belingud

### 🧪 Testing

- add multiple test files for LLM implementations - ([1489781](https://github.com/belingud/gptcomet/commit/1489781f2e939ea4f30d1862ccd85b99fcb43b03)) - belingud

### Build

- bump version to 0.1.1 - ([0939fc6](https://github.com/belingud/gptcomet/commit/0939fc63d7878917c7d75df43d0685ceec4bf575)) - belingud


---
## [0.1.1](https://github.com/belingud/gptcomet/compare/v0.1.0..v0.1.1) - 2024-12-12

### ⛰️  Features

- add API key masking to config retrieval - ([a886726](https://github.com/belingud/gptcomet/commit/a886726ddf4fdbeec41991a9cdb74355a6c08863)) - belingud

### 📚 Documentation

- update README table of contents - ([5e248b9](https://github.com/belingud/gptcomet/commit/5e248b98a5f75526bfb23d7dab40b340b47b445b)) - belingud
- update changelog for v0.1.0 release - ([e418a4c](https://github.com/belingud/gptcomet/commit/e418a4c09f40bd072244c184e79b0c92d4488d3c)) - belingud


---
## [0.1.0](https://github.com/belingud/gptcomet/compare/v0.0.24..v0.1.0) - 2024-12-12

### ⛰️  Features

- add new languages to documentation and mapping - ([06820d2](https://github.com/belingud/gptcomet/commit/06820d224c852f05bb83672e550cd2292264a748)) - belingud
- add default values for GPT configuration - ([d82e106](https://github.com/belingud/gptcomet/commit/d82e106a2204221da15a4d0839a7daaad0e33a2a)) - belingud
- add verbose logging and translation support - ([d77bf9c](https://github.com/belingud/gptcomet/commit/d77bf9c64a3b27e86b8001a2908641faa17b81e7)) - belingud
- enhance API key masking and CLI interaction - ([03dc20e](https://github.com/belingud/gptcomet/commit/03dc20ef913438efd126e6d7fd07aaa077a338ee)) - belingud

### 🐛 Bug Fixes

- correct package version in .bumpversion.cfg - ([018c170](https://github.com/belingud/gptcomet/commit/018c1700b531a0aa16e394c86e513f7b7e7122f7)) - belingud
- correct package version - ([0a492b6](https://github.com/belingud/gptcomet/commit/0a492b6db98c2a0b92e5f2ad3d7d9c34e31b9d1b)) - belingud

### 🚜 Refactor

- remove unused variable in commit entry - ([f00abb6](https://github.com/belingud/gptcomet/commit/f00abb6324f7c77ebbad1e3783307387c700710a)) - belingud
- handle empty inputs in commit and style functions - ([a3f88a2](https://github.com/belingud/gptcomet/commit/a3f88a2230111d94a1c99421b350c590aa21264a)) - belingud
- improve commit message console output readability - ([834cc8f](https://github.com/belingud/gptcomet/commit/834cc8f8025dc68b7b2e7c92e4c3a72c09a5044f)) - belingud

### 📚 Documentation

- update README with new configuration options - ([0e1feaf](https://github.com/belingud/gptcomet/commit/0e1feaffe1c0b80998a043c1e24fb6a3c4b5366e)) - belingud
- remove documentation files - ([117d511](https://github.com/belingud/gptcomet/commit/117d511091560687d1194793c17a1485a1b840f9)) - belingud

### 🧪 Testing

- add more detailed test cases - ([8b30e55](https://github.com/belingud/gptcomet/commit/8b30e5542b570f4befd4009b6f6eb49796bf2fd3)) - belingud
- fix and add comments in api key tests - ([cea0399](https://github.com/belingud/gptcomet/commit/cea03990a7454d9d3498705fc05ef98383aaaf00)) - belingud

### Build

- update version to 0.1.0 - ([b06b47b](https://github.com/belingud/gptcomet/commit/b06b47ba7edfc4d99210cf342f702839c470e5fb)) - belingud


---
## [0.1.0](https://github.com/belingud/gptcomet/compare/v0.0.24..v0.1.0) - 2024-12-11

### ⛰️  Features

- add new languages to documentation and mapping - ([06820d2](https://github.com/belingud/gptcomet/commit/06820d224c852f05bb83672e550cd2292264a748)) - belingud
- add default values for GPT configuration - ([d82e106](https://github.com/belingud/gptcomet/commit/d82e106a2204221da15a4d0839a7daaad0e33a2a)) - belingud
- add verbose logging and translation support - ([d77bf9c](https://github.com/belingud/gptcomet/commit/d77bf9c64a3b27e86b8001a2908641faa17b81e7)) - belingud
- enhance API key masking and CLI interaction - ([03dc20e](https://github.com/belingud/gptcomet/commit/03dc20ef913438efd126e6d7fd07aaa077a338ee)) - belingud

### 🐛 Bug Fixes

- correct package version in .bumpversion.cfg - ([018c170](https://github.com/belingud/gptcomet/commit/018c1700b531a0aa16e394c86e513f7b7e7122f7)) - belingud
- correct package version - ([0a492b6](https://github.com/belingud/gptcomet/commit/0a492b6db98c2a0b92e5f2ad3d7d9c34e31b9d1b)) - belingud

### 🚜 Refactor

- improve commit message console output readability - ([834cc8f](https://github.com/belingud/gptcomet/commit/834cc8f8025dc68b7b2e7c92e4c3a72c09a5044f)) - belingud

### 📚 Documentation

- update README with new configuration options - ([0e1feaf](https://github.com/belingud/gptcomet/commit/0e1feaffe1c0b80998a043c1e24fb6a3c4b5366e)) - belingud
- remove documentation files - ([117d511](https://github.com/belingud/gptcomet/commit/117d511091560687d1194793c17a1485a1b840f9)) - belingud

### 🧪 Testing

- fix and add comments in api key tests - ([cea0399](https://github.com/belingud/gptcomet/commit/cea03990a7454d9d3498705fc05ef98383aaaf00)) - belingud


---
## [0.0.23](https://github.com/belingud/gptcomet/compare/v0.0.22..v0.0.23) - 2024-11-29

### 🐛 Bug Fixes

- handle user cancellation and update proxy handling in provider config - ([519a505](https://github.com/belingud/gptcomet/commit/519a505d6ec9d2237b5c6a29cba21317c1d385e4)) - belingud
- enhance commit message editing with multi-line support and VIM mode - ([e3a57ee](https://github.com/belingud/gptcomet/commit/e3a57ee13c6428964bbacbc49c77101f6d1e8fd4)) - belingud

### 🚜 Refactor

- remove unused import and simplify provider config - ([da9b457](https://github.com/belingud/gptcomet/commit/da9b457419cb18bca51cfe4a8ba49440c6969abd)) - belingud

### 📚 Documentation

- update README.md and provider.py for new provider setup - ([46da242](https://github.com/belingud/gptcomet/commit/46da242b3810eb50e947ac3ed3c6a9dcff286540)) - belingud
- update README.md for setup, usage, and configuration - ([a44a853](https://github.com/belingud/gptcomet/commit/a44a853fa152ed376e414d309658ccb14c4820cf)) - belingud
- Update README.md for GPTComet configuration and features - ([46138b8](https://github.com/belingud/gptcomet/commit/46138b81976b50645ff4ac1660d0926e1dd979f6)) - belingud

### 🧪 Testing

- correct proxy argument name in test_llm_client.py - ([0031ba9](https://github.com/belingud/gptcomet/commit/0031ba91ff3908b4d99e681bc41bd814390474b3)) - belingud
- clean up tests and adjust config manager mock - ([397b86c](https://github.com/belingud/gptcomet/commit/397b86c36752176678e20992f3a67f32b9dc1001)) - belingud
- add unit tests for commit, log, and validator - ([da7934e](https://github.com/belingud/gptcomet/commit/da7934e3ad4fcb7bc1f7ad592d230f9f27af31a7)) - belingud
- improve readability in test cases - ([a867e63](https://github.com/belingud/gptcomet/commit/a867e6379eaad85f8bc3ff26e6ab33b761beec47)) - belingud

### ⚙️ Miscellaneous Tasks

- update pyrightconfig settings - ([3e67d61](https://github.com/belingud/gptcomet/commit/3e67d61f38c543b490a967f7c3dbedee2ae244a0)) - belingud

### Build

- bump version to 0.0.22 - ([749a7e4](https://github.com/belingud/gptcomet/commit/749a7e4ee917e6a447c400cafd2591be3b69abdc)) - belingud


---
## [0.0.22](https://github.com/belingud/gptcomet/compare/v0.0.21..v0.0.22) - 2024-11-26

### 🐛 Bug Fixes

- simplify api key masking in config manager - ([e872e1c](https://github.com/belingud/gptcomet/commit/e872e1c7a39edb49faf4d0b181bd60341ed16a0d)) - belingud

### 📚 Documentation

- update CHANGELOG for v0.0.21 - ([2ffd432](https://github.com/belingud/gptcomet/commit/2ffd4326ffc6e6ba180490d8a1c29f1acaa9d47d)) - belingud

### 🧪 Testing

- enhance mock config for LLM tests - ([93ee469](https://github.com/belingud/gptcomet/commit/93ee46999f2e32df0a0a969e836ccf00a3d4b500)) - belingud

### Build

- bump version to 0.0.21 - ([f97d183](https://github.com/belingud/gptcomet/commit/f97d183c82c00414f8907c1c790503ae2b067ddd)) - belingud


---
## [0.0.21](https://github.com/belingud/gptcomet/compare/v0.0.20..v0.0.21) - 2024-11-26

### 🐛 Bug Fixes

- improve error messaging in config removal - ([fec93b7](https://github.com/belingud/gptcomet/commit/fec93b79906c42095885da9d0c81d51f29929edc)) - belingud

### 🚜 Refactor

- remove unused commit hooks - ([477e2b1](https://github.com/belingud/gptcomet/commit/477e2b1233ed62e6415d139be2a90c1de6633f46)) - belingud

### ⚙️ Miscellaneous Tasks

- add uv.lock to default ignored files - ([b0a7c27](https://github.com/belingud/gptcomet/commit/b0a7c27a73162a7e808150197116c574667ed521)) - belingud

### Build

- update version to 0.0.20 - ([4924647](https://github.com/belingud/gptcomet/commit/49246478125094956576fcdf6442e8dc89ed3fea)) - belingud


---
## [0.0.20](https://github.com/belingud/gptcomet/compare/v0.0.19..v0.0.20) - 2024-11-25

### 🚜 Refactor

- optimize imports and internal structures - ([0615d8d](https://github.com/belingud/gptcomet/commit/0615d8d6de3a3062147e107d183ee1570e249bc0)) - belingud
- reorganize test dependencies in pyproject.toml - ([8b80013](https://github.com/belingud/gptcomet/commit/8b8001390f4d1b1acf9636791f3a254561b6f5b5)) - belingud
- improve error handling in commit process - ([bed8f5d](https://github.com/belingud/gptcomet/commit/bed8f5dfb30baf91bdc6e752c3884211d6b82606)) - belingud

### ⚙️ Miscellaneous Tasks

- remove GPTComet hook and prepare-commit-msg script - ([1bf52f9](https://github.com/belingud/gptcomet/commit/1bf52f9484cd5ba64a927de34cb7d2c4cc5cc6a5)) - belingud
- update project metadata and linting scope - ([7238915](https://github.com/belingud/gptcomet/commit/7238915c299d0f7920bdcd08d4b7c04d566968dc)) - belingud
- remove update_changelog script - ([fcc9884](https://github.com/belingud/gptcomet/commit/fcc9884cb4eb903d0ab58107f944360423dd8336)) - belingud

### Build

- update tox environments and tools - ([7c4006b](https://github.com/belingud/gptcomet/commit/7c4006b2faff47459678b157b2b192a16d545425)) - belingud

### Version

- update version to 0.0.19 - ([01a305a](https://github.com/belingud/gptcomet/commit/01a305ac64e29ecc68fa07163e022bad319999f8)) - belingud


---
## [0.0.19](https://github.com/belingud/gptcomet/compare/v0.0.18..v0.0.19) - 2024-11-21

### ⛰️  Features

- add model and provider print, change token color - ([f0f7127](https://github.com/belingud/gptcomet/commit/f0f712703159647637bce888db7b60e2fcf64caf)) - belingud

### 🚜 Refactor

- improve code readability and error handling - ([d0cf65d](https://github.com/belingud/gptcomet/commit/d0cf65d212faee6287bc4d7243c66077419933bf)) - belingud

### 📚 Documentation

- update changelog for version 0.0.18 - ([b1a5510](https://github.com/belingud/gptcomet/commit/b1a551070348ddf9307552eadcb024c624c7a7e4)) - belingud

### ⚙️ Miscellaneous Tasks

- disable auto commit in bumpversion config - ([6e352eb](https://github.com/belingud/gptcomet/commit/6e352ebdb7b643dda90aab814f6e04997edd4054)) - belingud
- update pdm lock - ([9dc802f](https://github.com/belingud/gptcomet/commit/9dc802f5ac4c84a176dddfd22de59691453f02a2)) - belingud

### Build

- update and restructure dev dependencies - ([49f541d](https://github.com/belingud/gptcomet/commit/49f541d4cb6150918b0945c6a535675644812979)) - belingud


---
## [0.0.18](https://github.com/belingud/gptcomet/compare/v0.0.17..v0.0.18) - 2024-11-19

### ⛰️  Features

- add loading message in LLMClient - ([d7bf851](https://github.com/belingud/gptcomet/commit/d7bf851d0e3fb92d94bce8fe33941ff4c03ab9a8)) - belingud
- add retry choices and improve commit message generation - ([89876cf](https://github.com/belingud/gptcomet/commit/89876cff7fd1a1c6578582736982debed967b7ba)) - belingud
- add ProviderConfig data class and value error handling - ([47665d6](https://github.com/belingud/gptcomet/commit/47665d61d406139f1a8f1cd758e85d0af5b17e30)) - belingud
- add URL validation for required fields - ([fdc9789](https://github.com/belingud/gptcomet/commit/fdc97898dab28c3d2cf4f12df3a7f6c804f0935a)) - belingud
- add rich commit message template and prompt - ([e277bd2](https://github.com/belingud/gptcomet/commit/e277bd2589b7d866a07190e952868495d1461dda)) - belingud
- add RequiredValidator and update logging formatter - ([756a468](https://github.com/belingud/gptcomet/commit/756a468cd60fb8c676b3e3cabb1d7f491cf2038a)) - belingud

### 🐛 Bug Fixes

- ensure lang is checked for None before using - ([7f91dd3](https://github.com/belingud/gptcomet/commit/7f91dd311e37087ef4ba35c9d7ac173a39bdd067)) - belingud
- simplify max_tokens parameter setting - ([40ec6f1](https://github.com/belingud/gptcomet/commit/40ec6f17ff7ded27cdd856ab6c036a83b940f3b4)) - belingud
- change default config path to yaml - ([59c1df4](https://github.com/belingud/gptcomet/commit/59c1df427bfba83b97686ba42c1ac62ceed961da)) - belingud

### 🚜 Refactor

- add xai api key masking support - ([3abc9d0](https://github.com/belingud/gptcomet/commit/3abc9d06ab960dabb6f7685f7280ec07e33bafd3)) - belingud
- update reset command to conditionally reset prompt - ([b2c7732](https://github.com/belingud/gptcomet/commit/b2c7732b6eb5cfe156087dea891b3ad646acb1dc)) - belingud
- update config_manager and utils for better reset functionality and type safety - ([cf4553d](https://github.com/belingud/gptcomet/commit/cf4553d99ac0af845289a72e5cac7b3dd857495a)) - belingud
- add repo_path parameter to MessageGenerator - ([061071b](https://github.com/belingud/gptcomet/commit/061071b54eddd9688521e63fb0f0cbafc129c041)) - belingud
- remove unused imports and simplify utils - ([a32155c](https://github.com/belingud/gptcomet/commit/a32155c7a24a678f3d40598646b821a51ea126f5)) - belingud
- update formatting and add validators in provider.py - ([f23c2e1](https://github.com/belingud/gptcomet/commit/f23c2e1e4c4ce8734a320cf32bfa073ab32178e6)) - belingud
- update code formatting commands - ([c387360](https://github.com/belingud/gptcomet/commit/c387360234f4fc2e6260e997e9b0758decaec67f)) - belingud

### 📚 Documentation

- add help text for cli app - ([5213e30](https://github.com/belingud/gptcomet/commit/5213e30dac48826fdaea81f128fdd3400a0bf9d6)) - belingud
- update command documentation and configuration options - ([4a0e902](https://github.com/belingud/gptcomet/commit/4a0e9023d141681cde9c3846923120397e76e3de)) - belingud
- Update CHANGELOG.md with version 0.0.17 details - ([c1a51e2](https://github.com/belingud/gptcomet/commit/c1a51e22f9788db8aa46e5ea090cfe5c07f3b5cf)) - belingud

### 🧪 Testing

- update message generator test implementation - ([8833c51](https://github.com/belingud/gptcomet/commit/8833c51a5eddf83f47a174118616f2c18d241799)) - belingud
- update commit message style in tests - ([264f231](https://github.com/belingud/gptcomet/commit/264f231db9167f6ccd01953b4c2fd080a058dc2c)) - belingud
- add test cases and new message generator module - ([c659eb5](https://github.com/belingud/gptcomet/commit/c659eb58d95a692d4f3c0c364729fbea49497273)) - belingud

### ⚙️ Miscellaneous Tasks

- update Justfile changelog generation command - ([83f411e](https://github.com/belingud/gptcomet/commit/83f411e0aa5b0339d2a3091d5889ea62628c4ed9)) - belingud
- update changelog generation script - ([2964477](https://github.com/belingud/gptcomet/commit/2964477b11d6da597894a725339812c418b2b64d)) - belingud


---
## [0.0.17](https://github.com/belingud/gptcomet/compare/v0.0.16..v0.0.17) - 2024-11-10

### ⛰️  Features

- support generating rich commit messages - ([22ca79e](https://github.com/belingud/gptcomet/commit/22ca79e8c7fcf454dc3a1215abc9b07217d4736a)) - belingud
- support generating rich commit messages - ([8172c0e](https://github.com/belingud/gptcomet/commit/8172c0ef4ad43020a087c0f459b8a00fc89faf53)) - belingud

### 🐛 Bug Fixes

- update git show format in commit gen - ([bd22e2a](https://github.com/belingud/gptcomet/commit/bd22e2a03232cfd2c0b32a21477983719aa32fde)) - belingud

### 📚 Documentation

- Update CHANGELOG.md with version 0.0.16 details - ([cab0e59](https://github.com/belingud/gptcomet/commit/cab0e59b35d62d6194799e3e721d0109c9d9548c)) - belingud


---
## [0.0.16](https://github.com/belingud/gptcomet/compare/v0.0.14..v0.0.16) - 2024-11-03

### 🐛 Bug Fixes

- Handle KeyboardInterrupt in commit CLI - ([12062f2](https://github.com/belingud/gptcomet/commit/12062f2a3ff6fa5efb021c5c019f98546fd44a9c)) - belingud
- Strip quotes from API key in config_manager - ([4664823](https://github.com/belingud/gptcomet/commit/4664823899247a9ca833ce906672a9807622ad64)) - belingud

### 🚜 Refactor

- Improve staged diff handling - ([5d556dc](https://github.com/belingud/gptcomet/commit/5d556dc3839edb0859e6d9aba2be784eafcce99f)) - belingud
- Simplify CLI version flag and remove unused signal handler - ([28f16b7](https://github.com/belingud/gptcomet/commit/28f16b74c5c9f52d816b0954eb8eadd3a8293da0)) - belingud

### 📚 Documentation

- Clarify git diff explanation in gptcomet.yaml - ([48466a2](https://github.com/belingud/gptcomet/commit/48466a2198dd5c6c709d678d8e077161b75a626a)) - belingud
- Clarify context in git diff example - ([2e24d83](https://github.com/belingud/gptcomet/commit/2e24d832b1351f31c73086a3b10bdc19947d174b)) - belingud
- Update CHANGELOG.md for version 0.0.14 - ([48e67d2](https://github.com/belingud/gptcomet/commit/48e67d28f0ea564fa39fa1509906dad8f6c67115)) - belingud

### ⚡ Performance

- Add diff option for better performance - ([93bb8db](https://github.com/belingud/gptcomet/commit/93bb8dbdfd650688ed8acf809584e2377d1a2839)) - belingud


---
## [0.0.14](https://github.com/belingud/gptcomet/compare/v0.0.13..v0.0.14) - 2024-11-03

### ⛰️  Features

- Add version command to CLI - ([3a10f73](https://github.com/belingud/gptcomet/commit/3a10f737e50dd4d7ddac8414d7ee6853161ad2b1)) - belingud

### 📚 Documentation

- Update CHANGELOG.md for version 0.0.13 - ([ee9f82a](https://github.com/belingud/gptcomet/commit/ee9f82a4bf13c752466491c9d64b2393d106177b)) - belingud


---
## [0.0.13](https://github.com/belingud/gptcomet/compare/v0.0.12..v0.0.13) - 2024-11-03

### 🐛 Bug Fixes

- Fix console output formatting in remove.py - ([efae39b](https://github.com/belingud/gptcomet/commit/efae39b485122de797043e8c32ffe0d582d12429)) - belingud
- Filter out index and metadata lines in diff output - ([7222b2b](https://github.com/belingud/gptcomet/commit/7222b2b357bd0d7879486c3d83e8d6eb267026cc)) - belingud
- Mask API keys in config dump - ([ba8e6fe](https://github.com/belingud/gptcomet/commit/ba8e6fe5f05efe177bf2629c47f50341d0ca7e94)) - belingud

### 📚 Documentation

- update CHANGELOG.md for version 0.0.12 - ([276dde0](https://github.com/belingud/gptcomet/commit/276dde0d61c25be5c85a929f4c50d468656f7fa0)) - belingud

### ⚙️ Miscellaneous Tasks

- Update Justfile default goal and help command - ([91fbd75](https://github.com/belingud/gptcomet/commit/91fbd75638c722a8912b254b4cc46b27560e1408)) - belingud


---
## [0.0.12](https://github.com/belingud/gptcomet/compare/v0.0.11..v0.0.12) - 2024-10-02

### ⛰️  Features

- Enhance command-line application interface and provider management - ([09b1e90](https://github.com/belingud/gptcomet/commit/09b1e90535ac2de90184d8a947dc871b65485049)) - belingud
- enhance text editing and input handling capabilities - ([2517224](https://github.com/belingud/gptcomet/commit/25172240d5af1d9fae4c226e2708209fae4c7b87)) - belingud
- add Provider type and new color variants - ([b45eb7b](https://github.com/belingud/gptcomet/commit/b45eb7bab09c241356441d1ece891e1f4825567d)) - belingud

### 🚜 Refactor

- exclude prompt config in config list command - ([3396533](https://github.com/belingud/gptcomet/commit/33965334155944b564250756659408328e182adf)) - belingud
- improve ask_for_retry function in commit.py - ([9fc7d04](https://github.com/belingud/gptcomet/commit/9fc7d044f93938886181d2b48bd2efe4c73565d7)) - belingud
- Enhance configuration management and provider handling - ([2f12cd8](https://github.com/belingud/gptcomet/commit/2f12cd832e4a5b31ef9e7b2fbc2ebd5168d8b3eb)) - belingud
- Refactor CLI application and enhance user interface - ([6533e3b](https://github.com/belingud/gptcomet/commit/6533e3ba3b78d8dee0a6b27eea2e016b36dc85ed)) - belingud
- Enhance LLMClient functionality and parameter handling - ([879c410](https://github.com/belingud/gptcomet/commit/879c410c2508e057714f5ce7dad1e14989f5562a)) - belingud
- Standardize default parameters for completions API - ([cab02d5](https://github.com/belingud/gptcomet/commit/cab02d5b475635b50fe8a60f063f55dd3c3fe29f)) - belingud

### 📚 Documentation

- update commit message guidelines in gptcomet.yaml - ([7367204](https://github.com/belingud/gptcomet/commit/73672047c83c5bcc59511528a41eff7620a77e5a)) - belingud

### 🧪 Testing

- Improve GPTComet branding and configuration management - ([b623690](https://github.com/belingud/gptcomet/commit/b6236905fa8d97c601ddf622c303a17ab821a2ce)) - belingud

### ⚙️ Miscellaneous Tasks

- update dependencies in pyproject.toml - ([d6b2b87](https://github.com/belingud/gptcomet/commit/d6b2b872774162ebdb65a1598dc2b6906a80ad84)) - belingud
- Update gptcomet configuration and commit guidelines - ([b0f128a](https://github.com/belingud/gptcomet/commit/b0f128afe8cfec92cb9b23a51d293d7af50aef39)) - belingud

# Changelog


---
## [0.0.10](https://github.com/belingud/gptcomet/compare/v0.0.9..v0.0.10) - 2024-09-14

### ⛰️  Features

- refine LLMClient config and enhance API handling - ([b8baf21](https://github.com/belingud/gptcomet/commit/b8baf211f8bfbce0da082d25e926be8047cfae56)) - belingud
- Add append and remove commands to config CLI - ([ea4431d](https://github.com/belingud/gptcomet/commit/ea4431dfeebf85aac0bfe8354c825a503721369f)) - belingud
- Rename gen to commit and refactor commit CLI - ([ccd059a](https://github.com/belingud/gptcomet/commit/ccd059aa6b363ff124189b2aa4aacc0a94bb7ebe)) - belingud

### 📚 Documentation

- Update README with static badges and TOC - ([1be3d34](https://github.com/belingud/gptcomet/commit/1be3d341b9757586abf229f807a250ba6f63fce9)) - belingud
- refactor README for CodeGPT Documentation and Enhancements - ([2f46e18](https://github.com/belingud/gptcomet/commit/2f46e18c535c454db452b5c8a30a306f1e63fbe4)) - belingud
- archive project and point to CodeGPT alternative - ([60b77e2](https://github.com/belingud/gptcomet/commit/60b77e24a6bb84d34ae358fd19801157d21feab5)) - belingud
- Update changelog - ([21bb74c](https://github.com/belingud/gptcomet/commit/21bb74cd27a1038b8e4d52fc267d6ef996aae361)) - belingud

### 🧪 Testing

- enhance and Refactor Test Suite - ([0b9e1ae](https://github.com/belingud/gptcomet/commit/0b9e1aee3529b2827c74230d6a7382b878040395)) - belingud

### ⚙️ Miscellaneous Tasks

- enhance project configuration and dependencies management - ([9c95295](https://github.com/belingud/gptcomet/commit/9c9529518a8040fb02b307ad97a3161bcf7f9527)) - belingud
- Update dependencies and remove litellm - ([78e090d](https://github.com/belingud/gptcomet/commit/78e090daf7f9ec3fe833e71800dd0ed03c497364)) - belingud


---
## [0.0.9](https://github.com/belingud/gptcomet/compare/v0.0.8..v0.0.9) - 2024-09-08

### 🚜 Refactor

- Simplify config manager and log module - ([cf15285](https://github.com/belingud/gptcomet/commit/cf15285038f3524ab57fbb4f1449fcb360eec30c)) - belingud


---
## [0.0.8](https://github.com/belingud/gptcomet/compare/v0.0.7..v0.0.8) - 2024-09-08

### ⛰️  Features

- Add "edit" option to commit message generation - ([5636646](https://github.com/belingud/gptcomet/commit/56366467c6e2f02e1978ed51e4648831ea7f6e41)) - belingud

### 🚜 Refactor

- Simplify commit output and use template - ([9c2663c](https://github.com/belingud/gptcomet/commit/9c2663cb520db6ffc5405b82f1e3f7695e85e010)) - belingud
- Correct import and skip isort directive - ([cbbef2d](https://github.com/belingud/gptcomet/commit/cbbef2d53f401cb23479208e733f16ad20803221)) - belingud

### 📚 Documentation

- Update CLI command descriptions and add 'keys' command - ([8bfff4f](https://github.com/belingud/gptcomet/commit/8bfff4f2f3db7e405bbdc73b3e8ce2304336ed3c)) - belingud
- Update Changelog for v0.0.7 release - ([55bf557](https://github.com/belingud/gptcomet/commit/55bf557eaed82e2985797f9445a70d200e7ce005)) - belingud



---
## [0.0.7](https://github.com/belingud/gptcomet/compare/v0.0.6..v0.0.7) - 2024-09-05

### 🚜 Refactor

- Refactor config management CLI commands - ([e2261f9](https://github.com/belingud/gptcomet/commit/e2261f961555d6e0b204291ca59671a94f08c1fe)) - belingud

### 📚 Documentation

- Update Changelog for v0.0.6 release - ([f1b7578](https://github.com/belingud/gptcomet/commit/f1b75789e84d5553a5947a216874c4ccb3b8fe4a)) - belingud

### 🧪 Testing

- Add smoke test for gmsg commands - ([523fe75](https://github.com/belingud/gptcomet/commit/523fe75e2157f04bc2b32d2edaf443cfcb3c6ba8)) - belingud

---
## [0.0.6](https://github.com/belingud/gptcomet/compare/v0.0.5..v0.0.6) - 2024-08-29

### 🚜 Refactor

- Refactor commit message generation logic - ([c91b3ab](https://github.com/belingud/gptcomet/commit/c91b3ab49dddb4caba180116b6dac7d8b8ef916d)) - belingud

### 📚 Documentation

- Update README and CHANGELOG for project renaming - ([19003e5](https://github.com/belingud/gptcomet/commit/19003e5201c17308f649caa9b811b1df5df8c0f8)) - belingud

### ⚙️ Miscellaneous Tasks

- Update changelog script for new tag handling - ([a943359](https://github.com/belingud/gptcomet/commit/a943359b545f706f7516bd8f5af096dff68b3af4)) - belingud
- Update rich dependency to 13.8.0 - ([0c391d7](https://github.com/belingud/gptcomet/commit/0c391d765511fced2c5826974ebe9effcc86261f)) - belingud

---
## [0.0.5](https://github.com/belingud/gptcomet/compare/v0.0.3..v0.0.5) - 2024-08-28

### ⛰️  Features

- Add path command to config management - ([5a4a7b8](https://github.com/belingud/gptcomet/commit/5a4a7b8abb2bc29645a985c76ba7777eebeb9726)) - belingud
- Add new CLI commands for managing configuration and generating commit messages - ([c9a8e5f](https://github.com/belingud/gptcomet/commit/c9a8e5fa3b65c23ffe7a0ff89b0ab5446a079363)) - belingud
- use yaml as config file, improve output language support - ([399f584](https://github.com/belingud/gptcomet/commit/399f584b773bf71fe04951821f5d4f1c425e7b61)) - belingud
- Add tests for message generator and utils functions - ([ab6f6f7](https://github.com/belingud/gptcomet/commit/ab6f6f73edaa8f650827f23bed5776f0d32c8cc6)) - belingud
- Update changelog and cliff.toml configuration - ([6feabe7](https://github.com/belingud/gptcomet/commit/6feabe7b25c53c6d4acba4d812c57ddf57dca970)) - belingud
- Update changelog generation script and profile tests - ([40ef425](https://github.com/belingud/gptcomet/commit/40ef4254053896c0d50e8b7d210e3872d5221b14)) - belingud
- Update ConfigManager to use YAML instead of TOML - ([d3d49d5](https://github.com/belingud/gptcomet/commit/d3d49d5a7682d6fbebaa3ddf3557d600f75ccec2)) - belingud

### 📚 Documentation

- Update README.md with new configuration options and usage examples - ([ced8861](https://github.com/belingud/gptcomet/commit/ced886146cad2203bc97c8b669fc08111b8d6625)) - belingud

### 🧪 Testing

- Update test import to use new stylize module - ([7fb22bd](https://github.com/belingud/gptcomet/commit/7fb22bd6e7947d0aa4c2872df7f9ffc055a94ab7)) - belingud
- Add rich text styling tests - ([198cfdf](https://github.com/belingud/gptcomet/commit/198cfdffd2c3f92f31a1d0fc9a9fc10efb3f2181)) - belingud

### ⚙️ Miscellaneous Tasks

- Update .gitignore file - ([857ff3f](https://github.com/belingud/gptcomet/commit/857ff3f8e8f83913c0489033b3cc8c857710497a)) - belingud
- Update gptcomet.yaml configuration - ([8860290](https://github.com/belingud/gptcomet/commit/886029022562ecb23dbc7cc8c3dd44630266546b)) - belingud
- Update dependencies and scripts in pyproject.toml - ([d2a9251](https://github.com/belingud/gptcomet/commit/d2a92516d21995760b0f8266f9e46aa4271715f8)) - belingud

---
## [0.0.3] - 2024-08-21

### ⛰️  Features

- Add line profiler and performance metrics for config manager functions - ([8fb90dc](https://github.com/belingud/gptcomet/commit/8fb90dc5c974b6096ae95366dc359beea94d6687)) - belingud
- Add config append and remove functionality - ([bf229db](https://github.com/belingud/gptcomet/commit/bf229db4b81b600c274f22c8edb864cdbb7b50a7)) - belingud
- Add file ignore configuration and validation - ([ee47d54](https://github.com/belingud/gptcomet/commit/ee47d54a474d2ff914cf70c2e9759288da7f70a2)) - belingud
- Add AICommit documentation and usage guide - ([6ecd0fa](https://github.com/belingud/gptcomet/commit/6ecd0fa4adeef4ca681e7666e01bd828b3884ca6)) - belingud
- Add initial project files and setup - ([7f64221](https://github.com/belingud/gptcomet/commit/7f642212efdd115eab490b0c18625ad975f88c3e)) - belingud
- Implement config manager and test cases - ([bf20745](https://github.com/belingud/gptcomet/commit/bf20745ee23a9cf8fc84aa030e6d8f8f5d8bf744)) - belingud
- Add support key generation script - ([72c0d77](https://github.com/belingud/gptcomet/commit/72c0d778af2c817ab6c24f1fae18057e81c27ccf)) - belingud
- Add documentation for aicommit library - ([5ef67ce](https://github.com/belingud/gptcomet/commit/5ef67ce28fe8e92489be0a8852cdaf4ab0bef07a)) - belingud
- Add AI-powered commit message generation - ([2952f0a](https://github.com/belingud/gptcomet/commit/2952f0a3661055a4937e003a5a3c00935fb5b232)) - belingud

### 🐛 Bug Fixes

- Update GPTComet CLI and LLM client for better error handling - ([35713a8](https://github.com/belingud/gptcomet/commit/35713a8a05a3f7a14a562989b18191a252c6f0ad)) - belingud
- Use repo.index.commit for committing changes - ([b9e9489](https://github.com/belingud/gptcomet/commit/b9e948927cdf7168b746b1e2031b5cc003c6b9a8)) - belingud
- implement aicommit CLI with OpenAI and GitPython integration - ([b93f37d](https://github.com/belingud/gptcomet/commit/b93f37db1b289a3fb171957e18d490f5b58a2d42)) - belingud (aider)

### 🚜 Refactor

- Init ConfigManager with config path and update get\_config\_path method - ([6fe9f2a](https://github.com/belingud/gptcomet/commit/6fe9f2a6207159c825e8f8a3f3b4efe5589c83ec)) - belingud

### 📚 Documentation

- Update README.md with contribution instructions - ([a9523b9](https://github.com/belingud/gptcomet/commit/a9523b9c20e4d9ac4a0b185431486245028e9b40)) - belingud

### Add

- Add optional socks dependency and pyinstrument profile script. - ([1f45bae](https://github.com/belingud/gptcomet/commit/1f45baef2d1a7b765e8a6506758acdfded8d25a1)) - belingud

### Init

- Add bumpversion configuration for versioning - ([6cd230d](https://github.com/belingud/gptcomet/commit/6cd230dc4756eb6ad1e0cb4e4aeac911dcbe8099)) - belingud

### Update

- Rename `aicommit` to `gptcomet` and update related files - ([62a8f78](https://github.com/belingud/gptcomet/commit/62a8f78b456c125b239aa1d7ec0ebed054669410)) - belingud


All notable changes to this project will be documented in this file. See [conventional commits](https://www.conventionalcommits.org/) for commit guidelines.

---
## [0.0.3] - 2024-08-21

### ⛰️  Features

- Add line profiler and performance metrics for config manager functions - ([8fb90dc](https://github.com/belingud/gptcomet/commit/8fb90dc5c974b6096ae95366dc359beea94d6687)) - belingud
- Add config append and remove functionality - ([bf229db](https://github.com/belingud/gptcomet/commit/bf229db4b81b600c274f22c8edb864cdbb7b50a7)) - belingud
- Add file ignore configuration and validation - ([ee47d54](https://github.com/belingud/gptcomet/commit/ee47d54a474d2ff914cf70c2e9759288da7f70a2)) - belingud
- Add AICommit documentation and usage guide - ([6ecd0fa](https://github.com/belingud/gptcomet/commit/6ecd0fa4adeef4ca681e7666e01bd828b3884ca6)) - belingud
- Add initial project files and setup - ([7f64221](https://github.com/belingud/gptcomet/commit/7f642212efdd115eab490b0c18625ad975f88c3e)) - belingud
- Implement config manager and test cases - ([bf20745](https://github.com/belingud/gptcomet/commit/bf20745ee23a9cf8fc84aa030e6d8f8f5d8bf744)) - belingud
- Add support key generation script - ([72c0d77](https://github.com/belingud/gptcomet/commit/72c0d778af2c817ab6c24f1fae18057e81c27ccf)) - belingud
- Add documentation for aicommit library - ([5ef67ce](https://github.com/belingud/gptcomet/commit/5ef67ce28fe8e92489be0a8852cdaf4ab0bef07a)) - belingud
- Add AI-powered commit message generation - ([2952f0a](https://github.com/belingud/gptcomet/commit/2952f0a3661055a4937e003a5a3c00935fb5b232)) - belingud

### 🐛 Bug Fixes

- Update GPTComet CLI and LLM client for better error handling - ([35713a8](https://github.com/belingud/gptcomet/commit/35713a8a05a3f7a14a562989b18191a252c6f0ad)) - belingud
- Use repo.index.commit for committing changes - ([b9e9489](https://github.com/belingud/gptcomet/commit/b9e948927cdf7168b746b1e2031b5cc003c6b9a8)) - belingud
- implement aicommit CLI with OpenAI and GitPython integration - ([b93f37d](https://github.com/belingud/gptcomet/commit/b93f37db1b289a3fb171957e18d490f5b58a2d42)) - belingud (aider)

### 🚜 Refactor

- Init ConfigManager with config path and update get\_config\_path method - ([6fe9f2a](https://github.com/belingud/gptcomet/commit/6fe9f2a6207159c825e8f8a3f3b4efe5589c83ec)) - belingud

### 📚 Documentation

- Update README.md with contribution instructions - ([a9523b9](https://github.com/belingud/gptcomet/commit/a9523b9c20e4d9ac4a0b185431486245028e9b40)) - belingud

### Add

- Add optional socks dependency and pyinstrument profile script. - ([1f45bae](https://github.com/belingud/gptcomet/commit/1f45baef2d1a7b765e8a6506758acdfded8d25a1)) - belingud

### Init

- Add bumpversion configuration for versioning - ([6cd230d](https://github.com/belingud/gptcomet/commit/6cd230dc4756eb6ad1e0cb4e4aeac911dcbe8099)) - belingud

### Update

- Rename `aicommit` to `gptcomet` and update related files - ([62a8f78](https://github.com/belingud/gptcomet/commit/62a8f78b456c125b239aa1d7ec0ebed054669410)) - belingud
