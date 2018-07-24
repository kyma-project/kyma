package ybundle

func (l *Loader) SetCreateTmpDir(tmpDir func(dir, prefix string) (name string, err error)) {
	l.createTmpDir = tmpDir
}
