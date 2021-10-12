package syncer

func (r *RelistRequester) WithTimeAfter(fn TimeAfterProvider) *RelistRequester {
	r.timeAfterProvider = fn
	return r
}
