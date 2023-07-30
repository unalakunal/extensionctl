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
    "kaapana_build_version": "0.0.0-latest" // version of your Kaapana instance, can be found in the bottom bar on the Kaapana website, such as "kaapana-admin-chart: 0.2.2"
}
```

* Run `extensionctl build image config.json` will save `images.tar` under the speficied `dir_path` in the config file. \
* This tar file can then be uploaded inside a Kaapana instance using the [extension upload component](https://kaapana.readthedocs.io/en/latest/user_guide/extensions.html#uploading-extensions-to-the-platform).

## TODO: build chart
