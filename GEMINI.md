# Project Instructions & Workflow Rules

## 1. Upstream PR Workflow (https://github.com/rairulyle/bitchord-selfhosted-addon)
- **Check PR status first:** Verify if the PR from `1ofx9:feat/navidrome-support` to `rairulyle:main` is open.
- **If open:** 
  - Incorporate all latest changes into `feat/navidrome-support`.
  - Update `CHANGELOG.md` under `## [Unreleased]` tailored to upstream.
  - Squash all commits on `feat/navidrome-support` into a single clean commit on top of `upstream/main`.
  - Force-push to `origin feat/navidrome-support`.
  - Provide the updated PR description in Markdown.
- **If closed:** STOP and ask the user what to do before proceeding.

## 2. Main Repository Workflow (https://github.com/1ofx9/bitchord-selfhosted-addon)
- **Push all files:** Commit and push all changes to `origin main`.
- **Release & Versioning:**
  - `CHANGELOG.md` on `main` tracks this repository's releases (`1.3.x`, `1.4.0`, etc.) and is distinct from the upstream PR changelog.
  - Before tagging or creating a release, update the changelog on `main`, commit and push, and **always ask the user if they want to cut a release and which version they want**.

## 3. General Principles
- Maintain zero external Go dependencies in `go.mod`.
- Ensure all tests pass across all packages (`go test -count=1 ./...`).
- Never hardcode personal domains, tokens, or credentials in production code or tests.
