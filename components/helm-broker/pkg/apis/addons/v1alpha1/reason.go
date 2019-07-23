package v1alpha1

type AddonStatusReason string

const (
	AddonLoadingError                        AddonStatusReason = "LoadingError"
	AddonFetchingError                       AddonStatusReason = "FetchingError"
	AddonConflictInSpecifiedRepositories     AddonStatusReason = "ConflictInSpecifiedRepositories"
	AddonConflictWithAlreadyRegisteredAddons AddonStatusReason = "ConflictWithAlreadyRegisteredAddon"
	AddonUnregisteringError                  AddonStatusReason = "UnregisteringError"
	AddonRegisteringError                    AddonStatusReason = "RegisteringError"
)

func (r AddonStatusReason) String() string {
	return string(r)
}

func (r AddonStatusReason) Message() string {
	switch r {
	case AddonConflictInSpecifiedRepositories:
		return "Specified repositories have addons with the same ID: %v"
	case AddonConflictWithAlreadyRegisteredAddons:
		return "An addon with the same ID is already registered: %v"
	case AddonFetchingError:
		return "Fetching failed due to error: '%v'"
	case AddonLoadingError:
		return "Loading failed due to error: '%v'"
	case AddonRegisteringError:
		return "Registering failed due to error: '%v'"
	case AddonUnregisteringError:
		return "Unregistering failed due to error: '%v'"
	default:
		return ""
	}
}

type RepositoryStatusReason string

const (
	RepositoryURLFetchingError RepositoryStatusReason = "FetchingIndexError"
)

func (r RepositoryStatusReason) String() string {
	return string(r)
}

func (r RepositoryStatusReason) Message() string {
	switch r {
	case RepositoryURLFetchingError:
		return "Fetching repository failed due to error: '%v'"
	default:
		return ""
	}
}
