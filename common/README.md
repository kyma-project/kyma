# Kyma common package

A set of Go libraries implementing common use cases encountered throughout Kyma components.

## Purpose

The Kyma project was consolidated and now all core projects written in Go are placed in one repository. Unfortunately, all of them are still written in a different way. This makes testing existing code as well as making new contributions a lot harder as you need to approach each and every component in its own special way.

This package aims to address this problem by providing a set of libraries that developers should prefer over introducing new solution to already solved problem.

## Criteria for adding code here

In order to be put here a new library must meet those criteria:

- It is used by multiple components in Kyma. Do not generalize when it is not necessary as it puts additional overhead in maintenance.
- It must be well covered with unit tests. Many components may depend on it so avoiding bugs here is even more crucial then anywhere else.
- It must have open source grade documentation. Many developers will be using the code so it must be easy to understand what it is doing without reaching author.

## Testing

To make this package even more bugproof tests must be executed with following flags:
- `-count 100` - to make sure they are stable
- `-race` - to make sure the code is safe to use in multi-threaded environment