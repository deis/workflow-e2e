### v2.0.0-beta1 -> v2.0.0-beta2

#### Features

 - [`e914e7c`](https://github.com/deis/workflow-e2e/commit/e914e7cf4aca70bc6f0e09c056f39cf89243e247) dockerignore: add .dockerignore
 - [`80281d0`](https://github.com/deis/workflow-e2e/commit/80281d0fb2ef90cadfe77b31b1d21fdede407436) apps: track created apps and delete their namespaces after all tests are done
 - [`7f7d775`](https://github.com/deis/workflow-e2e/commit/7f7d77558a966fc5c4c763356e07598c298671b7) tests: reenable tests addressed by deis/controller #557
 - [`33f133d`](https://github.com/deis/workflow-e2e/commit/33f133df5be7ab3ff34a76d71d14a8a4dbefee7d) dns: change xip to nip

#### Fixes

 - [`9139411`](https://github.com/deis/workflow-e2e/commit/9139411ec1cd427e52fab9e99c9906b5db322ee2) (all): Use max timeout waiting for build:create to exit
 - [`606069e`](https://github.com/deis/workflow-e2e/commit/606069e70c5c681ff840567b437a0d322beba437) ps_test: remove ps:list check for 'restart "one"` type
 - [`3c30f4e`](https://github.com/deis/workflow-e2e/commit/3c30f4e7b9f382e88b0944090fadb69a781b6bb3) ci: Dockerize bootstrap during Travis build
 - [`97948fb`](https://github.com/deis/workflow-e2e/commit/97948fb2515c1e7c4eba8bea40c1e9f2627b5fd1) tags: regex for tags did not include - in names
 - [`8f9ccbc`](https://github.com/deis/workflow-e2e/commit/8f9ccbcc9911502d6a5677cdf1061ca7be6dc81e) tests_suite_test.go: remove sleep
 - [`9526090`](https://github.com/deis/workflow-e2e/commit/9526090f2bfa88c710af89657f13a24317710d7f) makefile: Needed to wrap TEST_OPTS and PARALLEL_TEST_OPTS in {}

#### Documentation

 - [`c56b6f6`](https://github.com/deis/workflow-e2e/commit/c56b6f6d49e6783dd4545a0ea2a6716e26c28a88) README: link to canonical Deis chart installation guide

#### Maintenance

 - [`f6efbc9`](https://github.com/deis/workflow-e2e/commit/f6efbc947e4c762549300bb2c89048304c3d5a7f) glide: update ginkgo dependencies
 - [`0c02565`](https://github.com/deis/workflow-e2e/commit/0c0256523b3a8b27df67ce5d96786cfaafc08f88) jobs: remove workflow-e2e-pr.groovy DSL as it will be superceded by representation in charts repo

### v2.0.0-beta1

#### Features

 - [`bb02abd`](https://github.com/deis/workflow-e2e/commit/bb02abdc9b8de0b190ab772123c9a2ed5b1c06d6) builds/ps_test: add a few more checks to build_test scenarios, cleanup beforeEach
 - [`838ab28`](https://github.com/deis/workflow-e2e/commit/838ab283113fa9d8f3ce617e7429bd048aa50ffe) domains_test: increase domains curl cmd timeout to 15 seconds
 - [`d60919a`](https://github.com/deis/workflow-e2e/commit/d60919a256f892d803547190371fe4415edc6027) .travis.yml: add docker-build step for ci build command
 - [`f3e53aa`](https://github.com/deis/workflow-e2e/commit/f3e53aa5c28dd44fa53bfe26935bf79a266bcf0b) perms_test: we do not need to deploy the app for current perms tests
 - [`b6c6679`](https://github.com/deis/workflow-e2e/commit/b6c66799abea1beea060f032aa2b6e554d11ba0f) certs_test: cover scenario described in https://github.com/deis/workflow/pull/492
 - [`5172da0`](https://github.com/deis/workflow-e2e/commit/5172da086f26dce294e984cbf5709a85edeb9cea) tags_tests.go: implement tags-related tests
 - [`af4fda8`](https://github.com/deis/workflow-e2e/commit/af4fda8c8880179e952d68d5eb069f83948c85cc) certs_tests e2e: add initial e2e certs tests
 - [`b729f27`](https://github.com/deis/workflow-e2e/commit/b729f27547209175bcc166f5f3a9accf1e3d4e77) app_test: add test to deploy a custom buildpack app
 - [`003e0e6`](https://github.com/deis/workflow-e2e/commit/003e0e6b12398707174caad8cdb556e8227cce3f) tests: add limits tests
 - [`1023dd8`](https://github.com/deis/workflow-e2e/commit/1023dd8ee0e426d11a2ac237381f3c62c73dc906) releases_test: finish implementing releases test
 - [`4fb0e40`](https://github.com/deis/workflow-e2e/commit/4fb0e408f0ae861a733adbebd7f97fdda1e3d3dc) timeout: set package level defaultMaxTimeout to 5 minutes
 - [`75b0a33`](https://github.com/deis/workflow-e2e/commit/75b0a335e4b74ffbcfb70b10901282911440354c) builds_tests: add deis builds command tests
 - [`957d891`](https://github.com/deis/workflow-e2e/commit/957d89161332af008d1c6c3de9b24de4b208b23d) Makefile: produce git sha based and canary type docker image tags
 - [`db950bb`](https://github.com/deis/workflow-e2e/commit/db950bb8a47033e8509918343140194ce0a40327) domains_test: workflow #448 fixes behavior; test needed updating
 - [`bef73c3`](https://github.com/deis/workflow-e2e/commit/bef73c3e4d97c3259c650a74ab0ced6da5d8da6b) tests: add apps:transfer, apps:run
 - [`76ffb8a`](https://github.com/deis/workflow-e2e/commit/76ffb8a1029c01a8d6a14731d5b8257e167f4722) domains_tests: add domains tests
 - [`a0d40bc`](https://github.com/deis/workflow-e2e/commit/a0d40bc44efdcf79037f2f0f011b6b038397c977) reporters: add junit reporter if JUNIT=true
 - [`2c8b26f`](https://github.com/deis/workflow-e2e/commit/2c8b26fedd987172fbd9e49fff832c8aa73b379a) jenkins: add DSL for the workflow-e2e-pr job
 - [`19f3efb`](https://github.com/deis/workflow-e2e/commit/19f3efbf744affb077bf8607480ad4e409eee637) ps_test: implement table-driven processes tests
 - [`3ccfa49`](https://github.com/deis/workflow-e2e/commit/3ccfa4906f11d49de70c5e92cf39df23cd92b443) perms_test: add perms:* invocations on deployed app
 - [`dcbd715`](https://github.com/deis/workflow-e2e/commit/dcbd715c7ec5f8c040fa6bc284471d789a15ed39) deploy.sh: remove hardcoded VERSION to use git-<sha> declared in Makefile
 - [`a1648e1`](https://github.com/deis/workflow-e2e/commit/a1648e18011991a98e336008a8265bbe31bf69bf) apps_test: add further apps:info expectations
 - [`7a01d24`](https://github.com/deis/workflow-e2e/commit/7a01d243936865b30fa91b439b3bfbead7d6c481) apps: enable deis open test
 - [`a8f1154`](https://github.com/deis/workflow-e2e/commit/a8f115436741a98697e89b552fcb2cafa845ba47) apps: verify that the app url can be opened
 - [`ab029fb`](https://github.com/deis/workflow-e2e/commit/ab029fba4a70196a1f4d17e29895c2c2776881c2) apps: verify that a bogus app url returns a 404
 - [`374b18e`](https://github.com/deis/workflow-e2e/commit/374b18e2be799027de08f3e55405b2333af0ef86) travis: add webhook to Jenkins e2e
 - [`58bbdfc`](https://github.com/deis/workflow-e2e/commit/58bbdfc6e4e51d083e7d8fcf9fc024b6f6f7ad7f) healthcheck_test.go: add healthcheck tests
 - [`555afee`](https://github.com/deis/workflow-e2e/commit/555afee31958f6990829ccce964ea4d5c486aa2d) manifests: add k8s readiness check
 - [`2a551a6`](https://github.com/deis/workflow-e2e/commit/2a551a690c5875376869d0f4eba92bf5e9fed668) ci: compile and upload darwin release
 - [`05d9c4a`](https://github.com/deis/workflow-e2e/commit/05d9c4af05c72eb44112c640e720d4af189a4c8d) secrets: add support for deis minio secrets
 - [`3a6cd62`](https://github.com/deis/workflow-e2e/commit/3a6cd62b9b823177bb1870d14a011f52828fafcb) _tests: add config tests
 - [`5502cf5`](https://github.com/deis/workflow-e2e/commit/5502cf50a68ff17ecef9da25c27a3d265831cb8f) travis: deploy client builds to bintray
 - [`2753dc6`](https://github.com/deis/workflow-e2e/commit/2753dc6da124bf4a51e0dd5999b2208077d5872d) glide: add glide.lock
 - [`0d94379`](https://github.com/deis/workflow-e2e/commit/0d9437972f5cdf6e87fdf07dffe686469b33a5cf) _tests: re-enabling the app tests
 - [`226cbc9`](https://github.com/deis/workflow-e2e/commit/226cbc94d204570b5f862e2f28c1dcbf09eaddc7) workflow: identify app type and move envs to rc templates

#### Fixes

 - [`9b8f599`](https://github.com/deis/workflow-e2e/commit/9b8f5996618748f6d91146a25f4dd6089ed0d45d) ps: loop until 200 on restart processes test
 - [`fca1fb2`](https://github.com/deis/workflow-e2e/commit/fca1fb29bdb3f18113c51c51a3103b110448c425) ps/builds/config: move consistently failing tests to pending
 - [`3b06613`](https://github.com/deis/workflow-e2e/commit/3b0661377274bef41fdaa13c4092d923dc7d91c0) releases_test: rollback to actual build+config release in test
 - [`d6d5317`](https://github.com/deis/workflow-e2e/commit/d6d5317a794d3caf3666e8a8bd693d26a639b6af) destroyApp: allow max timeout for "deis destroy"
 - [`0830add`](https://github.com/deis/workflow-e2e/commit/0830add59d304b3dcb62f2827887fffc2071be75) certs: this brings the necessary certs files into proper place for tests container
 - [`4771c01`](https://github.com/deis/workflow-e2e/commit/4771c01fecde5f517e573d6575bb5eae6041433a) createApp: move Say checks to createApp, depending on gitRemote bool
 - [`28a9164`](https://github.com/deis/workflow-e2e/commit/28a916411fb1476399fd458f8110cca8e24145c3) Dockerfile: pass same options to test binary
 - [`77228b7`](https://github.com/deis/workflow-e2e/commit/77228b7852763e3edf7d3dcc5697e0a94528eeff) jobs: use credentials binding for DOCKER/QUAY password
 - [`813acea`](https://github.com/deis/workflow-e2e/commit/813acea2cec75c9b358a853d0d09daeb5821cd2d) jobs: adjust workflow-e2e-pr job parameters + secrets
 - [`d4e5a58`](https://github.com/deis/workflow-e2e/commit/d4e5a58c5a32759d9c733bb14bac3759984e91d3) Makefile: increase overall test timeout to 60 min
 - [`6a7bc3b`](https://github.com/deis/workflow-e2e/commit/6a7bc3bf021db8ca37dbb0304b9b50ed1626aff0) tests: enable the non-ascii config:set test
 - [`fb71837`](https://github.com/deis/workflow-e2e/commit/fb71837069220a57992a7363f6789952eabe5476) config_test.go: restore config command tests
 - [`ce39a43`](https://github.com/deis/workflow-e2e/commit/ce39a435a511177c9252a9208c410df4886f7ee5) Makefile: increase "slow test" warning to 2 min
 - [`b40e0c4`](https://github.com/deis/workflow-e2e/commit/b40e0c4154a23501d2c5b766380283ce5b0cc5be) domains_test: actually test add output
 - [`d6d7793`](https://github.com/deis/workflow-e2e/commit/d6d779318aace886062e217de764e567e128e5ed) version_test: version now includes 7-character git sha
 - [`25e803f`](https://github.com/deis/workflow-e2e/commit/25e803fb2598998bfc4cf00686eea04a31385eec) domains_test: domains can be added / removed on apps before they deploy
 - [`cf4d599`](https://github.com/deis/workflow-e2e/commit/cf4d5999531867d29425b91a140fdd6854cb2100) domains_test: app domain is created when an app is created
 - [`be9bebb`](https://github.com/deis/workflow-e2e/commit/be9bebbba23a02a651450cbbf04d47cace7b2c09) jobs: prevent workflow-e2e-pr job from pushing canaries to registries
 - [`555d5ba`](https://github.com/deis/workflow-e2e/commit/555d5ba9a19c5e9d66b75ba18b662093ac56ea57) domains_test: domains:remove invocation pre-deploy returns 404, not 500
 - [`8ff32e0`](https://github.com/deis/workflow-e2e/commit/8ff32e0df288b04132983570115260ae1b99b8aa) tests_suite: print commands again
 - [`c8e9ecd`](https://github.com/deis/workflow-e2e/commit/c8e9ecd787ff579a844623202557c99c7c53f3d2) tests/ps_test.go: fix Intn panic
 - [`755af2a`](https://github.com/deis/workflow-e2e/commit/755af2a9e90f0bef311f51e4befa299f4da77350) ps_test: increase scale and restart timeouts to 5 min
 - [`529bda0`](https://github.com/deis/workflow-e2e/commit/529bda0929088c8d73b27ecc70d2620ac7e8acaa) .travis.yml: wire up travis job to jenkins workflow-e2e job
 - [`907f1d3`](https://github.com/deis/workflow-e2e/commit/907f1d310a71eb7d1e925541261d5812b872ec07) apps_test: update apps:info expectations
 - [`2b7f906`](https://github.com/deis/workflow-e2e/commit/2b7f906daf73f066953461ad392a0ebf30aaec94) cancel: delete apps before auth:cancel
 - [`b22c6c2`](https://github.com/deis/workflow-e2e/commit/b22c6c2a02f9a720e1625c552e3c377d77a78c7b) start: echo command line
 - [`6d854fa`](https://github.com/deis/workflow-e2e/commit/6d854fabb03cc8755a80f3b53c8faa364abc2c0f) apps_test: use WriteFile for system shim to ensure safe readability
 - [`dee37ec`](https://github.com/deis/workflow-e2e/commit/dee37ecaaf107926894097a0e546b7fa761afc53) apps_test: Fix the error message we check for
 - [`33cb52f`](https://github.com/deis/workflow-e2e/commit/33cb52f709fd0c3ee6fc2b76b7018f6f959e2a59) keys: list keys, create + remove temporary key
 - [`957ee10`](https://github.com/deis/workflow-e2e/commit/957ee10e2d08e6e73f1cf24810268b1606742e57) (all): prevent auth:cancels from colliding
 - [`1c53bdd`](https://github.com/deis/workflow-e2e/commit/1c53bdde992849da19f71c2fccd8b3d5333ec20c) tests: get subset of integration tests passing
 - [`d11523b`](https://github.com/deis/workflow-e2e/commit/d11523bf2c91eec01ee26ed424763ebe45301b0d) tests: clean up .git and fix config:list syntax
 - [`48aee74`](https://github.com/deis/workflow-e2e/commit/48aee742dbe65107d4b4a8865550734bc966be27) deploy.sh: push docker images to deisci/ orgs
 - [`a229403`](https://github.com/deis/workflow-e2e/commit/a229403620ae7fdc0426eff491bdef80ccf85571) tests: change package names
 - [`786f42b`](https://github.com/deis/workflow-e2e/commit/786f42b5c63f04b9d26636b8268a703cd9a1fcbc) .travis.yml,_scripts: add to CI
 - [`4a50921`](https://github.com/deis/workflow-e2e/commit/4a509212354be1adccae619993faccb1d8bde44a) Dockerfile,Makefile,glide.yaml,glide.lock: dockerize the dev and test environment
 - [`98ba34f`](https://github.com/deis/workflow-e2e/commit/98ba34f1f94bbd3fc9c2f36f6cdf2ddc43ed2f49) tests_suite_test.go: include constants, vars and func needed by healthcheck tests
 - [`a0e83a3`](https://github.com/deis/workflow-e2e/commit/a0e83a3d55d0a57434bde1219f85b6f14d99b4eb) tests: migrate test upgrades
 - [`ff6f8f4`](https://github.com/deis/workflow-e2e/commit/ff6f8f4627576d94b50a3eddd629de500ddabcf8) (all): migrate latest integration test changes from deis/workflow
 - [`0a5259e`](https://github.com/deis/workflow-e2e/commit/0a5259e6d7c43517353cf022302a9c4262b12272) docs: hotfix oneliner
 - [`c9a829d`](https://github.com/deis/workflow-e2e/commit/c9a829d1071d17935150d1ffc324d92db54ff6ec) Makefile/travis: use standard vars
 - [`28e38c5`](https://github.com/deis/workflow-e2e/commit/28e38c5deda59f6f5e26189d8d8ab24d19e8aabe) scheduler: terminate all pods from previous release on deploy
 - [`d91209d`](https://github.com/deis/workflow-e2e/commit/d91209d58282addc1b1499c4c20cdede6948a5d8) makefile: Update makefile and deploy.sh to work like the other repos
 - [`767431b`](https://github.com/deis/workflow-e2e/commit/767431b29d1b9167154a6f92325a26950d27a347) bintray: add uploadPattern to bintray json
 - [`44663a7`](https://github.com/deis/workflow-e2e/commit/44663a72e0e7f1bd7dc9e10791ea538d9b628b92) travis: deploy client on every push to master
 - [`4214e04`](https://github.com/deis/workflow-e2e/commit/4214e04225b848dbd8e895edf72bcc71c95fd9e3) k8s: pull image from private registry and inspect image
 - [`efd4d1d`](https://github.com/deis/workflow-e2e/commit/efd4d1dfc730f6295a5ee1bb9e9de3228e15c687) _tests: add admin user check tests
 - [`5ca102d`](https://github.com/deis/workflow-e2e/commit/5ca102d1adb69561f6305d9c40d84032b0f5228c) api: delete new_release only if it exists
 - [`8a637c6`](https://github.com/deis/workflow-e2e/commit/8a637c62c4e70d7f5f84617c016aad30433564bc) _tests: help setup testing environment
 - [`872a250`](https://github.com/deis/workflow-e2e/commit/872a250b4154f473d5a91d338eceb6ecd6441cd1) _tests: correct parameters for account cancel

#### Documentation

 - [`4404e08`](https://github.com/deis/workflow-e2e/commit/4404e08b7c9debac38191f93d138217ed38b8d9a) README: update client dev build instructions
 - [`e7a7fd6`](https://github.com/deis/workflow-e2e/commit/e7a7fd6381dd826532fb47701a08bc054a5532af) install: remove duplicate `describe pod` text
 - [`e5aa933`](https://github.com/deis/workflow-e2e/commit/e5aa933433609c838283194fc66575678a375d04) quickstart: remove untested providers
 - [`e31dc71`](https://github.com/deis/workflow-e2e/commit/e31dc71f77910e2147851c09a19fecb931003cea) readme: note that travis only builds linux clients
 - [`6d2a53d`](https://github.com/deis/workflow-e2e/commit/6d2a53d602229638524a73d0acf8255ffc3b0442) installing-deis: add domain configuration

#### Maintenance

 - [`a0d9929`](https://github.com/deis/workflow-e2e/commit/a0d99294dcb255fa6bfcb429ce205b820758c624) glide: update to go-dev 0.9 and glide 0.9
 - [`37f6a34`](https://github.com/deis/workflow-e2e/commit/37f6a347413b6da4403fa2587d7367b5b2dc8fba) tests: update for new Django REST framework
 - [`66c2974`](https://github.com/deis/workflow-e2e/commit/66c29748b276e2fa1e34e2071f62e01ec58a0736) cmd_test.go: remove dead code
 - [`70f6529`](https://github.com/deis/workflow-e2e/commit/70f6529643f5178c0186f2d595b4660b8e7bbd68) README: add pointer to docs
