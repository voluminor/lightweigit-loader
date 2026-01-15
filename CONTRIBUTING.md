Thanks for your interest in contributing.

This repository focuses on being:
- lightweight
- dependency-free
- provider-agnostic (with strong GitHub support)

Ways to contribute:
- bug fixes
- improvements to provider detection
- adding/improving providers (GitLab, Bitbucket, Gogs-family, self-hosted variants)
- test coverage (URL parsing, edge cases, request building)
- documentation and examples

Development requirements:
- Go (recent stable version)

Local setup:
1) Fork the repository
2) Create a feature branch:
   go: git checkout -b feat/my-change
3) Make changes
4) Format code:
   go: gofmt -w .
5) Run tests:
   go: go test ./...

Pull request guidelines:
- explain what you are changing and why
- keep changes focused and minimal
- include tests for new or changed behavior
- avoid new dependencies unless absolutely necessary

Provider-related changes:
- include URL examples that should work
- add tests for detection + request building + parsing
- prefer explicit rules over heuristic guessing
- handle self-hosted domains carefully
- return meaningful errors

Compatibility note:
- GitHub support is production-verified
- other providers may be covered by tests but are not yet validated in real projects

Reporting issues:
- open a GitHub issue with a minimal reproduction
- include repository URL examples when detection is involved
- include expected vs actual behavior

Contact:
git@sunsung.fun
