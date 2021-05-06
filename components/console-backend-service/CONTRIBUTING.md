# Overview
To contribute to this project, follow the rules from the general [CONTRIBUTING.md](https://github.com/kyma-project/community/blob/main/CONTRIBUTING.md) document in the `community` repository.
For additional, project-specific guidelines, see the respective sections of this document.

## Contribution rules
Before you make a pull request, review the following rules.

> **NOTE:** These rules mention terms described in the [Terminology](./docs/terminology.md) document.

### Project structure
- Place all GraphQL types in the `gqlschema` package. Generate them with the `gqlgen.sh` script. If you need any customization, move them to separate files, modify, and include them in the `config.yml` file.
- Keep the first level of a domain package consistent. Create these files for every resource:
    - `{NAME}_resolver.go`, which contains resolver type. They usually call services and convert types.
    - `{NAME}_service.go`, which contains the business logic. The service uses data transfer object (DTO) type.
    - `{NAME}_converter.go`, which is used for DTO to GraphQL type conversion. The type conversion for basic types, for example `string`, can be performed in a resolver.
- You can create subpackages and define their custom structure.
- Put all kind of generic utilities in the `pkg` directory. For example, see the [Gateway](./internal/domain/application) utility in the `application` domain.
- Place utilities tied to a specific domain in domain subpackages. For example, see the [Resource](./pkg/resource) utility.
- Place cross-domain utils in the `internal` directory. For example, see the [Pager](./internal/pager) utility.
- The domain resolver must be composed of resource resolvers defined in that package.
- Every domain must have the main file with the same name as the name of the domain package. For example, there must be the `servicecatalog.go` file in `servicecatalog` domain package. This is the root of the domain which should expose the `Resolver` type. Optionally, it can expose the `Config` type for passing the configuration values. For cross-domain implementations, it can contain the `Container` type which must contain the `Resolver` field and other fields needed by other packages.
- Place interfaces, which are shared between files in single domain in `interfaces.go`.

### Implementation guidelines
Follow these rules while you develop new features for this project.

**General implementation rules:**
- Every domain resolver which is not required to run Console Backend Service should be pluggable. It means that it should implement the `PluggableModule` interface from the [`module`](./internal/module) package. To see an example implementation, review the [Service Catalog](./internal/domain/servicecatalog) module.
- Avoid creating cross-domain packages. Do not create domain-to-domain, direct dependencies. Use interfaces defined in the [`shared`](./internal/domain/shared) package to avoid circular dependencies.
- Do not make direct dependencies between a resolver type and services. Define interfaces, which contain only used methods. Use these interface types in a constructor of a resolver type.
- Do not export domain's interfaces and types which are not used in other places. For testing purposes, use the `export_test.go` file, which exports constructors only for tests.
- If an error appears, the resolver must return a general error message, which hides the applied solutions and logic behind them. Log the details of the error using the `glog` logger.
- Avoid creating Functions in domains as they are accessible in whole domain. Create types and define their methods.
- Return pointers for objects that represents resources in services and converters. Pass objects by pointer as method arguments in converters.
- If a specific resource does not exist during `find` operation, return `nil` without an error.
- Use cache whenever possible for small pieces of data. Monitor resources usage and consider invalidating cache after some period of inactivity that lasts, for example, one day.

**GraphQL:**
- For queries and mutations that have more than three arguments, use [input types](http://graphql.org/learn/schema/#input-types).
- Define the mutated object as a result of the mutation.
- For a query that returns a collection of objects, always return an empty array instead of `nil`. Mark all array elements as non-nullable. For example, define a query in the GraphQL schema that returns an array of service instances as `serviceInstances: [ServiceInstance!]!`.

**Kubernetes resources:**
- For read only operations, use SharedIndexInformers, a client-side caching mechanism, which synchronize with Kubernetes API. Use them to find and list resources. SharedIndexInformers have different API from IndexInformers, but it is possible to attach multiple event handlers to them to facilitate future modifications.
- To categorize items in the cache store, add indexers for SharedIndexInformers in services.
- Use Kubernetes Go client for `create`, `update`, and `delete` operations. Do not operate on cache.

**Acceptance tests:**
- Query all possible fields during testing queries and mutations.
- To check if nested objects are correctly resolved, perform a minimum validation and check the required fields, such as the name.

### Naming guidelines
Use these patterns for naming GraphQL operations:
- Use the imperative mood to name mutations. For example, name a mutation that creates a new service instance `createServiceInstance`.
- Name queries with singular or plural nouns. For example, name the query that returns a single service instance `serviceInstance`. Name the query that returns all service instances `serviceInstances`.

Use these patterns for naming types in the first level of a domain package:
- `Config` for an exported type, which stores configuration values
- `Resolver` for an exported type, which is composed of resources resolvers in the domain package
- `Container` for an exported type, which exports `Resolver` and other types required by other domains
- `{RESOURCE_NAME}Resolver` for a resolver type of a specific resource
- `{RESOURCE_NAME}Service` for a service type of a specific resource
- `{RESOURCE_NAME}Converter` for a converter type of a specific resource

Use these patterns for naming methods in resource resolvers:
- `{RESOURCE_NAME}Query` for a query resolver
- `{RESOURCE_NAME}Mutation` for a mutation resolver
- `{RESOURCE_NAME}Subscription` for a subscription resolver

Use these patterns for naming methods in services:
- `Create` to create a single resource
- `Find` to get a single resource
- `List` to get multiple resources
- `Update` to update a resource
- `Delete` to delete a resource
- For specific operations, use short, meaningful names. For example, a method of `instanceService` that lists instances for a specific class should be named `ListForClass`.

Use this pattern for naming methods in converters:
- `ToGQL` for DTO to GraphQL type conversion
- For the conversion in the opposite direction, use a similar naming convention. For example, `ToK8S`.

### Code quality
- All Go code must have unit and acceptance tests for all business logic.
- All Go code must pass `go vet ./...`. The CI build job performs the check automatically.
- Format the Go code with `gofmt`.
- Describe any new application configuration options in the [README.md](./README.md) document.
