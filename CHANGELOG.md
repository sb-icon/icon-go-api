# Changelog

## [0.5.7](https://github.com/sb-icon/icon-go-api/compare/v0.5.6...v0.5.7) (2024-11-18)


### Bug Fixes

* address doc string ([ecdaa8d](https://github.com/sb-icon/icon-go-api/commit/ecdaa8d1e636767b4c319bb5f571a3645800a207))

## [0.5.6](https://github.com/sb-icon/icon-go-api/compare/v0.5.5...v0.5.6) (2024-10-15)


### Bug Fixes

* transactions and address doc string ([bbae026](https://github.com/sb-icon/icon-go-api/commit/bbae0269689d3fdc42758d3a86b7a3480855c168))

## [0.5.5](https://github.com/sb-icon/icon-go-api/compare/v0.5.4...v0.5.5) (2024-10-15)


### Bug Fixes

* update doc string for stats ([dfb7442](https://github.com/sb-icon/icon-go-api/commit/dfb74428f10adb28f2453a14ec00ba8c8a6a88cb))
* update doc string for supplies ([42709a1](https://github.com/sb-icon/icon-go-api/commit/42709a18ad90aaa34b3f87ca5521aadd7401d5c6))
* version and metadata endpoint doc strings ([c87de9d](https://github.com/sb-icon/icon-go-api/commit/c87de9dac3264b68e61b08e48c30d77d6138cdca))

## [0.5.4](https://github.com/sb-icon/icon-go-api/compare/v0.5.3...v0.5.4) (2024-06-27)


### Bug Fixes

* make circ supply update every 5 min ([8ce81c9](https://github.com/sb-icon/icon-go-api/commit/8ce81c9fbc72fd23809e183fbf62a22ba0faa554))

## [0.5.3](https://github.com/sb-icon/icon-go-api/compare/v0.5.2...v0.5.3) (2024-03-26)


### Bug Fixes

* transaction log on bad query ([56f64ea](https://github.com/sb-icon/icon-go-api/commit/56f64eab788e489905b7b146bcaeda7c8c0b7a6c))

## [0.5.2](https://github.com/sb-icon/icon-go-api/compare/v0.5.1...v0.5.2) (2024-03-26)


### Bug Fixes

* improve the error logging statements ([7753db4](https://github.com/sb-icon/icon-go-api/commit/7753db48272f98861b7b71aea2928a85695f7ca1))

## [0.5.1](https://github.com/sb-icon/icon-go-api/compare/v0.5.0...v0.5.1) (2024-03-26)


### Bug Fixes

* flip strict mode on query paths ([9e249b9](https://github.com/sb-icon/icon-go-api/commit/9e249b931308597aac2da426afc866231c5fd5a1))
* handle api error gracefully ([21b3ba1](https://github.com/sb-icon/icon-go-api/commit/21b3ba1b40d648d4b2be2e1cfe912479c0bd2f1b))
* handle coingecko errors gracefully - not getting through in docker but are locally and in cluster ([a25f653](https://github.com/sb-icon/icon-go-api/commit/a25f653a328ff43c4f447c46ef534e6a7f1bd216))

## [0.5.0](https://github.com/sb-icon/icon-go-api/compare/v0.4.5...v0.5.0) (2023-11-01)


### Features

* add csv option to endpoints with header [#39](https://github.com/sb-icon/icon-go-api/issues/39) ([3d57ea7](https://github.com/sb-icon/icon-go-api/commit/3d57ea71b49f310bbbb3560961adf4e650df539f))


### Bug Fixes

* add supplies endpoint vs stats ([b4025d5](https://github.com/sb-icon/icon-go-api/commit/b4025d59b98899819234c0b117608cac3a0420ff))

## [0.4.5](https://github.com/sb-icon/icon-go-api/compare/v0.4.4...v0.4.5) (2023-04-22)


### Bug Fixes

* bump the MaxPageSkip to 1.5M ([cfe77bf](https://github.com/sb-icon/icon-go-api/commit/cfe77bfce7de2598c5ba6a2c6da860b1ca29a525))

## [0.4.4](https://github.com/sb-icon/icon-go-api/compare/v0.4.3...v0.4.4) (2023-03-14)


### Bug Fixes

* add total-supply endpoint ([90764f5](https://github.com/sb-icon/icon-go-api/commit/90764f5e6bc1ba0ed684a8762792ad9bd26c32db))

## [0.4.3](https://github.com/sb-icon/icon-go-api/compare/v0.4.2...v0.4.3) (2023-03-01)


### Bug Fixes

* add transaction_index to token_transfers to allow sorting [#71](https://github.com/sb-icon/icon-go-api/issues/71) ([a25b10a](https://github.com/sb-icon/icon-go-api/commit/a25b10a7e583aa1fb98e30988ba9ab551747a39c))

## [0.4.2](https://github.com/sb-icon/icon-go-api/compare/v0.4.1...v0.4.2) (2023-02-14)


### Bug Fixes

* build ([1d0c464](https://github.com/sb-icon/icon-go-api/commit/1d0c4641b0aa72466f88dcebb21b74c5543d6713))

## [0.4.1](https://github.com/sb-icon/icon-go-api/compare/v0.4.0...v0.4.1) (2022-12-20)


### Bug Fixes

* rm the MaxPageSkip on addresses ([2e6d4c2](https://github.com/sb-icon/icon-go-api/commit/2e6d4c294ecaa08caa9612c1ffef23f8f7beed70))

## [0.4.0](https://github.com/sb-icon/icon-go-api/compare/v0.3.0...v0.4.0) (2022-12-13)


### Features

* add sort to addresses/contracts [#18](https://github.com/sb-icon/icon-go-api/issues/18) ([4875b36](https://github.com/sb-icon/icon-go-api/commit/4875b366e8d782117384cae68cc92f5a399340cd))

## [0.3.0](https://github.com/sb-icon/icon-go-api/compare/v0.2.0...v0.3.0) (2022-12-08)


### Features

* add mkt cap stats endpoint ([b3bf738](https://github.com/sb-icon/icon-go-api/commit/b3bf73882bb3271d2b26f665e68a67edf7d20e23))
* add stats endpoint with mkt cap ([6f2d97a](https://github.com/sb-icon/icon-go-api/commit/6f2d97a43e48d2363d6d214ad5f53b922cc53638))


### Bug Fixes

* omit  in sort string on /addresses and switch meaning of ~/code/sb-icon/icon-explorer/api/src/api [#20](https://github.com/sb-icon/icon-go-api/issues/20) ([54e0e26](https://github.com/sb-icon/icon-go-api/commit/54e0e26f15a19a4b78f5d932bd2fd15e9d98a1b8))

## [0.2.0](https://github.com/sb-icon/icon-go-api/compare/v0.1.6...v0.2.0) (2022-11-24)


### Features

* add new properties to addresses and contracts list and details ([1baf8ae](https://github.com/sb-icon/icon-go-api/commit/1baf8ae15d71c9ea24dbb1d77c069ea1b54dd8d9))


### Bug Fixes

* add limit and skip to icx txs endpoint docs ([a9970f2](https://github.com/sb-icon/icon-go-api/commit/a9970f27cf40fd6564b51809d7e2f3c90e9d7719))
