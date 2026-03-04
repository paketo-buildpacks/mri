# Compile Ruby

To compile ruby, follow the below steps.

### Build the dockerfile:
```
docker build --platform <os>/<arch> --tag <target>-compile --file <target>.Dockerfile .
```

#### For Jammy (arm64):
```
docker build --platform linux/arm64 --tag jammy-compile --file jammy.Dockerfile .
```

#### For Noble (amd64):
```
docker build --platform linux/amd64 --tag noble-compile --file noble.Dockerfile .
```

### Run compilation:
```
docker run --platform <os>/<arch> --volume $PWD:/tmp/compilation <target>-compile --outputDir /tmp/compilation --target <target> --version <version> --os <os> --arch <arch>
```

#### For Jammy (arm64):

```
docker run --platform linux/arm64 --volume $PWD:/tmp/compilation jammy-compile --outputDir /tmp/compilation --target jammy --version 3.4.0 --os linux --arch arm64
```

#### For Noble (amd64):

```
docker run --platform linux/amd64 --volume $PWD:/tmp/compilation noble-compile --outputDir /tmp/compilation --target noble --version 3.4.0 --os linux --arch amd64
```
