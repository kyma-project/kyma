# Kyma Common Package

This package is a set of Go libraries that implement common use cases found in different Kyma components.

## Purpose

With the consolidated structure of the Kyma project, all core components written in Go are stored in a single repository. Although all components have the same programming language as a cornerstone, they are written differently. This makes testing the existing code and making new contributions difficult as each component must be approached in a different way.

This package aims to address this problem by providing a set of libraries that the developers should use rather than  introduce new solutions to an already solved problem.

## How to Add Code to This Package

To add a new library to this package, it must meet these criteria:

- It is used by multiple components in Kyma. Do not generalize when it is not necessary as it puts additional overhead in maintenance.
- It must be well covered with unit tests. Many components may end up depending on the library you add.
- It must have open source-grade documentation. The library will be widely used and must be easy to use without the need to contact the author.

## Naming Conventions

If the common package is a Kubernetes client for a custom resource, add a `-client` suffix to the folder name.

## Testing

To ensure that the common package has no bugs, the tests must be executed with these flags:

- `-count 100` - ensures the stability of tests
- `-race` - ensures that the code is safe to use in a multi-threaded environment
