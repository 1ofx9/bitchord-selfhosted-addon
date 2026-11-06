# Project Instructions & Workflow Rules

## 1. Upstream PR Workflow (https://github.com/rairulyle/bitchord-selfhosted-addon)
- **Check PR status first:** Verify if the PR from `1ofx9:feat/navidrome-support` to `rairulyle:main` is open.
- **If open:** 
  - Update the changelog on the PR under `## [Unreleased]` documenting all additions/changes fully, and **always confirm the Changelog with the user first before finalizing**.
  - Incorporate all latest changes into `feat/navidrome-support`.
  - Squash all commits on `feat/navidrome-support` into a single clean commit on top of `upstream/main`.
  - Force-push to `origin feat/navidrome-support`.
  - Provide the updated PR description in Markdown.
- **If closed:** STOP and ask the user what to do before proceeding.

## 2. Main Repository Workflow (https://github.com/1ofx9/bitchord-selfhosted-addon)
- **Push all files:** Commit and push regular working files and ongoing changes (under `## [Unreleased]`) to `origin main`.
- **Release & Versioning:**
  - `CHANGELOG.md` on `main` tracks this repository's releases (`1.3.x`, `1.4.0`, etc.) and is distinct from the upstream PR changelog.
  - **CRITICAL**: Committing a release version header in `CHANGELOG.md` on `main` triggers a GitHub Action that cuts a release and builds Docker images.
  - **NEVER** commit a release changelog version without explicit user approval.
  - Always ask the user if they want to cut a release and which version they want first, and **only proceed once the user approves**.

## 3. General Principles
- Maintain zero external Go dependencies in `go.mod`.
- Ensure all tests pass across all packages (`go test -count=1 ./...`).
- Never hardcode personal domains, tokens, or credentials in production code or tests.

