# Architecture Plan: Distribution Workflows

Status: review

## Goal

Implement the distribution-facing structure for installing, building, verifying, documenting, and validating `scip-search` without coupling those workflows to SCIP index generation or query execution.

## Context

`scip-search` is specified as a thin Go CLI, but this worktree currently contains planning artifacts, README documentation, pre-commit configuration, and no checked-in Go module, `install.sh`, or `Makefile`. This plan therefore defines the component boundaries that downstream code-planners should create or extend when implementation artifacts are introduced.

The distribution epic has four overlapping concerns: release installation, source installation, version verification, and documentation/validation. The key structural constraint is that install workflows prove success through the top-level `scip-search --version` command while query command routing, `--index`, JSON query output, SCIP traversal, and query fixtures remain owned by sibling epics.

### References

- Goal spec: `README.md#installation`, lines 125-160.
- Parent tasks: `epic-planning-5-us-writing-0`, `epic-planning-5-us-writing-1`, `epic-planning-5-us-writing-2`, `epic-planning-5-us-writing-3`.
- Epic: `specs/epics/readme/20260517-151311-epic-planning-5.md`.
- Story docs: `specs/stories/readme/20260517-151311-epic-planning-5/CAP-001-01-release-install-script.md`, `CAP-001-02-release-install-failure-and-verification.md`, `CAP-002-01-branch-install-script.md`, `CAP-002-02-local-clone-make-install.md`, `CAP-003-01-version-output-contract.md`, `CAP-003-02-build-metadata-for-release-and-source-builds.md`, `CAP-004-01-readme-installation-documentation.md`, `CAP-004-02-packaging-and-release-validation.md`.
- Sibling scope references: `specs/epics/readme/20260517-095535-epic-planning-1.outputs.json`, `specs/epics/readme/20260517-100328-epic-planning-2-output.json`, `specs/epics/readme/20260517-134857-epic-planning-3.outputs.json`, `specs/epics/readme/20260517-141006-epic-planning-4.outputs.json`.
- Codebase surveyed: `README.md`, `.pre-commit-config.yaml`, `.editorconfig`, `.gitignore`, `specs/epics/readme/20260517-151311-epic-planning-5.outputs.json`.

### Constraints

- The repository already has `.pre-commit-config.yaml`; no bootstrap-precommit output entry is emitted.
- Distribution workflows must not install Go, make, language indexers, generated SCIP indexes, query fixtures, or host-level tooling.
- `scip-search --version` is distribution verification output, not query JSON output.
- `--version` must run without `--index`, without a query command, without loading a SCIP index, and without network lookup.
- Release install behavior is scoped to macOS and Linux.
- Release hosting setup, signing, notarization, package manager formulas, containers, CI provider selection, and release approval policy are out of scope.
- The installer, source build tooling, version metadata, documentation, and packaging validation must produce automation-friendly success and failure signals.

### Assumptions

- **ASM-001**: The implementation will introduce or extend conventional Go CLI, shell installer, and Makefile artifacts even though they are not present in this planning worktree. - *Why*: README and stories specify a Go CLI, `install.sh`, and `make install`; no implementation files currently exist. - Confidence: HIGH
- **ASM-002**: Build identity can cross from release/source build tooling into the CLI through any implementation mechanism that supports offline `--version` output. - *Why*: CAP-003 requires provenance but explicitly does not prescribe linker flags, file formats, or metadata storage. - Confidence: HIGH
- **ASM-003**: Packaging tests may use controlled local artifacts or controlled source checkouts rather than hosted public releases where the same observable installer contract is exercised. - *Why*: CAP-004 excludes hosted release setup but allows controlled artifacts or locally produced build outputs. - Confidence: MEDIUM

### Open Questions

- **OQ-001**: The specs do not define behavior when both `VERSION` and `BRANCH` are set for the installer. - *Impact*: Code-planners should treat combined selector behavior as outside the required acceptance surface unless an implementation story explicitly adds a user-error contract.

---

## Components

### CLI Version Surface (`cmd/scip-search`, internal version boundary)

**Responsibility:** Own top-level `scip-search --version` routing and output of installed build identity.

**Boundaries:**
- Exposes: a top-level invocation that exits `0`, writes non-empty version verification output to stdout, leaves stderr empty, and does not require query command state.
- Depends on: build identity supplied by build/release/source tooling.

**Key decisions:**
- Keep version handling before query runtime validation: installation verification must not trigger missing-command, missing-index, or SCIP loading failures.
- Treat version output as a separate distribution response: query JSON stdout contracts remain owned by query runtime stories.
- Represent build identity behind a narrow version boundary: installer and build tooling provide identity, while CLI invocation owns displaying it offline.

