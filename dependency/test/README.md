### To test locally:

1. From the `<buildpack>/dependency` directory:
```
make test tarballPath="path/to/ruby.tgz" version="<version>"
```
2. The make target will build Docker containers, mount the artifact into them,
   and run a series of tests.
3. If the make target commpletes without error, the tests have passed.

### Important Version and Stack Compatibility Notes:

- **Ruby 4.x**: Can only be tested on Noble (requires GLIBC 2.38+)
- **Ruby 3.x**: Can be tested on both Jammy and Noble
- Attempting to test Ruby 4.x against Jammy will fail with GLIBC version errors
