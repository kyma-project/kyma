# Table Generator: Autogenerate CRD Documentation Tables

## Overview

This package contains a tool that automatically generates a documentation table documenting a CRD, and writes it to specified `.md` files.

## Generate Tables

You can run the table generator in two ways:

- [Add to Makefile](#use-makefile-recommended) (recommended for regular updates): Configure your CRD once and generate tables with simple `make` commands
- [Run from command line](#use-command-line) (for one-time generation): Execute directly with `go run` for quick testing or ad-hoc documentation

### Use Makefile (Recommended)
To update the makefile, just introduce a new label for your CRD, and then add it to the `generate`. Alternatively, if you want to group your `go run` commands, you can create different labels, group them under the one, and include it to the `generate`, the same way as with `make telemetry-docs`

1. Prepare the parameters' descriptions in the CR's specification file. For example, for the Telemetry CR, prepare the description in [`operator.kyma-project.io_telemetries.yaml`](https://github.com/kyma-project/telemetry-manager/blob/main/helm/charts/default/templates/operator.kyma-project.io_telemetries.yaml).

2. Add the following mappings to the module's Makefile:

- `--crd-filename` - full or relative path to the `.yaml` file with the CRD
- `--md-filename` - full or relative path to the `.md` file in which you want to generate the table

   For example, see the [Telemetry module's Makefile](https://github.com/kyma-project/telemetry-manager/blob/main/Makefile#L185).

3. Set up the table generator in the `.md` file in which you want to generate the table. Add the `TABLE-START` and `TABLE-END` tags in the exact place in the document where you want to generate the table.

   ```bash
      <!-- TABLE-START -->
   
      <!-- TABLE-END -->
   ```

4. In the terminal, run the following command from root:

   ```bash
   make generate
   ```
   To verify the result, go to the `.md` files and check that the table has been generated as specified.


### Use Command Line

You can also call the table generator from the command line, without needing to add it to the Makefile. to do this, you can either build it and start it, or use `go run`. See the following example:
   ```bash
   go run main.go --crd-filename ../../installation/resources/crds/telemetry/logpipelines.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-01-logpipeline.md
   ```
