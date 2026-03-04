# Compile Ruby

To compile ruby, follow the below steps.

### Build the dockerfile:
```
docker build --tag <target>-compile --file <target>.Dockerfile .
```

#### For Jammy:
```
docker build --tag jammy-compile --file jammy.Dockerfile .
```

### Run compilation:
```
docker run --volume $PWD:/tmp/compilation <target>-compile --outputDir /tmp/compilation --target <target> --version <version>
```

#### For Jammy:

```
docker run --volume $PWD:/tmp/compilation jammy-compile --outputDir /tmp/compilation --target jammy --version <version>
```
