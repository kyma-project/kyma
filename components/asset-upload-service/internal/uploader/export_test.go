package uploader

func (u *Uploader) PopulateErrors(errorsCh chan error) []error {
	return u.populateErrors(errorsCh)
}

func (u *Uploader) PopulateResults(resultsCh chan *UploadResult) []UploadResult {
	return u.populateResults(resultsCh)
}
