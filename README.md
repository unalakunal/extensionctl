# extensionctl

Kaapana Extension Manager helper cli tool

## Make 

- `make build` generates the executable for the system it is run on \
- `make release` generates multiple executables for `darwin_amd64`, `linux_amd64`,  `linux_arm64` and `windows_amd64` \
- All the binaries can be found in `/build` folder.

## Build and save images as tar files

`extensionctl build <config-path>`

Here is how a config file looks like for the `hello-world` services in Kaapana examples.
`config.json`:
```
{
    "dockerfile_paths": ["/path/to/kaapana/templates_and_examples/examples/services/hello-world/docker/Dockerfile"], 
    "dir_path": "/path/to/kaapana/templates_and_examples/examples/services/hello-world", // root dir of extension
    "kaapana_path": "/path/to/kaapana", // root dir of Kaapana repo
    "kaapana_build_version": "0.0.0-latest", // version of your Kaapana instance, can be found in the bottom bar on the Kaapana website, such as "kaapana-admin-chart: 0.2.2"
    "custom_registry_url": "docker.io/kaapana" // registry url including project Gitlab template: "registry.<gitlab-url>/<group-or-user>/<project>"
}
```

* Run `extensionctl build image config.json` will save `images.tar` under the speficied `dir_path` in the config file. \
* This tar file can then be uploaded inside a Kaapana instance using the [extension upload component](https://kaapana.readthedocs.io/en/latest/user_guide/extensions.html#uploading-extensions-to-the-platform).

## TODO: build chart


## FAQ

### `Error searching and replacing in file`
One of the steps is to adapt the python files of the operators where the image is passed to KaapanaBaseOperator. The script will change `{DEFAULT_REGISTRY}` to `custom_registry_url` and `{KAAPANA_BUILD_VERSION}` to `kaapana_build_version`. If `{DEFAULT_REGISTRY}` and `{KAAPANA_BUILD_VERSION}` can not be found in the expected way inside py files, it is assumed that this linking is taken care of by the user. To omit this step of searching/replacing patterns use the flag `--no_overwrite_operators`



## Future work
- add support for using a registry url instead of local kaapana_path
- `--no_prereqs` flag (bool) disables building prereq images, assumes they are already built
- `--overwrite_file_extensions` flag (default: .py)
- `--overwrite_pattern` flag (default: {DEFAULT_REGISTRY},"docker.io/kaapana",{KAAPANA_BUILD_VERSION},"0.0.0-latest")
- `podman` support
- only save last layers if prerequisites are already available
- `--no_optimize` flag (bool) disable optimizing by saving only the last layer
