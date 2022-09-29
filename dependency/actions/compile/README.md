# Compile Ruby

To compile ruby, follow the below steps.

### Build the dockerfile:
```
docker build --tag bionic-compile --file <target>.Dockerfile .
```

#### For Bionic:
```
docker build --tag bionic-compile --file bionic.Dockerfile .
```

#### For Jammy:
```
docker build --tag jammy-compile --file jammy.Dockerfile .
```

### Run compilation:
```
docker run --volume $PWD:/tmp/compilation <target>-compile --outputDir /tmp/compilation --target <target> --version <version>
```

#### For Bionic:
```
docker run --volume $PWD:/tmp/compilation bionic-compile --outputDir /tmp/compilation --target bionic --version <version>
```

#### For Jammy:
**Note** that only Ruby 3.1.0 and above are supported on Jammy
```
docker run --volume $PWD:/tmp/compilation jammy-compile --outputDir /tmp/compilation --target bionic --version <version>
```
