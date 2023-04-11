# Paketo Buildpack for MRI

## `gcr.io/paketo-buildpacks/mri`

The MRI CNB provides the Matz's Ruby Interpreter (or MRI).
The buildpack installs MRI onto the `$PATH` which makes it available
for subsequent buildpacks and in the final running container. It also sets
the `$GEM_PATH` environment variable.

## Integration

The MRI CNB provides `mri` as a dependency. Downstream buildpacks,
can require the mri dependency by generating
[Build Plan TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]
  # The name of the MRI dependency is "mri". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "mri"
  # The version of the MRI dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "3.*", "3.2.*", or even
  # "3.2.1".
  version = "3.2.1"
  # The MRI buildpack supports some non-required metadata options.
  [requires.metadata]
    # Setting the build flag to true will ensure that the MRI
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to use MRI
    # during its build process, this flag should be set to true.
    build = true
```

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```

## MRI Configurations

Specifying the `MRI` version through `buildpack.yml` configuration will be
deprecated in MRI Buildpack v1.0.0.

To migrate from using `buildpack.yml` please set the `$BP_MRI_VERSION`
environment variable at build time either directly (ex. `pack build my-app
--env BP_MRI_VERSION=3.2.*`) or through a [`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md)

```shell
$BP_MRI_VERSION="3.2.1"
```
This will replace the following structure in `buildpack.yml`:
```yaml
mri:
  version: 3.2.1
```

## Logging Configurations

To configure the level of log output from the **buildpack itself**, set the
`$BP_LOG_LEVEL` environment variable at build time either directly (ex. `pack
build my-app --env BP_LOG_LEVEL=DEBUG`) or through a [`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md)
If no value is set, the default value of `INFO` will be used.

The options for this setting are:
- `INFO`: (Default) log information about the progress of the build process
- `DEBUG`: log debugging information about the progress of the build process

```shell
$BP_LOG_LEVEL="DEBUG"
```

## Compatibility

This buildpack is currently only supported on the Paketo Bionic and Jammy stack
distributions. Pre-compiled distributions of Ruby are provided for the Paketo stacks (i.e.
`io.buildpacks.stack.jammy` and `io.buildpacks.stacks.bionic`).

Jammy stack support only applies to Ruby version 3.1 and above at this time.

## Development

Paketo buildpacks are going through an uniformization of the dev experience across buildpacks,
for now just check the [`scripts/`](scripts/) folder.
