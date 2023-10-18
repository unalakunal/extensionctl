# extensionctl

Kaapana Extension Manager helper cli tool

## Make 

- `make build` generates an executable for the system it runs on
- `make release` generates multiple executables for `darwin_amd64`, `linux_amd64`,  `linux_arm64` and `windows_amd64`
- All the binaries can be found under `/build` folder.

## Build extension

### 1. Requirements

#### A. Container Engine and Helm
Helm and a container engine (currently supporting Docker and Podman) are necessary for the script to run. Make sure to install the latest versions of each.

#### B. Having the right extension folder structure

Kaapana supports two different types of extensions, workflows and services. Both have slightly different folder structures. For more details, refer to the [docs](https://kaapana.readthedocs.io/en/stable/user_guide/extensions.html)

1. Folder structure for workflows, example workflow described here is [otsus-method](https://github.com/kaapana/kaapana/tree/develop/templates_and_examples/examples/processing-pipelines/otsus-method) in Kaapana examples:
```
/Users/unal/Documents/dev/kaapana/templates_and_examples/examples/processing-pipelines/otsus-method/
├── extension // everything related to the extension
│   ├── docker // docker folder for Airflow DAG, should also contain custom Operators used in the DAG
│   │   ...
│   └── otsus-method-workflow // for helm chart, requirements should contain dag-installer-chart 
│       ...
└── processing-containers // scripts and algorithms used in the workflow
    └── otsus-method
        ...
```
2. Folder structure for services, example is [hello-world-service](https://github.com/kaapana/kaapana/tree/develop/templates_and_examples/examples/services/hello-world) in Kaapana examples:
```
├── docker // contains Dockerfile and additional source code
│   ...
└── hello-world-chart // for helm chart, should contain Kubernetes resource files such as job.yaml or deployment.yaml
    ...
```

#### C. Kaapana repository
For now, it is required that [Kaapana repository](https://github.com/kaapana/kaapana) is already cloned.



### 2. Edit config file

extensionctl takes a config json file as input. Use `config-template.json` as the template and fill in the values.

Here is how a config file looks like for the `otsus-method` explained above.

`config.json`:
```
{
    "dockerfile_paths": [], // if empty or removed, all dockerfiles will be found across kaapana_path and appended into the list
    "dir_path": "/path/to/kaapana/templates_and_examples/examples/processing-pipelines/otsus-method", // root directory of the extension, doesn't necessarily have to be under kaapana_path
    "kaapana_path": "/path/to/kaapana", // root dir of Kaapana repo
    "kaapana_build_version": "0.0.0-latest", // version of your Kaapana instance, can be found in the bottom bar on the Kaapana platform, such as "kaapana-admin-chart: 0.2.2". If empty or removed, script will assume a platform is running on the machine and will try to fetch it from deployments
    "custom_registry_url": "docker.io/kaapana" // registry url including project Gitlab template: "registry.<gitlab-url>/<group-or-user>/<project>". Keep the default value unless there is a need to include a specific registry in the image tag. If empty or removed, script will assume a platform is running on the machine and will try to fetch it from deployments
    "container_engine": "docker" // docker or podman
}
```

### 3. Build and save images
* Running `extensionctl build image config.json` will save `images.tar` under the speficied `dir_path` in the config file.
* This tar file can then be uploaded inside a Kaapana instance using the [extension upload component](https://kaapana.readthedocs.io/en/latest/user_guide/extensions.html#uploading-extensions-to-the-platform).

### 4. Build and package Helm chart
* `extensionctl build chart config.json` will generate a `<chart-name>.tgz` file under the extension folder specified in `dir_path`.
* Similar to the image tar file, this tgz file can also be uploaded to the platform via drag and drop. After it appears on the extension list, it can be installed via the UI

## FAQ

### `Error searching and replacing in file`
One of the steps is to adapt the python files of the operators where the image is passed to KaapanaBaseOperator. The script will change `{DEFAULT_REGISTRY}` to `custom_registry_url` and `{KAAPANA_BUILD_VERSION}` to `kaapana_build_version`. If `{DEFAULT_REGISTRY}` and `{KAAPANA_BUILD_VERSION}` can not be found in the expected way inside py files, it is assumed that this linking is taken care of by the user. To omit this step of searching/replacing patterns use the flag `--no_overwrite_operators`


## Future work

- perform operations in a /build folder to avoid overwriting original files
- add log levels for verbose output
- change kaapana_build_version to build_version in config yaml. If another templating is added to the dag-installer chart, there is no need that build_version == kaapana_build_version
- add -o for specifying output path
- add support for using a registry url instead of local kaapana_path and fetch the repo
- `--no_prereqs` flag (bool) disables building prereq images, assumes they are already built
- `--overwrite_file_extensions` flag (default: .py)
- `--overwrite_pattern` flag (default: {DEFAULT_REGISTRY},"docker.io/kaapana",{KAAPANA_BUILD_VERSION},"0.0.0-latest")
- `podman` support
- only save last layers if prerequisites are already available (not sure if importing layers in microk8s ctr is possible)
