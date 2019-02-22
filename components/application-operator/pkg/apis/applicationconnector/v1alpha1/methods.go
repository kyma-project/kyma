package v1alpha1

import "strings"

func (app *Application) SetInstallationStatus(status InstallationStatus) {
	app.Status.InstallationStatus = status
}

func (app *Application) SetAccessLabel() {
	app.Spec.AccessLabel = app.Name
}

func (app *Application) SetFinalizer(finalizer string) {
	if !app.HasFinalizer(finalizer) {
		app.addFinalizer(finalizer)
	}
}

func (app *Application) addFinalizer(finalizer string) {
	app.Finalizers = append(app.Finalizers, finalizer)
}

func (app *Application) HasFinalizer(finalizer string) bool {
	return app.finalizerIndex(finalizer) != -1
}

func (app *Application) RemoveFinalizer(finalizer string) {
	finalizerIndex := app.finalizerIndex(finalizer)
	if finalizerIndex == -1 {
		return
	}

	app.Finalizers = append(app.Finalizers[:finalizerIndex], app.Finalizers[finalizerIndex+1:]...)
}

func (app *Application) finalizerIndex(finalizer string) int {
	for i, e := range app.Finalizers {
		if e == finalizer {
			return i
		}
	}

	return -1
}

// HasTenant returns true if ApplicationSpec has a non-empty value for Tenant field set
func (appSpec ApplicationSpec) HasTenant() bool {
	return strings.TrimSpace(appSpec.Tenant) != ""
}

// HasGroup returns true if ApplicationSpec has a non-empty value for Group field set
func (appSpec ApplicationSpec) HasGroup() bool {
	return strings.TrimSpace(appSpec.Group) != ""
}
