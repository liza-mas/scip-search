# User Stories: Validate Install Packaging Before Release

Status: review

## Goal
`scip-search` maintainers can run release-facing validation commands and packaging tests that prove supported distribution workflows install a verifiable binary without duplicating query or traversal validation.

## Parent Epic
specs/epics/readme/20260517-151311-epic-planning-5.md - Capability CAP-004

## Context
This document covers maintainer-facing distribution validation. It defines release validation commands, packaging tests, and drift checks that exercise installability, install location, source-build workflows, and `--version` verification for the supported README workflows. It intentionally excludes query behavior, SCIP traversal, real indexer execution, and release approval policy.

## Personas
- **Automation Agent**: an AI or script-driven caller running terminal commands in constrained environments, needing reproducible install commands and a quick way to verify the installed binary.
- **CLI Maintainer**: a Go developer preparing releases for macOS and Linux users, needing packaging checks and README examples that match the supported install behavior.
- **Source Builder**: a developer or early adopter with Go and make available, installing from a branch or local clone before a release artifact exists.

## General information

Applies to: release-facing validation commands, packaging tests, and distribution drift checks for supported install workflows.

### References
- goal spec: README.md#installation, lines 125-160 - Defines all user-facing installation and verification examples.
- goal spec: README.md#existing-indexers, lines 15-35 - Separates `scip-search` installation from language indexer prerequisites.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#general-information - Defines distribution-wide NFRs, external components, interfaces, assumptions, and out-of-scope boundaries.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-004---document-and-validate-distribution-workflows - Requires release validation commands and packaging tests focused on distribution workflows.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-001-01-release-install-script.md - Defines latest release, explicit `VERSION`, and custom `INSTALL_DIR` release install behavior.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-001-02-release-install-failure-and-verification.md - Defines release install failure handling and version verification.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-002-01-branch-install-script.md - Defines branch source install behavior and source-build failures.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-002-02-local-clone-make-install.md - Defines local clone `make install` behavior and source-build failures.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-01-version-output-contract.md - Defines top-level `scip-search --version` behavior.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-02-build-metadata-for-release-and-source-builds.md - Defines build identity for release and source builds.
- task: epic-planning-5-us-writing-3 - Requires packaging tests covering release install, source install, custom install directory, and `--version` workflows while excluding query-specific and traversal-specific validation.

### Non-Functional Requirements
- NFR-000-1: Distribution validation must be runnable by automation without an interactive UI.
- NFR-000-2: Packaging tests and release validation must not require large real-world repositories, real language indexer execution, generated SCIP indexes, query fixtures, traversal fixtures, or query golden JSON.
- NFR-000-3: Validation output must let maintainers distinguish install packaging failures from query runtime failures.
- NFR-000-4: Validation must cover release binary installation, source installation, custom install directories, and `--version` verification before a release is treated as distribution-ready.

### Related External Components
- Component C-001 - Release installer: The user-facing `install.sh` workflow invoked from the README for latest release, specific version, branch, and custom-directory installs.
- Component C-002 - Release artifacts: The published binaries or archives consumed by release install workflows.
- Component C-003 - Local Go and make toolchain: The caller-provided tools required for branch and local clone source builds.
- Component C-004 - README installation documentation: The user-facing installation and verification guidance maintained with distribution behavior.

### Interfaces
- I-001-001 - Installer invocation contract (Interface 001 of Component C-001): The shell invocation and environment options users rely on for release, branch, and custom-directory installs.
- I-003-001 - Source build contract (Interface 001 of Component C-003): The local clone build and install command surface used by source builders.
- I-004-001 - Version verification contract (Interface 001 of Component C-004): The command-line version output users and release automation use to confirm the installed executable.

### Out of Scope
- Query-specific fixtures, golden JSON cases, SCIP traversal fixtures, real indexer execution, or validation requiring a SCIP index.
- Query command behavior for `symbols`, `packages`, `references`, or `implementations`.
- Shared query runtime semantics for `--index`, JSON stdout, stderr diagnostics, or process exit statuses.
- Release approval process, changelog policy, hosted release provider setup, signing, notarization, package manager formulas, container images, and unsupported operating systems.
- Installing Go, make, or language indexers on the caller's host.

### Assumptions
- **ASM-000-1**: Packaging tests may use controlled release artifacts or locally produced build outputs as long as the user-observable install workflow matches the supported README command behavior. - *Why*: The capability requires packaging validation but excludes hosted release setup and release approval policy. - Confidence: MEDIUM
- **ASM-000-2**: A successful `--version` check is the distribution smoke-test proof for an installed binary; query commands belong to sibling query and traversal validation. - *Why*: The README uses `scip-search --version` for verification and the capability explicitly excludes query-specific validation. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Provide Release-Facing Validation Commands

### References
- goal spec: README.md#installation, lines 131-160 - Defines install commands and `scip-search --version` verification.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-004---document-and-validate-distribution-workflows - Requires release-facing validation commands.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-004-01-readme-installation-documentation.md - Defines README examples that validation commands must stay aligned with.

### User Story
**As a** CLI Maintainer preparing a release, **I want to** run a documented set of distribution validation commands for the README workflows, **so that** I can prove the install instructions produce a verifiable `scip-search` executable before release users follow them.

