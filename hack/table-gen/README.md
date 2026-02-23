# Table Generator: Autogenerate CRD Documentation Tables

## Overview

This package contains a tool that automatically generates a documentation table documenting a CRD, and writes it to specified `.md` files.

## Generate Tables

You can run the table generator in two ways:

- [Use Makefile (Recommended)](#use-makefile-recommended): Configure your CRD once and generate tables with simple `make` commands.
- [Run from command line](#use-command-line): Execute directly with `go run` for quick testing or ad-hoc documentation.

### Use Makefile (Recommended)
Follow the steps:

1. Prepare the parameters' descriptions in the CR's specification file. For example, for the Telemetry CR, prepare the description in [`operator.kyma-project.io_telemetries.yaml`](https://github.com/kyma-project/telemetry-manager/blob/main/helm/charts/default/templates/operator.kyma-project.io_telemetries.yaml).

2. Set up the table generator in the `.md` file in which you want to generate the table. Add the `TABLE-START` and `TABLE-END` tags in the exact place in the document where you want to generate the table.

   ```bash
      <!-- TABLE-START -->
   
      <!-- TABLE-END -->
   ```

3. Add a new target to your module's Makefile with the table generator commands:

   ```makefile
   .PHONY: crd-docs-gen
   crd-docs-gen: $(TABLE_GEN)
      $(TABLE_GEN) --crd-filename ./path/to/your-crd.yaml --md-filename ./docs/user/your-doc.md
      $(TABLE_GEN) --crd-filename ./path/to/your-crd-2.yaml --md-filename ./docs/user/your-doc-2.md
      ...
   ```

   Use the following parameters:
   - **--crd-filename**: Path to the CRD YAML file
   - **--md-filename**: Path to the Markdown file where the table will be inserted
  
   You can create different labels, group them under one, and add your target to the `generate` command. For a complete example, see the [Telemetry module's Makefile](https://github.com/kyma-project/telemetry-manager/blob/main/Makefile#L185). 

4. In the terminal, run the following command from root:

   ```bash
   make generate
   ```
   To verify the result, go to the `.md` files and check that the table has been generated as specified.


### Use Command Line

You can also call the table generator from the command line, without needing to add it to the Makefile. To do this, you can either build it and start it, or use `go run`. See the following example:
   ```bash
   go run main.go --crd-filename ../../installation/resources/crds/telemetry/logpipelines.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-01-logpipeline.md
   ```
