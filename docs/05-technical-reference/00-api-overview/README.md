# API Overview

Kyma API versioning and deprecation closely follows the Kubernetes API versioning and deprecation policy. Hence, this document is based on the Kubernetes official documentation.

## API Versioning 

(Derived from [API Overview](https://kubernetes.io/docs/reference/using-api/#api-versioning))

Different API versions indicate different levels of stability and support.

Here's a summary of each level:

- Alpha:
  - The version names contain `alpha`, for example, `v1alpha1`.
  - The software may contain bugs. Enabling a feature may expose bugs. A feature may be disabled by default.
  - The support for a feature may be dropped at any time without notice.
  - The API may change in incompatible ways in a later software release without notice.
  - The software is recommended for use only in short-lived testing clusters, due to increased risk of bugs and lack of long-term support.

- Beta:
  - The version names contain `beta`, for example, `v2beta3`.
  - The software is well tested. Enabling a feature is considered safe. Features are enabled by default.
  - The support for a feature will not be dropped, though the details may change.
  - The schema and/or semantics of objects may change in incompatible ways in a subsequent beta or stable release. When this happens, migration instructions are provided. Schema changes may require deleting, editing, and re-creating API objects. The editing process may not be straightforward. The migration may require downtime for applications that rely on the feature.
  - The software is not recommended for production uses. Subsequent releases may introduce incompatible changes. If you have multiple clusters which can be upgraded independently, you may be able to relax this restriction.

 > [!NOTE]
 > Try beta features and provide feedback. After the features exit beta, it may not be practical to make more changes.

- Stable:
  - The version name is `vX`, where `X` is an integer.
  - The stable versions of features appear in released software for many subsequent versions.
  
## Deprecating Parts of the API

(Derived from [Kubernetes Deprecation Policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/#deprecating-parts-of-the-api))

API versions fall into 3 main tracks. Each of the tracks has different policies for deprecation:

| Example  | Track                            |
|----------|----------------------------------|
| v1       | GA (generally available, stable) |
| v1beta1  | Beta (pre-release)               |
| v1alpha1 | Alpha (experimental)             |

The following rules govern the deprecation of elements of the API which 
include:

   * REST resources (also known as API objects)
   * Fields of REST resources
   * Annotations on REST resources, including `beta` annotations, but not including `alpha` annotations
   * Enumerated or constant values
   * Component config structures

These rules are enforced between official releases, not between arbitrary commits to main or release branches.

**Rule #1: API elements may only be removed by incrementing the version of the API group.**

Once an API element has been added to an API group at a particular version, it cannot be removed from that version or have its behavior significantly changed, regardless of the track.

**Rule #2: API objects must be able to round-trip between API versions in a given release without information loss, with the exception of whole REST resources that do not exist in some versions.**

For example, an object can be written as v1 and then read back as v2 and converted to v1, and the resulting v1 resource is identical to the original.  The representation in v2 might be different from v1, but the system knows how to convert between them in both directions.  Additionally, any new field added in v2 must be able to round-trip to v1 and back, which means v1 might have to add an equivalent field or represent it as an annotation.

**Rule #3: An API version in a given track may not be deprecated in favor of a less stable API version.**

  * GA API versions can replace beta and alpha API versions.
  * Beta API versions can replace earlier beta and alpha API versions, but *may not* replace GA API versions.
  * Alpha API versions can replace earlier alpha API versions, but *may not* replace GA or beta API versions.

**Rule #4a: minimum API lifetime is determined by the API stability level.**

   * **GA API versions may be marked as deprecated, but must not be removed within a major version of a Kyma module**
   * **Beta API versions must be supported for 6 months or 3 releases (whichever is longer) after deprecation**
   * **Alpha API versions may be removed in any release without prior deprecation notice**

**Rule #4b: The `preferred` API version and the `storage version` for a given group may not advance until after a release has been made that supports both the new version and the previous version.**

Users must be able to upgrade to a new release of a Kyma module and then roll back to a previous release, without converting anything to the new API version or suffering breakages unless they explicitly choose to use features only available in the newer version. This is particularly evident in the stored representation of objects.

### REST Resources (aka API Objects)

Consider a hypothetical REST resource named Widget, which was present in API v1 in the above timeline, and which needs to be deprecated. The deprecation is documented and announced in sync with release X+1. The Widget resource still exists in API version v1 (deprecated) but not in v2alpha1. The Widget resource continues to exist and function in releases up to and including X+5. The Widget resource ceases to exist, and the behavior gets removed in release X+6, when API v1 has aged out.  

### Fields of REST Resources

As with whole REST resources, an individual field which was present in API v1 must exist and function until API v1 is removed.  Unlike whole resources, the v2 APIs may choose a different representation for the field, as long as it can be round-tripped. For example, a v1 field named `magnitude` which was deprecated might be named `deprecatedMagnitude` in API v2. When v1 is eventually removed, the deprecated field can be removed from v2.

### Enumerated or Constant Values

As with whole REST resources and their fields, a constant value which was supported in API v1 must exist and function until API v1 is removed.

### Component Config Structures

Component configs are versioned and managed similarly to REST resources.
