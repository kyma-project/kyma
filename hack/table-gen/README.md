# The Table Generator: Generate Documentation Tables Automatically from CRDs

## Overview

This package contains a tool that automatically generates a documentation table documenting a CRD, and writes it to specified `.md` files.

## Parameters

You must specify the following parameters:

- `crd-filename` - full or relative path to the `.yaml` file containing the CRD
- `md-filename` - full or relative path to the `.md` file in which to insert the table rows

## Set Up the Table Generator

Open the `.md` file you want to generate table in, and in the place where you want to insert a table, enter the tags `TABLE-START` and `TABLE-END`.

```
   <!-- TABLE-START -->

   <!-- TABLE-END -->
```

### Call the Table Generator

You can call the table generator either from the command line, or with the makefile:

- If you want to call the table generator from the command line, you can either build it and start it, or use `go run`. See the following example:
  `go run main.go --crd-filename ../../installation/resources/crds/telemetry/logpipelines.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-01-logpipeline.md`

- If you update a CRD that is already present in the makefile, you can just call `make generate`.

  If you want to compare only a particular operator or a specific CRD, specify the label you need while calling `make`; for example, `make telemetry-docs`.

  To update the makefile, just introduce a new label for your CRD, and then add it to the `generate`.
  Alternatively, if you want to group your `go run` commands, you can create different labels, group them under the one, and include it to the `generate`, the same way as with `make telemetry-docs`.

## Verifying the Result

Go to the `.md` files and check that the table has been generated as specified.
