# Compile Ruby

## Important: Ruby Version and Stack Compatibility

**Ruby 4.x can only be compiled on Noble**

Ruby 4.x requires GLIBC 2.38 or later. The Jammy stack only provides GLIBC 2.35.
Attempting to compile Ruby 4.x for Jammy will fail with an error message.

Stack GLIBC versions:
- Jammy: GLIBC 2.35
- Noble: GLIBC 2.39

For Jammy deployments, use Ruby 3.x versions (3.2 and later).

## Compilation Steps

To compile ruby, follow the below steps.

### Build the dockerfile:
```
docker build --platform <os>/<arch> --tag <target>-compile --file <target>.Dockerfile .
```

#### For Jammy (arm64 - Ruby 3.x only):
```
docker build --platform linux/arm64 --tag jammy-compile --file jammy.Dockerfile .
```

#### For Noble (amd64 - any Ruby version):
```
docker build --platform linux/amd64 --tag noble-compile --file noble.Dockerfile .
```

### Run compilation:
```
docker run --platform <os>/<arch> --volume $PWD:/tmp/compilation <target>-compile --outputDir /tmp/compilation --target <target> --version <version> --os <os> --arch <arch>
```

#### For Jammy (arm64 - Ruby 3.x only):

```
docker run --platform linux/arm64 --volume $PWD:/tmp/compilation jammy-compile --outputDir /tmp/compilation --target jammy --version 3.4.0 --os linux --arch arm64
```

#### For Noble (amd64 - any Ruby version):

```
docker run --platform linux/amd64 --volume $PWD:/tmp/compilation noble-compile --outputDir /tmp/compilation --target noble --version 4.0.1 --os linux --arch amd64
```
