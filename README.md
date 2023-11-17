# kubectl-catalogd
`kubectl-catalogd` is a kubectl plugin (can also be used as a standalone binary if desired) that enables the listing, inspecting, and searching of the File-Based Catalog contents of catalogd's `Catalog` resources.

>[!NOTE]
>This is a kubectl plugin based on an alpha project.
>Due to this, this plugin will remain in an alpha state and be prone to 
>breaking changes due to changes made in the catalogd project. 
>It is recommended that you only use a version of this plugin that is compatible with the
>version of catalogd installed on your cluster. To see what version is compatible with an installed
>version of the plugin, run `kubectl catalogd version`. The output will contain the current version
>of the kubectl-catalogd plugin _and_ the version of catalogd it was built against.

## Subcommands
These examples assume a running Kubernetes cluster with catalogd installed and an unpacked `Catalog` resource.
These examples use a minimal catalog to keep the output brief and easier to read. The catalog used can be found under `test/testdata/`.

### `list`

```sh
$ kubectl catalogd list -h
Lists catalog objects

Usage:
  catalogd list [flags]

Flags:
      --catalog string   specify the catalog that should be used. By default it will fetch from all catalogs
  -h, --help             help for list
      --name string      specify the FBC object name that should be used to filter the resulting output
      --package string   specify the FBC object package that should be used to filter the resulting output
      --schema string    specify the FBC object schema that should be used to filter the resulting output
```

**Example**: _List all catalog contents_
```sh
$ kubectl catalogd list
 test-catalog  olm.package  prometheus
 test-catalog  olm.channel prometheus alpha
 test-catalog  olm.channel prometheus beta
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.0
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.1
 test-catalog  olm.bundle prometheus prometheus-operator.1.2.0
 test-catalog  olm.bundle prometheus prometheus-operator.2.0.0
 test-catalog  olm.package  plain
 test-catalog  olm.channel plain beta
 test-catalog  olm.bundle plain plain.0.1.0

```

>[!NOTE]
>The output of the `list` subcommand is stylized for easier visual separation of each "field" in the output but these styles don't show up in the examples due to it using the markdown code block syntax.

**Example**: _List all catalog contents with schema of `olm.package`_
```sh
$ kubectl catalogd list --schema olm.package
 test-catalog  olm.package  prometheus
 test-catalog  olm.package  plain

```

**Example**: _List all catalog contents with schema of `olm.bundle` that belong to package `plain`_
```sh
$ kubectl catalogd list --schema olm.bundle --package plain
 test-catalog  olm.bundle plain plain.0.1.0
```

### `search`

```sh
$ kubectl catalogd search -h
Searches catalog objects

Usage:
  catalogd search [input] [flags]

Flags:
      --catalog string   specify the catalog that should be used. By default it will fetch from all catalogs
  -h, --help             help for search
      --package string   specify the FBC object package that should be used to filter the resulting output
      --schema string    specify the FBC object schema that should be used to filter the resulting output
```

**Example**: _Search for catalog contents that contain `prom` in the name_
```sh
$ kubectl catalogd search prom
 test-catalog  olm.package  prometheus
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.0
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.1
 test-catalog  olm.bundle prometheus prometheus-operator.1.2.0
 test-catalog  olm.bundle prometheus prometheus-operator.2.0.0

```

>[!NOTE]
>The output of the `search` subcommand is stylized for easier visual separation of each "field" in the output but these styles don't show up in the examples due to it using the markdown code block syntax.

**Example**: _Search for catalog contents that contain `prom` in the name and have schema of `olm.package`_
```sh
$ kubectl catalogd search prom --schema olm.package
 test-catalog  olm.package  prometheus

```

**Example**: _Search for catalog contents that contain the `prom` in the name and belong to the `prometheus` package_
```sh
$ kubectl-catalogd search prom --package prometheus
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.0
 test-catalog  olm.bundle prometheus prometheus-operator.1.0.1
 test-catalog  olm.bundle prometheus prometheus-operator.1.2.0
 test-catalog  olm.bundle prometheus prometheus-operator.2.0.0
```

### `inspect`

```sh
$ kubectl catalogd inspect -h
Inspects catalog objects

Usage:
  catalogd inspect [schema] [name] [flags]

Flags:
      --catalog string   specify the catalog that should be used. By default it will fetch from all catalogs and use the first match
  -h, --help             help for inspect
      --output string    specify the output format. Valid values are 'json' and 'yaml' (default "json")
      --package string   specify the FBC object package that should be used to filter the resulting output
      --style string     specify the style to use for syntax highlighting. If this value is empty syntax highlighting is disabled.
```

**Example**: _Inspect the `olm.channel` object with a name of `beta` that belongs to the `plain` package_
```sh
$ kubectl catalogd inspect olm.channel beta --package plain
{
  "entries": [
    {
      "name": "plain.0.1.0"
    }
  ],
  "name": "beta",
  "package": "plain",
  "schema": "olm.channel"
}
```

with YAML as the output:
```sh
$ kubectl catalogd inspect olm.channel beta --package plain --output yaml
entries:
- name: plain.0.1.0
name: beta
package: plain
schema: olm.channel
```

>[!NOTE]
>By default styling/syntax highlighting on the output is disabled so that the output can be piped to tools
>like `jq` and `yq` that expect plain text. If you do want syntax highlighted output, the style can be
>specified using the `--style` flag. The set of available styles can be found at https://github.com/alecthomas/chroma/tree/master/styles