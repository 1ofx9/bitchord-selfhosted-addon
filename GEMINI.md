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
- **Push all files:** Commit and push regular working files, code, tests, styles, and templates to `origin main`.
- **Release & Versioning:**
  - `CHANGELOG.md` on `main` tracks this repository's releases (`1.3.x`, `1.4.0`, etc.) and is distinct from the upstream PR changelog.
  - **Always handle CHANGELOG.md separately** in its own commit (never bundled with regular code changes).
  - Releases trigger via Git tags matching `v*`. When cutting a release, only tag the commit that contains the changelog update (`git tag -a v<version> -m "v<version>"`).
  - **Leave the push process to the user**: Never push release changelog commits or version tags to GitHub (`origin main`). Leave pushing to the user so they control the release and workflow trigger.

## 3. General Principles
- Maintain zero external Go dependencies in `go.mod`.
- Ensure all tests pass across all packages (`go test -count=1 ./...`).
- Never hardcode personal domains, tokens, or credentials in production code or tests.

