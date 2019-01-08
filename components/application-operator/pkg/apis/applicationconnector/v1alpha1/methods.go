package v1alpha1

func (app *Application) SetInstallationStatus(status InstallationStatus) {
	app.Status.InstallationStatus = status
}

func (app *Application) SetAccessLabel() {
	app.Spec.AccessLabel = app.Name
}

func (app *Application) AddFinalizer(finalizer string) {
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
