# The tool to generate documentation table out of CRDs

## Overview

This package contains a tool used by kyma developers which purpose is to automatically generate documentation table which documents CRD, and write it to specified MD files. 

## Functionality

This particular tool allows us to:
- Skip some specific element during the table generation by using the tag SKIP-ELEMENT
- Skip specified elements with all its children by using the tag SKIP-WITH-CHILDREN

## Parameters

There are two parameters one should specify in order to use the tool:
- `crd-filename` - full or relative path to the .yaml file containing crd
- `md-filename` - full or relative path to the .md file containing the file where we should insert table rows

## How to use?

Most importantly, you need to go to the MD file you want to generate table in, and put down tags TABLE-START and TABLE-END on a place where you want to insert a table. 

```
<!-- TABLE-START -->

<!-- TABLE-END -->
```

Now, if you want to skip some elements, you can use `SKIP-ELEMENT` or `SKIP-WITH-CHILDREN` tags, brefore start and end tags

```
<!-- SKIP-ELEMENT status.conditions -->
<!-- SKIP-WITH-CHILDREN spec.output -->

<!-- TABLE-START -->

<!-- TABLE-END -->
```

### Call the tool using command line
You can call this tool from command line. In order to do so, you should either build it and start it, or use `go run`.
Here is an example on how to call the tool:
`go run table-gen.go --crd-filename ../../installation/resources/crds/telemetry/logpipelines.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-01-logpipeline.md`

### Call the tool using makefile
- If you update a CRD that is already present in the makefile, you can just call `make generate`.

  If you want to compare only a particular operator or a specific CRD, specify the label you need while calling `make`; for example, `make telemetry-docs`.

  To update the makefile, just introduce a new label for your CRD, and then add it to the `generate`.
  Alternatively, if you want to group your `go run` commands, you can create different labels, group them under the one, and include it to the `generate`, the same way as with `make telemetry-docs`.

## Verifying the result
Go to the .md files and check that the table has been generated as specified.