### Acceptance Criteria
- AC-001-1: Given release-facing validation commands are documented, when a maintainer reviews them, then they include a latest-release install command equivalent to `curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | bash` followed by `scip-search --version`.
- AC-001-2: Given release-facing validation commands are documented, when a maintainer reviews them, then they include an explicit released-version install command equivalent to `curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | VERSION=<release> bash` followed by `scip-search --version` for the installed executable.
- AC-001-3: Given release-facing validation commands are documented, when a maintainer reviews them, then they include a custom install directory command equivalent to `curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | INSTALL_DIR=<directory> bash` followed by `<directory>/scip-search --version`.
- AC-001-4: Given release-facing validation commands are documented, when a maintainer reviews source install validation, then they include a branch source install command equivalent to `curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | BRANCH=<branch> bash` followed by version verification of the installed executable.
- AC-001-5: Given release-facing validation commands are documented, when a maintainer reviews local source validation, then they include a local clone validation path equivalent to `git clone https://github.com/liza-mas/scip-search.git`, entering that clone, running `make install`, and running `scip-search --version`.
- AC-001-6: Given any documented distribution validation path completes successfully, when the maintainer reviews its success criteria, then success is based on an installed executable and `--version` output rather than a query command, SCIP index file, language indexer, traversal fixture, or query golden JSON.

### Depends on:
Implementation ordering:
- Story document CAP-004-01-readme-installation-documentation.md - README workflow examples must be defined before release validation commands can check alignment with them.
- Story document CAP-001-01-release-install-script.md - Release install behavior must exist before release validation can exercise it.
- Story document CAP-002-01-branch-install-script.md - Branch source install behavior must exist before source validation can exercise it.
- Story document CAP-002-02-local-clone-make-install.md - Local clone install behavior must exist before source validation can exercise it.
- Story document CAP-003-01-version-output-contract.md - Version verification behavior must exist before release validation can use it as success proof.

Run time coupling:
- Interface I-001-001 - Installer invocation contract
- Interface I-003-001 - Source build contract
- Interface I-004-001 - Version verification contract

### Out of Scope
- Release approval gates, changelog checks, signing/notarization checks, hosted release publication, or package-manager publication.
- Running `symbols`, `packages`, `references`, or `implementations` as release validation.
- Defining exact shell script names or CI provider configuration.

### Assumptions
- **ASM-001-1**: Validation commands can be documented in a maintainer-facing location separate from the public README as long as they are traceable to README installation examples. - *Why*: The capability requires release-facing validation commands but does not require them to be user-facing README content. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-002 - Test Packaging Workflows and Documentation Drift

### References
- goal spec: README.md#installation, lines 125-160 - Defines the user-facing workflows that packaging tests must cover.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-004---document-and-validate-distribution-workflows - Requires packaging tests and drift checks between docs and install behavior.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-001-02-release-install-failure-and-verification.md - Defines install failure and version verification outcomes that packaging tests can observe.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-02-build-metadata-for-release-and-source-builds.md - Defines release/source build identity expectations observable through `--version`.

### User Story
**As a** CLI Maintainer maintaining distribution packaging, **I want to** run packaging tests that compare supported install behavior with README examples, **so that** documentation and release packaging drift is caught before users copy stale commands.

### Acceptance Criteria
- AC-002-1: Given packaging tests run for release distribution, when they exercise release install behavior, then they cover latest-release install, explicit `VERSION` install, and custom `INSTALL_DIR` install outcomes.
- AC-002-2: Given packaging tests run for source distribution, when they exercise source install behavior, then they cover branch source install and local clone `make install` outcomes using caller-provided Go and make.
- AC-002-3: Given packaging tests install a binary through any supported workflow, when they verify the result, then they run that installed executable with `--version` and observe successful version output.
- AC-002-4: Given packaging tests inspect README installation documentation, when supported workflow examples drift from the installer, source build, or version verification contracts, then the tests fail with a distribution documentation mismatch rather than silently accepting stale examples.
- AC-002-5: Given packaging tests complete successfully, when a maintainer reviews their coverage, then the covered workflows include release install, source install, custom install directory, and `--version` verification.
- AC-002-6: Given packaging tests are run in a distribution validation environment, when they execute their checks, then they do not require generated SCIP indexes, real language indexer execution, query-specific fixtures, traversal fixtures, or query golden JSON.
- AC-002-6b: Given packaging tests need a failure signal for unsupported or unavailable install inputs, when they validate failure behavior, then they observe installer or source-build failure outcomes without changing expected behavior to make failing installs appear successful.

### Depends on:
Implementation ordering:
- Story document CAP-004-01-readme-installation-documentation.md - README examples must be defined before drift checks can compare them with behavior.
- Story document CAP-001-01-release-install-script.md - Release install behavior must exist before packaging tests can exercise release workflows.
- Story document CAP-001-02-release-install-failure-and-verification.md - Release verification and failure outcomes must exist before packaging tests can assert them.
- Story document CAP-002-01-branch-install-script.md - Branch source install behavior must exist before packaging tests can exercise it.
- Story document CAP-002-02-local-clone-make-install.md - Local clone install behavior must exist before packaging tests can exercise it.
- Story document CAP-003-01-version-output-contract.md - Version verification must exist before packaging tests can use `--version` as proof.
- Story document CAP-003-02-build-metadata-for-release-and-source-builds.md - Build identity behavior must exist before packaging tests can compare release and source provenance.

Run time coupling:
- Interface I-001-001 - Installer invocation contract
- Interface I-003-001 - Source build contract
- Interface I-004-001 - Version verification contract

### Out of Scope
- Query command result schemas, query runtime error contracts, or query fixture generation.
- Running real language indexers or requiring large repositories.
- Defining release approval policy, changelog policy, hosted release setup, signing, notarization, or CI provider selection.

### Assumptions
- **ASM-002-1**: Documentation drift checks can inspect README examples and compare them with supported command surfaces without executing every copied README command in exactly the public network context. - *Why*: The task asks for drift checks but excludes hosted release setup and release approval policy. - Confidence: MEDIUM

### Open Questions
- None.
