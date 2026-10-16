# ReleaseRail

ReleaseRail is a lightweight CI/CD automation and deployment reliability starter built with Go, Bash, Docker, and GitHub Actions.

## Local development

```bash
go test ./...
./scripts/deploy.sh staging releaserail:local --dry-run
docker build -t releaserail:local .
```

## CLI usage

```bash
go run ./cmd/releaserail healthcheck
```