### Build Identity Producer (`Makefile`, release build commands, source build commands)

**Responsibility:** Produce `scip-search` binaries whose embedded or associated identity distinguishes release builds from source builds.

**Boundaries:**
- Exposes: build outputs that the CLI version surface can identify as release or source builds.
- Depends on: caller-provided Go and make for source workflows; release workflow inputs for release artifacts.

**Key decisions:**
- Keep metadata mechanism implementation-local: the architecture requires the identity contract, not specific flags or storage.
- Source builds must not masquerade as releases: source provenance must be explicit when no release identity is available.
- Release build identity must be comparable with the selected release without network lookup during `--version`.

### Installer Entry Point (`install.sh`)

**Responsibility:** Select and execute the supported user-facing install workflow for release binaries, branch source builds, explicit versions, and custom install directories.

**Boundaries:**
- Exposes: environment-driven installer invocation using `VERSION`, `BRANCH`, and `INSTALL_DIR`, with observable installed executable path or actionable nonzero failure.
- Depends on: release artifacts for release workflows, caller-provided Go and make for `BRANCH`, and install filesystem permissions for the selected directory.

**Key decisions:**
- Keep release and branch workflows in one entry point: the README exposes one installer script, so selection belongs at that boundary.
- Preserve release/source separation: `VERSION` selects released artifacts; `BRANCH` selects a source build path requiring Go and make.
- Success must name the installed executable path: automation needs a direct target for `<INSTALL_DIR>/scip-search --version` when PATH is not involved.
- Failure diagnostics remain human-readable and nonzero, not a new machine-readable installer schema.

### Local Source Install Tooling (`Makefile`)

**Responsibility:** Build and install `scip-search` from the caller's current local checkout through `make install`.

**Boundaries:**
- Exposes: a local clone install command that builds the current checkout and installs a runnable `scip-search`.
- Depends on: caller-provided Go and make, local checkout contents, and the build identity producer.

**Key decisions:**
- `make install` owns local checkout installation, not release artifact lookup.
- Source prerequisite failures must fail before claiming install success.
- Installed binaries must be runnable outside the clone directory for verification.

### Distribution Documentation (`README.md`, release validation docs)

**Responsibility:** Keep user-facing installation examples and maintainer-facing validation commands aligned with supported workflows.

**Boundaries:**
- Exposes: copyable README install examples and traceable maintainer release validation commands.
- Depends on: installer, source install, and version contracts.

**Key decisions:**
- Documentation describes supported workflows only: no package managers, containers, signing, hosted release setup, or query-validation install paths.
- README prerequisite text separates language indexers from installing `scip-search`.
- Custom install directory verification must show direct executable invocation when PATH is not guaranteed.

### Distribution Validation (`tests` or release validation scripts)

**Responsibility:** Exercise supported distribution workflows and catch drift between README examples, installer/source build behavior, and `--version` verification.

**Boundaries:**
- Exposes: automated packaging tests and release-facing validation commands focused on installability and build identity.
- Depends on: controlled release artifacts or locally produced build outputs, source build tooling, installer, and README docs.

**Key decisions:**
- Validation success is an installed executable plus `--version`, not a query command.
- Tests must not require real language indexers, generated SCIP indexes, traversal fixtures, query golden JSON, large repositories, or network lookup during `--version`.
- Drift checks compare documented command surfaces with supported behavior and fail as distribution documentation mismatches.

---

## Interfaces

### Installer Entry Point -> Release Artifacts

**Contract:** The installer maps latest-release or `VERSION=<release>` inputs to a release artifact compatible with the current macOS/Linux environment, installs `scip-search`, and reports the installed path.
**Direction:** Installer fetches/selects artifact; installed executable is the output.
**Invariants:** Release workflows do not require local clone, Go, make, language indexers, or SCIP indexes.

### Installer Entry Point -> Source Build Tooling

**Contract:** `BRANCH=<branch>` switches the installer to a source build workflow that requires caller-provided Go and make and produces an installed executable from the requested branch.
**Direction:** Installer selects source path and invokes build/install tooling.
**Invariants:** Branch workflows do not silently fall back to release artifacts and do not provision Go or make.

### Local Source Install Tooling -> CLI Version Surface

**Contract:** Local `make install` builds the current checkout and supplies available source provenance to the binary so `scip-search --version` can identify it as source-built.
**Direction:** Make/build tooling produces binary identity; CLI reports it.
**Invariants:** Source builds do not claim release provenance unless the build workflow explicitly supplies a release identity.

### Release Build Tooling -> CLI Version Surface

