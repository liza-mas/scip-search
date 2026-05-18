# User Stories: Keep Installation README Examples Accurate

Status: review

## Goal
`scip-search` maintainers can keep README installation guidance aligned with the supported release install, source install, custom install directory, and version verification workflows.

## Parent Epic
specs/epics/readme/20260517-151311-epic-planning-5.md - Capability CAP-004

## Context
The README is the user-facing entry point for installing `scip-search`. This document covers documentation stories for installation examples, prerequisite notes, and drift checks against the supported distribution behavior defined by CAP-001, CAP-002, and CAP-003. It does not define the installer, source build, or version command behavior itself.

## Personas
- **Automation Agent**: an AI or script-driven caller running terminal commands in constrained environments, needing reproducible install commands and a quick way to verify the installed binary.
- **CLI Maintainer**: a Go developer preparing releases for macOS and Linux users, needing README examples that match supported install behavior and avoid unsupported validation steps.
- **Source Builder**: a developer or early adopter with Go and make available, installing from a branch or local clone before a release artifact exists.

## General information

Applies to: README installation documentation and documentation drift checks for supported distribution workflows.

### References
- goal spec: README.md#installation, lines 125-160 - Defines prerequisites, latest release install, `VERSION`, `BRANCH`, `INSTALL_DIR`, local clone install, and `scip-search --version` verification.
- goal spec: README.md#existing-indexers, lines 15-35 - Separates language indexers from `scip-search` installation.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#personas - Defines Automation Agent, CLI Maintainer, and Source Builder personas.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#general-information - Defines distribution-wide NFRs, external components, interfaces, assumptions, and out-of-scope boundaries.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-004---document-and-validate-distribution-workflows - Defines README documentation and distribution validation scope.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-001-01-release-install-script.md - Defines release install examples for latest release, `VERSION`, and `INSTALL_DIR`.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-002-01-branch-install-script.md - Defines `BRANCH` source install behavior.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-002-02-local-clone-make-install.md - Defines local clone `make install` behavior.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-01-version-output-contract.md - Defines top-level `scip-search --version` verification behavior.
- consistency check: existing CAP-001, CAP-002, and CAP-003 story documents in specs/stories/readme/20260517-151311-epic-planning-5 - Read to keep documentation stories aligned with prior workflow boundaries.
- task: epic-planning-5-us-writing-3 - Requires README installation examples, user-facing prerequisite notes, release-facing validation commands, packaging tests, and drift checks while excluding query-specific and traversal-specific validation.

### Non-Functional Requirements
- NFR-000-1: README installation guidance must be automation-friendly: commands are copyable, non-interactive, and have observable verification steps.
- NFR-000-2: README prerequisite notes must keep `scip-search` installation separate from language indexer installation and SCIP index generation.
- NFR-000-3: README guidance must not direct users to validate installation by running query commands, real indexers, or traversal fixtures.
- NFR-000-4: Documentation drift checks must compare README examples with supported install behavior for CAP-001, CAP-002, and CAP-003.

### Related External Components
- Component C-001 - Release installer: The user-facing `install.sh` workflow invoked from the README for latest release, specific version, branch, and custom-directory installs.
- Component C-003 - Local Go and make toolchain: The caller-provided tools required for branch and local clone source builds.
- Component C-004 - README installation documentation: The user-facing installation and verification guidance maintained with distribution behavior.

### Interfaces
- I-001-001 - Installer invocation contract (Interface 001 of Component C-001): The shell invocation and environment options users rely on for release, branch, and custom-directory installs.
- I-003-001 - Source build contract (Interface 001 of Component C-003): The local clone build and install command surface used by source builders.
- I-004-001 - Version verification contract (Interface 001 of Component C-004): The command-line version output users and release automation use to confirm the installed executable.

### Out of Scope
- Implementing or changing `install.sh`, `make install`, release artifacts, or `scip-search --version`.
- Query command documentation for `symbols`, `packages`, `references`, or `implementations`.
- Query-specific fixtures, golden JSON cases, SCIP traversal fixtures, real indexer execution, or validation requiring a SCIP index.
- Unsupported package managers, container images, ctags fallback, MCP server usage, UI, graph visualization, embeddings, vector search, signing, notarization, hosted release setup, release approval process, and changelog policy.

### Assumptions
- **ASM-000-1**: README documentation can reference the supported behavior from CAP-001, CAP-002, and CAP-003 without restating every acceptance criterion from those story documents. - *Why*: CAP-004 is responsible for alignment and drift prevention, not redefining sibling workflow behavior. - Confidence: HIGH
- **ASM-000-2**: Documentation drift checks are expected to prove that README examples remain present and consistent with supported commands, not that the underlying installer implementation works. - *Why*: Packaging and release validation behavior is covered in the companion CAP-004 validation story document. - Confidence: HIGH

### Open Questions
- None.

---

## Story ST-001 - Document Supported Installation Commands

### References
- goal spec: README.md#installation, lines 131-154 - Defines quick release install, `VERSION`, `BRANCH`, `INSTALL_DIR`, and local clone examples.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#capability-cap-004---document-and-validate-distribution-workflows - Requires README installation examples to stay aligned with supported install workflows.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-001-01-release-install-script.md - Defines release install command behavior.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-002-01-branch-install-script.md - Defines branch install command behavior.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-002-02-local-clone-make-install.md - Defines local clone install command behavior.

