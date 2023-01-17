If you don't want to provide the whole configuration file and only specify certain parameters, you can copy here your extended `.conf` files.
These files will be injected as a ConfigMap and add/overwrite the default configuration using the `include_dir` directive that allows settings to be loaded from files other than the default `postgresql.conf`.

More info in the [bitnami-docker-postgresql README](https://github.com/bitnami/containers/tree/main/bitnami/postgresql#configuration-file).