**Contract:** Release build automation supplies release identity to the binary so `scip-search --version` identifies the executable as a release build and includes the selected release identity.
**Direction:** Release build tooling produces binary identity; CLI reports it.
**Invariants:** `--version` performs no network lookup to confirm release identity.

### Documentation -> Distribution Validation

**Contract:** README and release validation examples provide the command surfaces validation checks compare against installer, source install, and version contracts.
**Direction:** Docs define expected user-facing commands; tests validate presence, alignment, and workflow smoke behavior.
**Invariants:** Documentation and validation stay distribution-scoped and do not introduce query validation requirements.

---

## Data Flow

### Release Install

```text
README curl command
  -> install.sh workflow selector
  -> latest release or VERSION release artifact
  -> install scip-search into default INSTALL_DIR or requested INSTALL_DIR
  -> report installed executable path
  -> installed scip-search --version
  -> stdout release build identity, exit 0
```

### Branch Source Install

```text
README curl command with BRANCH
  -> install.sh workflow selector
  -> prerequisite check for caller-provided Go and make
  -> fetch/build requested branch source
  -> install scip-search into selected binary directory
  -> installed scip-search --version
  -> stdout source build provenance, exit 0
```

### Local Clone Install

```text
local checkout
  -> make install
  -> prerequisite check for caller-provided Go and make
  -> build current checkout with source provenance
  -> install scip-search into make-selected binary directory
  -> installed scip-search --version outside clone
  -> stdout source build provenance, exit 0
```

### Documentation And Packaging Validation

```text
README installation examples
  -> distribution drift checks
  -> controlled release/source packaging smoke checks
  -> installed executable --version
  -> validation success or distribution-scoped failure diagnostics
```

---

## Cross-Cutting Concerns

| Concern | Approach |
|---------|----------|
| Error handling | Installer and source tooling return nonzero on unsupported platform, unavailable release/version/branch, unusable install directory, missing Go/make, build failure, or install failure. Diagnostics identify the failing user input or prerequisite without introducing query runtime error contracts. |
| Observability | Successful install paths report the installed executable path and selected workflow identity. Validation commands use process status plus `--version` output as the observable proof. |
| Configuration | User-facing configuration remains environment-based for `VERSION`, `BRANCH`, and `INSTALL_DIR`; source build prerequisites come from caller PATH. |
| Testing | Distribution tests cover release install, explicit version, custom install directory, branch source install, local clone install, docs drift, failure outcomes, and `--version` without SCIP indexes or query fixtures. |
| Security | Installer must not execute host provisioning for Go, make, language indexers, package managers, or global tools. Install paths and diagnostics should avoid leaking secrets and should treat external release/source inputs as untrusted. |
| Scope isolation | Version verification bypasses query routing and index loading; packaging validation avoids generated SCIP indexes, traversal fixtures, query command execution, and query golden JSON. |

---

## Structural Decisions

1. **Version first, then installers.** `--version` is the common verification boundary for all distribution workflows, so downstream implementation should establish the CLI version surface and build identity contract before installer validation relies on it.
2. **One installer entry point, separate workflow branches.** The README exposes one `install.sh` command surface, so release and branch paths share selector/error infrastructure while preserving release-vs-source behavior.
3. **Local clone install stays in build tooling.** `make install` is a local source workflow and should not be coupled to release artifact discovery.
4. **Docs and validation depend on behavior contracts.** README and release validation should be updated after installer/source/version contracts are known so examples and tests do not drift or invent unsupported workflows.
5. **Distribution validation remains smoke-level.** A successful installed executable plus `--version` is sufficient for this epic; query correctness belongs to sibling epics.

---

## Decomposition

Each scope becomes a code-planning child task.

### Scope 1: Version and build identity

**Component(s):** CLI Version Surface; Build Identity Producer.
**Boundary:** Top-level `scip-search --version` behavior, stdout/stderr/status contract, offline build identity, release identity, and source provenance are in scope. Query command routing details beyond bypassing missing query and missing index failures are out of scope.
**Desc:** `scip-search` users and automation can run a top-level `scip-search --version` command that reports offline build identity for release and source builds without requiring an index or query runtime.
**Done when:** `scip-search --version` is accepted without a query command or `--index`, exits `0`, writes non-empty `scip-search` build identity to stdout with empty stderr, distinguishes release identity from source provenance, and does not load SCIP data or perform network lookup.
**Scope:** CLI version invocation behavior, build identity contract between release/source build tooling and the binary, release/source provenance output, and tests for no-index version verification. Excludes query JSON schemas, query runtime error taxonomy, remote release lookup, installer workflow selection, README documentation, and packaging validation.
**Depends on:** None.

### Scope 2: Release installer workflows

