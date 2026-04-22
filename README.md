# Stratus — OCI Registry

A monorepo with two components:

- **`cmd/stratus`** — Read-only OCI Distribution Spec v1 registry (Docker Registry V2 API) backed by S3. Images are served to Docker/containerd clients via blob redirect and manifest streaming.
- **`pkg/push`** — Go library for pushing a BuildKit OCI image to the S3 registry. Accepts an `fs.FS` produced by BuildKit's OCI exporter (`--output type=oci,dest=image.tar` or via `docker buildx build --output type=oci`; see [BuildKit OCI/Docker exporters](https://docs.docker.com/build/exporters/oci-docker/)), uploads blobs concurrently, and atomically merges the image index.

---

## cmd/stratus — Read-Only Registry API

### Implemented endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET, HEAD | `/v2/` | Registry support check (end-1) |
| GET, HEAD | `/v2/:ns/:repo/blobs/:digest` | Blob pull via presigned-URL redirect (end-2) |
| GET, HEAD | `/v2/:ns/:repo/manifests/:reference` | Manifest pull, tag or digest (end-3) |

Write endpoints (push, delete, catalog) are not implemented — this is intentional.

### Configuration

Set the following environment variables before starting:

```bash
S3_ENDPOINT=minio.example.com:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_BUCKET_NAME=zeabur-oci-registry   # optional, this is the default
S3_USE_SSL=false                      # set true for HTTPS
S3_REGION=us-east-1                   # optional
PORT=3000                             # optional
```

### Running with Docker

Prebuilt multi-platform images (`linux/amd64`, `linux/arm64`) are available on Docker Hub:

```bash
docker pull zeabur/stratus:2
```

```bash
docker run -p 3000:3000 \
  -e S3_ENDPOINT=minio.example.com:9000 \
  -e S3_ACCESS_KEY_ID=minioadmin \
  -e S3_SECRET_ACCESS_KEY=minioadmin \
  zeabur/stratus:2
```

To build from source instead:

```bash
docker build -t zeabur/stratus:2 .
```

### Building multi-platform images

[docker-bake.hcl](./docker-bake.hcl) builds `linux/amd64` and `linux/arm64` and tags the image as `2.1.2`, `2.0`, `2`, and `latest`.

```bash
# Build locally (no push)
VERSION=2.1.2 docker buildx bake

# Build and push to Docker Hub
VERSION=2.1.2 docker buildx bake --push
```

Override `REGISTRY` or `IMAGE` variables to target a different registry:

```bash
REGISTRY=ghcr.io IMAGE=your-org/stratus VERSION=2.1.2 docker buildx bake --push
```

### S3 bucket layout

The registry expects images to be stored using the following key structure:

```
<namespace>/<repository>/index.json
<namespace>/<repository>/blobs/sha256/<hex-digest>
```

`index.json` is an [OCI image index](https://github.com/opencontainers/image-spec/blob/main/image-index.md) with `schemaVersion: 2`.
Each manifest entry must include an `org.opencontainers.image.ref.name` annotation for tag-based pulls.

---

## pkg/push — OCI Image Pusher

`pkg/push` is a Go library that pushes a BuildKit OCI layout to the S3 registry. It is intended to be embedded in build pipelines that produce OCI images via BuildKit.

### How to produce a compatible OCI image

Use BuildKit's OCI exporter to produce an OCI layout directory or tarball:

```bash
# Export to a local directory (requires --output with type=local to untar first, or use a tar file)
docker buildx build --output type=oci,dest=image.tar .

# Unpack the tar into a directory for use with pkg/push
mkdir image-layout && tar -xf image.tar -C image-layout
```

See the [BuildKit OCI/Docker exporters documentation](https://docs.docker.com/build/exporters/oci-docker/) for full options.

### Usage

Install the module:

```bash
go get github.com/zeabur/stratus/v2
```

```go
import (
    "context"
    "fmt"
    "os"

    stratusconfig  "github.com/zeabur/stratus/v2/pkg/config"
    stratuspush    "github.com/zeabur/stratus/v2/pkg/push"
    stratusstorage "github.com/zeabur/stratus/v2/pkg/storage"
)

func pushImage(ctx context.Context, ociLayoutFS fs.FS, imageName, tag string) error {
    cfg := stratusconfig.Load()

    storage, err := stratusstorage.MinioStorageFromConfig(cfg)
    if err != nil {
        return fmt.Errorf("create storage: %w", err)
    }

    err = stratuspush.PushOciLayout(
        ctx,
        storage,
        cfg.BucketName,
        ociLayoutFS,
        imageName,
        tag,
        stratuspush.WithLogOutput(os.Stderr),
    )
    if err != nil {
        return fmt.Errorf("push oci layout to S3: %w", err)
    }

    return nil
}
```

**`ociLayoutFS`** must be an [`fs.FS`](https://pkg.go.dev/io/fs#FS) rooted at an OCI image layout directory — i.e. a directory containing `index.json`, `oci-layout`, and `blobs/sha256/`. Obtain one from a BuildKit export:

```bash
docker buildx build --output type=oci,dest=image.tar .
mkdir image-layout && tar -xf image.tar -C image-layout
```

Then pass it as `os.DirFS("image-layout")`.

**`config.Load()`** reads configuration from environment variables:

```bash
S3_ENDPOINT=minio.example.com:9000   # required — no scheme
S3_ACCESS_KEY_ID=minioadmin          # required
S3_SECRET_ACCESS_KEY=minioadmin      # required
S3_BUCKET_NAME=zeabur-oci-registry   # optional, this is the default
S3_USE_SSL=false                     # optional, default false
S3_REGION=us-east-1                  # optional
S3_PATH_STYLE=false                  # optional, set true for MinIO path-style access
```

**Available options for `PushOciLayout`:**

| Option | Default | Description |
|--------|---------|-------------|
| `WithLogOutput(w io.Writer)` | `os.Stderr` | Progress log destination |
| `WithBlobUploadConcurrency(n int)` | `4` | Number of blobs uploaded in parallel |

`PushOciLayout` uploads blobs concurrently (skipping any already present in S3), then atomically merges the local image index with the existing remote index before writing it back.

## Development

See [CLAUDE.md](./CLAUDE.md) for development commands and architecture notes.