### User Story
**As a** CLI Maintainer maintaining installation docs, **I want to** present one README installation example for each supported distribution workflow, **so that** users and automation can select the correct install command without relying on undocumented behavior.

### Acceptance Criteria
- AC-001-1: Given a maintainer reads the README installation section, when they inspect release install guidance, then the README shows a latest-release macOS/Linux install command equivalent to `curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | bash`.
- AC-001-2: Given a maintainer reads the README installation section, when they inspect pinned release guidance, then the README shows an explicit released-version installer example equivalent to `curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | VERSION=<release> bash`.
- AC-001-3: Given a maintainer reads the README installation section, when they inspect branch source-build guidance, then the README shows a branch installer example equivalent to `curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | BRANCH=<branch> bash` and states that Go and make are caller-provided prerequisites for that workflow.
- AC-001-4: Given a maintainer reads the README installation section, when they inspect custom directory guidance, then the README shows a custom-directory installer example equivalent to `curl -fsSL https://raw.githubusercontent.com/liza-mas/scip-search/main/install.sh | INSTALL_DIR=<directory> bash`.
- AC-001-5: Given a maintainer reads the README installation section, when they inspect local clone guidance, then the README shows the local clone workflow ending in `make install`.
- AC-001-6: Given a maintainer compares README install examples with CAP-001 and CAP-002 story scope, when all supported examples are present, then the README does not introduce unsupported package manager, container, signing, notarization, hosted release setup, or query-validation workflows as installation paths.

### Depends on:
Implementation ordering:
- Story document CAP-001-01-release-install-script.md - Supported release install examples must be known before documenting them.
- Story document CAP-002-01-branch-install-script.md - Supported branch source install behavior must be known before documenting it.
- Story document CAP-002-02-local-clone-make-install.md - Supported local clone install behavior must be known before documenting it.

Run time coupling:
- Interface I-001-001 - Installer invocation contract
- Interface I-003-001 - Source build contract

### Out of Scope
- Choosing release artifact URLs, archive layouts, checksum policy, source retrieval strategy, or make target internals.
- Adding documentation for package manager formulas, container images, signing, notarization, or release hosting setup.
- Validating query command behavior through README install examples.

### Assumptions
- **ASM-001-1**: README examples may use placeholder values such as `<release>`, `<branch>`, or `<directory>` as long as each supported workflow remains copyable after substitution. - *Why*: The README currently uses concrete examples, but the capability requirement is behavioral alignment rather than a specific placeholder style. - Confidence: MEDIUM

### Open Questions
- None.

---

## Story ST-002 - Document Prerequisites and Version Verification

### References
- goal spec: README.md#existing-indexers, lines 15-35 - Defines language indexers as separate prerequisites for generating SCIP indexes.
- goal spec: README.md#installation, lines 125-160 - Defines installation prerequisites and `scip-search --version` verification.
- parent epic: specs/epics/readme/20260517-151311-epic-planning-5.md#general-information - Requires distribution workflows to stay separate from SCIP indexer installation.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-01-version-output-contract.md - Defines version verification without an index.
- related story: specs/stories/readme/20260517-151311-epic-planning-5/CAP-003-02-build-metadata-for-release-and-source-builds.md - Defines release and source build provenance visible through version output.

### User Story
**As an** Automation Agent following README installation guidance, **I want to** see prerequisite notes and a version verification command that match the install workflow I used, **so that** I can confirm `scip-search` is installed without preparing SCIP indexes or running query commands.

### Acceptance Criteria
- AC-002-1: Given the README installation section is visible, when a caller reads prerequisite guidance, then it states that language indexers are separate tools needed for generating SCIP indexes and are not installed by the `scip-search` install workflow.
- AC-002-2: Given the README describes source install workflows, when a caller reads the `BRANCH` and local clone guidance, then the README identifies Go and make as caller-provided prerequisites for source builds.
- AC-002-3: Given the README describes release binary workflows, when a caller reads latest-release, `VERSION`, or `INSTALL_DIR` guidance, then the README does not state that Go, make, a local clone, language indexers, or SCIP indexes are required for successful release install verification.
- AC-002-4: Given any supported install workflow is documented, when the caller reaches the verification step, then the README shows `scip-search --version` as the installation verification command.
- AC-002-4b: Given a custom `INSTALL_DIR` example is documented, when the caller reaches the verification step for that workflow, then the README makes it clear that verification can invoke the executable from the requested install directory without relying on `PATH`.
- AC-002-5: Given the README verification guidance is present, when a maintainer compares it with CAP-003, then it does not require `--index`, a query subcommand, generated SCIP data, or network release lookup to verify installation.

### Depends on:
Implementation ordering:
- Story document CAP-003-01-version-output-contract.md - The version verification command must be defined before README verification guidance can rely on it.
- Story document CAP-003-02-build-metadata-for-release-and-source-builds.md - Release and source provenance expectations must be known before README verification guidance references installed build identity.

Run time coupling:
- Interface I-004-001 - Version verification contract

### Out of Scope
- Exact `--version` output wording, separators, or machine-readable schema.
- Teaching users how to generate SCIP indexes or install language-specific indexers beyond noting that those are separate prerequisites for query workflows.
- Query examples, traversal fixtures, or real indexer execution.

### Assumptions
- **ASM-002-1**: The README can keep indexer prerequisite guidance near installation as long as it clearly separates indexer setup from installing the `scip-search` binary. - *Why*: The goal spec currently places prerequisite notes before install commands. - Confidence: HIGH

### Open Questions
- None.
