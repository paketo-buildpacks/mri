# Compile Ruby

To compile ruby, follow the below steps.

### Build the dockerfile:
```
docker build --platform <os>/<arch> --tag <target>-compile --file <target>.Dockerfile .
```

#### For Bionic:
```
docker build --tag bionic-compile --file bionic.Dockerfile .
```

#### For Jammy:
```
docker build --platform linux/arm64 --tag jammy-compile --file jammy.Dockerfile .
```

#### For Noble:
```
docker build --platform linux/arm64 --tag noble-compile --file noble.Dockerfile .
```

### Run compilation:
```
docker run --platform <os>/<arch> --volume $PWD:/tmp/compilation <target>-compile --outputDir /tmp/compilation --target <target> --version <version> --os <os> --arch <arch>
```

#### For Bionic:
```
docker run --volume $PWD:/tmp/compilation bionic-compile --outputDir /tmp/compilation --target bionic --version <version>
```

#### For Jammy:
**Note** that only Ruby 3.1.0 and above are supported on Jammy
```
docker run --platform linux/arm64 --volume $PWD:/tmp/compilation jammy-compile --outputDir /tmp/compilation --target jammy --version <version> --os linux --arch arm64
```

#### For Noble:
**Note** that only Ruby 3.1.0 and above are supported on Noble
```
docker run --platform linux/arm64 --volume $PWD:/tmp/compilation noble-compile --outputDir /tmp/compilation --target noble --version <version> --os linux --arch arm64
```
