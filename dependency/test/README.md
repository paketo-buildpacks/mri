### To test locally:

1. From the `<buildpack>/dependency` directory:
```
make test tarballPath="path/to/ruby.tgz" version="<version>"
```
2. The make target will build Docker containers, mount the artifact into them,
   and run a series of tests.
3. If the make target commpletes without error, the tests have passed.