**Component(s):** Installer Entry Point; Release Build Tooling; Distribution Validation test hooks for release install behavior.
**Boundary:** Latest release, explicit `VERSION`, `INSTALL_DIR`, installed executable path reporting, executable permission, and release install failures are in scope. Branch builds, local clone builds, and query validation are out of scope.
**Desc:** `scip-search` users can install latest or explicit released binaries on macOS/Linux into default or requested directories through `install.sh` and receive observable success or actionable release-install failure.
**Done when:** The installer supports latest-release, `VERSION=<release>`, and `INSTALL_DIR=<directory>` release workflows on macOS/Linux; successful installs produce an executable `scip-search` at an observable path whose `--version` identifies the release; unsupported platform, unavailable release/version, and unusable install directory failures exit nonzero without claiming success.
**Scope:** `install.sh` release workflow selection, release artifact installation boundary, default and custom install directory behavior, executable availability, release install diagnostics, and release-install tests using `--version` for verification. Excludes `BRANCH`, local clone `make install`, source provenance, release hosting setup, signing, package managers, language indexer installation, SCIP index generation, and query command execution.
**Depends on:** Scope 1.

### Scope 3: Source install workflows

**Component(s):** Installer Entry Point; Local Source Install Tooling; Build Identity Producer.
**Boundary:** `BRANCH` installer builds, caller-provided Go/make prerequisite handling, local clone `make install`, installed source-built executable outcome, and source-build failures are in scope. Release artifact selection and query validation are out of scope.
**Desc:** `scip-search` contributors and early adopters can build and install from a requested branch or local clone using caller-provided Go and make, with installed source-built binaries verifiable outside the source checkout.
**Done when:** `BRANCH=<branch>` installer workflow and local clone `make install` build from source with caller-provided Go and make, install an executable `scip-search`, expose source-build provenance through `--version`, fail nonzero for missing Go, missing make, unavailable branch, build failure, or install failure, and never provision host tools or silently use release artifacts for source workflows.
**Scope:** `install.sh` branch workflow, source prerequisite checks, local clone `make install`, source build/install output, source-build version provenance handoff, and source-install tests. Excludes release binary download behavior, release artifact fallback, cross-compilation, CI publication, package manager formulas, Go/make installation, language indexer installation, SCIP indexes, and query command validation.
**Depends on:** Scope 1, Scope 2.

### Scope 4: Distribution documentation and packaging validation

**Component(s):** Distribution Documentation; Distribution Validation.
**Boundary:** README installation examples, prerequisite notes, custom directory verification guidance, release-facing validation commands, packaging tests, and docs drift checks are in scope. Implementing installer internals, version internals, and query/traversal validation are out of scope.
**Desc:** `scip-search` maintainers can keep README installation guidance, release-facing validation commands, packaging tests, and drift checks aligned with supported release install, source install, custom directory, and `--version` workflows.
**Done when:** README installation guidance documents latest release, explicit `VERSION`, `BRANCH`, custom `INSTALL_DIR`, local clone `make install`, source prerequisites, separate language-indexer prerequisites, and `--version` verification; release validation commands and packaging tests cover release install, source install, custom install directory, actionable failure signals, docs drift, and version verification without query-specific or traversal-specific validation.
**Scope:** README installation documentation, maintainer release validation command documentation, packaging tests for distribution workflows, README-vs-behavior drift checks, and distribution-scoped failure reporting. Excludes changing installer/source/version behavior, query-specific fixtures, traversal fixtures, real indexer execution, unsupported packaging channels, release approval policy, changelog policy, hosted release setup, signing, and notarization.
**Depends on:** Scope 1, Scope 2, Scope 3.

### Spec Coverage

| Spec Requirement | Scope |
|------------------|-------|
| Latest release install on macOS/Linux | Scope 2 |
| Explicit `VERSION` release install | Scope 2 |
| Custom `INSTALL_DIR` release install and direct verification | Scope 2, Scope 4 |
| Installed executable success and path observability | Scope 2, Scope 3 |
| Actionable release install failures | Scope 2, Scope 4 |
| `BRANCH` installer source build requiring Go and make | Scope 3 |
| Local clone `make install` source workflow | Scope 3 |
| Source workflow prerequisite failures without host tool provisioning | Scope 3 |
| Top-level `scip-search --version` without index | Scope 1 |
| Release and source build provenance in version output | Scope 1, Scope 3 |
| Install verification without local clone, SCIP index, language indexers, or query execution | Scope 1, Scope 2, Scope 3, Scope 4 |
| README install examples and prerequisite notes | Scope 4 |
| Release-facing validation commands | Scope 4 |
| Packaging tests for release, source, custom directory, and `--version` workflows | Scope 4 |
| Exclusion of query-specific and traversal-specific validation | Scope 4 |
