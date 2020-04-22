# MRI Cloud Native Buildpack

The MRI CNB provides the Matz's Ruby Interpreter (or MRI).
The buildpack installs MRI onto the `$PATH` which makes it available
for subsequent buildpacks and in the final running container. It also sets
the `$GEM_PATH` environment variable.

## Integration

The MRI CNB provides `ruby` as a dependency. Downstream buildpacks,
can require the ruby dependency by generating
[Build Plan TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]
  # The name of the MRI dependency is "ruby". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "ruby"
  # The version of the MRI dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "2.*", "2.7.*", or even
  # "2.7.1".
  version = "2.7.1"
  # The MRI buildpack supports some non-required metadata options.
  [requires.metadata]
    # Setting the build flag to true will ensure that the MRI
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to use MRI
    # during its build process, this flag should be set to true.
    build = true
```

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.
