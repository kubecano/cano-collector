# Changelog

## [0.0.13](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.12...cano-collector-v0.0.13) (2025-07-09)


### Features

* CU-8699hqm33 - Basic Slack Message Formatting ([#146](https://github.com/kubecano/cano-collector/issues/146)) ([d9f4ace](https://github.com/kubecano/cano-collector/commit/d9f4ace525a1689921557e13b361d0bf5eef0841))
* CU-8699hqn2f - Basic Enrichment Framework ([#147](https://github.com/kubecano/cano-collector/issues/147)) ([efe51ca](https://github.com/kubecano/cano-collector/commit/efe51ca2cded5c53c4014e92f4de3385485259c9))
* CU-8699hqnyn - Label and Annotation Enrichment ([#149](https://github.com/kubecano/cano-collector/issues/149)) ([1985590](https://github.com/kubecano/cano-collector/commit/19855905189debed3815e637b85e1d6de6bf3679))
* CU-8699hqpay - Enhanced Slack Message with Enrichments ([#150](https://github.com/kubecano/cano-collector/issues/150)) ([0f5442f](https://github.com/kubecano/cano-collector/commit/0f5442f7e55f3a90c08e7e38755c74948cc76d52))
* CU-8699hqpay - Enhanced Slack Message with Enrichments ([#151](https://github.com/kubecano/cano-collector/issues/151)) ([1e547af](https://github.com/kubecano/cano-collector/commit/1e547af74d9d0aa6498a6e73db79ef5da7c6efa5))
* CU-8699hqpay - Enhanced Slack Message with Enrichments ([#152](https://github.com/kubecano/cano-collector/issues/152)) ([e3b5d04](https://github.com/kubecano/cano-collector/commit/e3b5d042f8a9cb49b2dc251eec11f213d686520d))

## [0.0.12](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.11...cano-collector-v0.0.12) (2025-07-07)


### Features

* CU-8699hqkbv - Alert Data Structure and Validation ([#137](https://github.com/kubecano/cano-collector/issues/137)) ([d77aea1](https://github.com/kubecano/cano-collector/commit/d77aea109a4e4b332967c71b11a5777b8a88ac4a))
* CU-8699hqkp6 - Alert to Issue Conversion ([#140](https://github.com/kubecano/cano-collector/issues/140)) ([4686e47](https://github.com/kubecano/cano-collector/commit/4686e479cb579388f81ae538c23fbff3d50bd068))
* CU-8699hu9hk - Health Checks and Basic Metrics ([#139](https://github.com/kubecano/cano-collector/issues/139)) ([54eb50f](https://github.com/kubecano/cano-collector/commit/54eb50f870b0779d1edbe4e234ba9d46209e038c))


### Bug Fixes

* **deps:** update module github.com/slack-go/slack to v0.17.3 ([#142](https://github.com/kubecano/cano-collector/issues/142)) ([cb633c6](https://github.com/kubecano/cano-collector/commit/cb633c640010dbab9f3e41b287612d57f6a47c0c))

## [0.0.11](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.10...cano-collector-v0.0.11) (2025-07-01)


### Bug Fixes

* CU-8699hgr18 - Raw alert to Slack Delivery - fix configuration ([#135](https://github.com/kubecano/cano-collector/issues/135)) ([bbf1746](https://github.com/kubecano/cano-collector/commit/bbf1746b3e9c8b283b91efdf9b41fc25e1c2ab61))

## [0.0.10](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.9...cano-collector-v0.0.10) (2025-07-01)


### Features

* CU-8699gn6fk - Add README and documentation for cano-collector ([39bfdf9](https://github.com/kubecano/cano-collector/commit/39bfdf9ad9cf5e546d6fc0f162659b641df670f8))
* CU-8699gn6fk - Add README and documentation for cano-collector ([#120](https://github.com/kubecano/cano-collector/issues/120)) ([8cc8c02](https://github.com/kubecano/cano-collector/commit/8cc8c0295110618452960d43bcf9f14c1c96b674))
* CU-8699gpzq5 - Render cano-collector docs and deploy to S3 ([#121](https://github.com/kubecano/cano-collector/issues/121)) ([ba56c67](https://github.com/kubecano/cano-collector/commit/ba56c67471ed9772cef99afa5580018a0678cef6))
* CU-8699hgr18 - Raw alert to Slack Delivery ([#124](https://github.com/kubecano/cano-collector/issues/124)) ([c7abbd7](https://github.com/kubecano/cano-collector/commit/c7abbd73a9beae93d80c52cc41e4562c3e0de338))
* CU-8699hgr18 - Raw alert to Slack Delivery ([#125](https://github.com/kubecano/cano-collector/issues/125)) ([45bd3b5](https://github.com/kubecano/cano-collector/commit/45bd3b58ff1f3162637cfb86b29c214b251612c7))
* CU-8699hgr18 - Raw alert to Slack Delivery ([#126](https://github.com/kubecano/cano-collector/issues/126)) ([0e2f440](https://github.com/kubecano/cano-collector/commit/0e2f440b77d93962a250396f1c0b89a15dc3db86))


### Bug Fixes

* **deps:** update module github.com/getsentry/sentry-go to v0.34.0 ([#127](https://github.com/kubecano/cano-collector/issues/127)) ([a27348d](https://github.com/kubecano/cano-collector/commit/a27348dbb88ea800008203ea5622f9bdb21b0971))
* **deps:** update module github.com/getsentry/sentry-go/gin to v0.34.0 ([#128](https://github.com/kubecano/cano-collector/issues/128)) ([fce87c2](https://github.com/kubecano/cano-collector/commit/fce87c21e15062484930131364b5534e34943940))
* **deps:** update module go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin to v0.62.0 ([#130](https://github.com/kubecano/cano-collector/issues/130)) ([e2680cc](https://github.com/kubecano/cano-collector/commit/e2680cc34c04f6913d234b0388da8fcd9b4544e4))
* **deps:** update opentelemetry-go monorepo to v1.37.0 ([#131](https://github.com/kubecano/cano-collector/issues/131)) ([e0e46da](https://github.com/kubecano/cano-collector/commit/e0e46da16d0207f3321ed183748c884419928b05))

## [0.0.9](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.8...cano-collector-v0.0.9) (2025-06-15)


### Bug Fixes

* CU-869972wjv - Deploy Kubecano on dev cluster ([0c64c04](https://github.com/kubecano/cano-collector/commit/0c64c0461a113c667cf2e5ed8be18f81d0c4a069))

## [0.0.8](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.7...cano-collector-v0.0.8) (2025-06-15)


### Bug Fixes

* Fix Docker image versioning ([17b501d](https://github.com/kubecano/cano-collector/commit/17b501d4483f3dd06ab1867660c2054a48c7dd7b))

## [0.0.7](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.6...cano-collector-v0.0.7) (2025-06-15)


### Bug Fixes

* Fix Docker image versioning ([22e1475](https://github.com/kubecano/cano-collector/commit/22e1475a5e1be139b810b14484f3981c2d7d38b0))

## [0.0.6](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.5...cano-collector-v0.0.6) (2025-06-15)


### Bug Fixes

* Fix Docker image versioning ([d39f8e8](https://github.com/kubecano/cano-collector/commit/d39f8e8541533c5a6a79a49605236461497991af))

## [0.0.5](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.4...cano-collector-v0.0.5) (2025-06-13)


### Bug Fixes

* Fix Chart and image pipeline pushes ([9b779c6](https://github.com/kubecano/cano-collector/commit/9b779c64d40ed80f0e2e362214db108b92205932))

## [0.0.4](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.3...cano-collector-v0.0.4) (2025-06-13)


### Bug Fixes

* Fix Chart and image pipeline pushes ([fceed57](https://github.com/kubecano/cano-collector/commit/fceed57e7bb6fccb859683f9b24c8b4f0b7c1438))

## [0.0.3](https://github.com/kubecano/cano-collector/compare/cano-collector-v0.0.2...cano-collector-v0.0.3) (2025-06-13)


### Features

* CU-86985a80p - Add AlertManager alerts handler ([#60](https://github.com/kubecano/cano-collector/issues/60)) ([1cf86e7](https://github.com/kubecano/cano-collector/commit/1cf86e7b2247145d0b01dc2831efcbf5dc0511f4))
* CU-86985aa3g - Add bundled kube-prom-stack to collector helm chart ([#57](https://github.com/kubecano/cano-collector/issues/57)) ([7ad12ed](https://github.com/kubecano/cano-collector/commit/7ad12edf1fb4e90ad300dd79c48dff0493ba3f48))
* CU-86985axww - Send alert using strategy pattern - factory test ([#80](https://github.com/kubecano/cano-collector/issues/80)) ([d8a2278](https://github.com/kubecano/cano-collector/commit/d8a2278ca924f18f193c0988df787ac27ff543db))
* CU-86985axww - Send alert using strategy pattern ([#79](https://github.com/kubecano/cano-collector/issues/79)) ([230ab24](https://github.com/kubecano/cano-collector/commit/230ab247863cce3a8b4dec1508cf918a4212f9b1))
* CU-86985axxj - Add destinations to Helm Chart ([#64](https://github.com/kubecano/cano-collector/issues/64)) ([4f2dc0e](https://github.com/kubecano/cano-collector/commit/4f2dc0e14542fc1ddea9520fcfe5e8b7218b08f8))
* CU-8698ck4q1 - Dependency Injection refactor ([#65](https://github.com/kubecano/cano-collector/issues/65)) ([3f3146e](https://github.com/kubecano/cano-collector/commit/3f3146e4d48e240f05912e001de7f702d4b514ea))


### Bug Fixes

* **deps:** update module github.com/getsentry/sentry-go to v0.33.0 ([f8f3d3f](https://github.com/kubecano/cano-collector/commit/f8f3d3f29c9b36497f6aef0b77b2a9adc569018b))
* **deps:** update module github.com/getsentry/sentry-go/gin to v0.33.0 ([64cb22b](https://github.com/kubecano/cano-collector/commit/64cb22b9b149a11af9d612bb08af4833e07fd182))
* **deps:** update module github.com/gin-contrib/zap to v1.1.5 ([07efcb2](https://github.com/kubecano/cano-collector/commit/07efcb27604948d22468fd5012f8eb4dd1942ee8))
* **deps:** update module github.com/gin-gonic/gin to v1.10.1 ([d9fc02e](https://github.com/kubecano/cano-collector/commit/d9fc02e669e759e6913c52f1f21622de0a9e1d01))
* **deps:** update module github.com/hellofresh/health-go/v5 to v5.5.4 ([f7c2caf](https://github.com/kubecano/cano-collector/commit/f7c2cafe962ea662b018afa80748179dd0830d53))
* **deps:** update module github.com/prometheus/client_golang to v1.22.0 ([c3f3a81](https://github.com/kubecano/cano-collector/commit/c3f3a81843f5a140941239debea79d7332fc568a))
* **deps:** update module go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin to v0.61.0 ([fab3842](https://github.com/kubecano/cano-collector/commit/fab384258629a8e009812555300f97873184c1f2))
* **deps:** update opentelemetry-go monorepo to v1.36.0 ([cabbf61](https://github.com/kubecano/cano-collector/commit/cabbf6168978cc73617540778fb51beafa1d1e9c))
* NO-ISSUE - Removed compromised action tj-actions/changed-files ([#61](https://github.com/kubecano/cano-collector/issues/61)) ([4894a5c](https://github.com/kubecano/cano-collector/commit/4894a5c23a605a2e56bbdab6bc7d77bde584ae2d))
* rename health mock file ([3e25528](https://github.com/kubecano/cano-collector/commit/3e255282e15bb0f0834023dc2c0bb4b84a0e6291))

## [0.0.2](https://github.com/kubecano/cano-collector/compare/cano-collector-0.0.1...cano-collector-v0.0.2) (2025-03-10)


### Features

* CU-8697896e6 - Add golang static analysis in GitHub actions pip… ([#40](https://github.com/kubecano/cano-collector/issues/40)) ([c66db5f](https://github.com/kubecano/cano-collector/commit/c66db5fdcacd5b95c4c2b4e4e636933884d67d74))
* CU-8697896e6 - Add golang static analysis in GitHub actions pipelines ([#39](https://github.com/kubecano/cano-collector/issues/39)) ([b3c1cf1](https://github.com/kubecano/cano-collector/commit/b3c1cf16734f18caa090034222b250acaaa0b590))
* CU-86978gu16 - Connect cano-collector to Sonar ([#36](https://github.com/kubecano/cano-collector/issues/36)) ([19e9424](https://github.com/kubecano/cano-collector/commit/19e942410efdf6b9b61401b61114f826c6c9565b))
* CU-86978pwb5 - Add collector to Sentry.io ([#48](https://github.com/kubecano/cano-collector/issues/48)) ([a01fe53](https://github.com/kubecano/cano-collector/commit/a01fe53cd1ec640ac226801ff746dad3a283a10e))
* CU-86978pwb5 - Add collector to Sentry.io ([#50](https://github.com/kubecano/cano-collector/issues/50)) ([e31cf50](https://github.com/kubecano/cano-collector/commit/e31cf5071fcbd1ef50e7e9cef3f4a08ce8c5f26a))
* CU-8697906gb - Add metrics endpoint to Collector ([#49](https://github.com/kubecano/cano-collector/issues/49)) ([c106520](https://github.com/kubecano/cano-collector/commit/c1065206638079be4e76301422535c9a648a748f))
* CU-8697ep0yb - Add structured logger to Collector with slog or uber-go/zap - refactoring ([#53](https://github.com/kubecano/cano-collector/issues/53)) ([c6bb241](https://github.com/kubecano/cano-collector/commit/c6bb24193c89185c7db005b357b7cc8888ce1d97))
* CU-8697ep0yb - Add structured logger to Collector with slog or uber-go/zap ([#51](https://github.com/kubecano/cano-collector/issues/51)) ([9fe2696](https://github.com/kubecano/cano-collector/commit/9fe2696c60ac107a720858c39d0fd64ad2a0d5cd))
* CU-8697ep13a - Add tracing to Collector with Jaeger and OpenTelemetry ([#54](https://github.com/kubecano/cano-collector/issues/54)) ([075951c](https://github.com/kubecano/cano-collector/commit/075951ca1020d220cc1e9c7cd6296da649a19208))
* CU-869877xmz - Add healthcheck endpoint ([#52](https://github.com/kubecano/cano-collector/issues/52)) ([6e57b1d](https://github.com/kubecano/cano-collector/commit/6e57b1d0f724c1855e1a63cb02f3365ff7639ada))


### Bug Fixes

* **deps:** update kubernetes packages to v0.32.2 ([10c62ef](https://github.com/kubecano/cano-collector/commit/10c62ef898024a9cf13937807a2c06dfa3e9fed0))
* **deps:** update kubernetes packages to v0.32.2 ([e75d73c](https://github.com/kubecano/cano-collector/commit/e75d73c6f6a7e8f4b0ca3876022f985d205e42a5))
* NO-ISSUE - Fix renovate schedule to cost saving ([a54373c](https://github.com/kubecano/cano-collector/commit/a54373c9366b9248a156aba7d15ee36f442cef9a))

## 0.0.1 (2025-01-02)


### Features

* CU-86977fhvz - Add GitHub action pipeline ([#8](https://github.com/kubecano/cano-collector/issues/8)) ([62ed895](https://github.com/kubecano/cano-collector/commit/62ed89580d5cfc029da2f758329dc7d387c2c098))
* CU-86977fmk7 - Add basic Helm chart ([#4](https://github.com/kubecano/cano-collector/issues/4)) ([e404252](https://github.com/kubecano/cano-collector/commit/e4042528bc330a89397494f29655dfc09ba195cc))


### Bug Fixes

* NO-ISSUE - Fix releases ([#10](https://github.com/kubecano/cano-collector/issues/10)) ([a215330](https://github.com/kubecano/cano-collector/commit/a21533009f1da7004b7f094b1becec20fe727fe4))
