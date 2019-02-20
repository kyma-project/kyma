package uploader

func (u *Uploader) PopulateErrors(errorsCh chan *UploadError) []UploadError {
	return u.populateErrors(errorsCh)
}

func (u *Uploader) PopulateResults(resultsCh chan *UploadResult) []UploadResult {
	return u.populateResults(resultsCh)
}
