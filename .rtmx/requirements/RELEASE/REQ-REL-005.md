# REQ-REL-005: Docker Images

## Metadata
- **Category**: RELEASE
- **Subcategory**: Container
- **Priority**: MEDIUM
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-043

## Requirement

Release workflow shall publish multi-arch Docker images to GitHub Container Registry (ghcr.io), enabling containerized deployments.

## Design

### Dockerfile

```dockerfile
FROM alpine:latest
RUN apk add --no-cache ca-certificates git
COPY rtmx /usr/bin/rtmx
ENTRYPOINT ["rtmx"]
```

### GoReleaser Configuration

```yaml
dockers:
  - image_templates:
      - 'ghcr.io/rtmx-ai/rtmx:{{ .Version }}-amd64'
    use: buildx
    build_flag_templates:
      - "--platform=linux/amd64"
    goarch: amd64
  - image_templates:
      - 'ghcr.io/rtmx-ai/rtmx:{{ .Version }}-arm64'
    use: buildx
    build_flag_templates:
      - "--platform=linux/arm64"
    goarch: arm64

docker_manifests:
  - name_template: 'ghcr.io/rtmx-ai/rtmx:{{ .Version }}'
    image_templates:
      - 'ghcr.io/rtmx-ai/rtmx:{{ .Version }}-amd64'
      - 'ghcr.io/rtmx-ai/rtmx:{{ .Version }}-arm64'
  - name_template: 'ghcr.io/rtmx-ai/rtmx:latest'
    image_templates:
      - 'ghcr.io/rtmx-ai/rtmx:{{ .Version }}-amd64'
      - 'ghcr.io/rtmx-ai/rtmx:{{ .Version }}-arm64'
```

### Usage

```bash
docker run --rm -v $(pwd):/project -w /project ghcr.io/rtmx-ai/rtmx:latest status
```

## Acceptance Criteria

1. `docker pull ghcr.io/rtmx-ai/rtmx:latest` works
2. Multi-arch manifest covers linux/amd64 and linux/arm64
3. Image size < 20MB (alpine base + static binary)
4. `rtmx version` works inside container
5. Volume mount works for project access

## Files to Create/Modify

- `Dockerfile` - New file
- `.goreleaser.yaml` - Add `dockers:` and `docker_manifests:` sections
- `.github/workflows/release.yml` - Add `packages: write` permission (already present)
