package extractor_test

import (
	"context"
	"errors"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/extractor"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/matador/automock"
	"testing"
	"time"

	fautomock "github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader/automock"
	"github.com/onsi/gomega"
)

func TestExtractor_Process(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		file := &fautomock.File{}
		file.On("Close").Return(nil)

		mock1 := &fautomock.FileHeader{}
		mock1.On("Filename").Return("test1.yaml")
		mock1.On("Size").Return(int64(3213)).Once()
		mock1.On("Open").Return(file, nil).Once()

		mock2 := &fautomock.FileHeader{}
		mock2.On("Filename").Return("test2.yaml")
		mock2.On("Size").Return(int64(213)).Once()
		mock2.On("Open").Return(file, nil).Once()

		files := []extractor.Job{
			{
				FilePath: "test/test1.yaml",
				File:     mock1,
			},
			{
				FilePath: "test/test2.yaml",
				File:     mock2,
			},
		}

		expectedResult := []extractor.ResultSuccess{
			{
				FilePath: "test/test1.yaml",
				Metadata: map[string]interface{}{
					"foo": "bar",
					"bar": 3,
				},
			},
			{
				FilePath: "test/test2.yaml",
				Metadata: map[string]interface{}{
					"foo": 32,
					"bar": "test.example.com",
				},
			},
		}

		timeout, err := time.ParseDuration("10h")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		jobCh, jobCount := fixJobCh(files)

		matadorMock := new(automock.Matador)
		matadorMock.On("ReadMetadata", mock1).Return(map[string]interface{}{
			"foo": "bar",
			"bar": 3,
		}, nil).Once()
		matadorMock.On("ReadMetadata", mock2).Return(map[string]interface{}{
			"foo": 32,
			"bar": "test.example.com",
		}, nil).Once()
		defer matadorMock.AssertExpectations(t)


		e := extractor.New(5, timeout)
		e.SetMatador(matadorMock)

		// When
		res, errs := e.Process(context.TODO(),jobCh, jobCount)

		// Then
		g.Expect(errs).To(gomega.BeEmpty())
		for _, r := range expectedResult {
			g.Expect(res).To(gomega.ContainElement(r))
		}
	})

	t.Run("ResultError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testErr := errors.New("Test error")

		file := &fautomock.File{}
		file.On("Close").Return(nil)

		mock1 := &fautomock.FileHeader{}
		mock1.On("Filename").Return("test1.yaml")
		mock1.On("Size").Return(int64(3213)).Once()
		mock1.On("Open").Return(file, nil).Once()

		mock2 := &fautomock.FileHeader{}
		mock2.On("Filename").Return("test2.yaml")
		mock2.On("Size").Return(int64(213)).Once()
		mock2.On("Open").Return(file, nil).Once()

		files := []extractor.Job{
			{
				FilePath: "test/test1.yaml",
				File:     mock1,
			},
			{
				FilePath: "test/test2.yaml",
				File:     mock2,
			},
		}

		timeout, err := time.ParseDuration("10h")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		jobCh, jobCount := fixJobCh(files)

		matadorMock := new(automock.Matador)
		matadorMock.On("ReadMetadata", mock1).Return(nil, testErr).Once()
		matadorMock.On("ReadMetadata", mock2).Return(nil, testErr).Once()
		defer matadorMock.AssertExpectations(t)

		e := extractor.New(5, timeout)
		e.SetMatador(matadorMock)

		// When
		_, errs := e.Process(context.TODO(),jobCh, jobCount)

		// Then
		g.Expect(errs).To(gomega.HaveLen(2))

		for _, err := range errs {
			g.Expect(err.Error.Error()).To(gomega.ContainSubstring("Test error"))
		}
	})
}

func TestExtractor_PopulateErrors(t *testing.T) {
	t.Run("Errors", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		elem1 := extractor.ResultError{
			Error: errors.New("Test 1"),
		}
		elem2 := extractor.ResultError{
			FilePath: "test/path.js",
			Error:    errors.New("Test 2"),
		}

		errCh := make(chan *extractor.ResultError, 3)
		errCh <- &elem1
		errCh <- &elem2
		errCh <- nil
		close(errCh)

		e := extractor.Extractor{}

		// When
		errs := e.PopulateErrors(errCh)

		// Then
		g.Expect(errs).To(gomega.HaveLen(2))
		g.Expect(errs).To(gomega.ContainElement(elem1))
		g.Expect(errs).To(gomega.ContainElement(elem2))
	})

	t.Run("No Errors", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		errCh := make(chan *extractor.ResultError)
		close(errCh)

		e := extractor.Extractor{}

		// When
		errs := e.PopulateErrors(errCh)

		// Then
		g.Expect(errs).To(gomega.BeEmpty())
	})
}

func TestExtractor_PopulateResults(t *testing.T) {
	t.Run("Results", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		res1 := extractor.ResultSuccess{
			FilePath: "test.yaml",
		}
		res2 := extractor.ResultSuccess{
			FilePath: "test2.yaml",
		}

		resultsCh := make(chan *extractor.ResultSuccess, 3)
		resultsCh <- &res1
		resultsCh <- &res2
		resultsCh <- nil
		close(resultsCh)

		e := extractor.Extractor{}

		// When
		res := e.PopulateResults(resultsCh)

		// Then
		g.Expect(res).To(gomega.HaveLen(2))
		g.Expect(res).To(gomega.ContainElement(res1))
		g.Expect(res).To(gomega.ContainElement(res2))
	})

	t.Run("No Results", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		resultsCh := make(chan *extractor.ResultSuccess, 3)
		close(resultsCh)

		e := extractor.Extractor{}

		// When
		res := e.PopulateResults(resultsCh)

		// Then
		g.Expect(res).To(gomega.BeEmpty())
	})

}

func fixJobCh(jobs []extractor.Job) (chan extractor.Job, int) {
	jobLen := len(jobs)

	jobsChannel := make(chan extractor.Job, jobLen)
	for _, job := range jobs {
		jobsChannel <- job
	}
	close(jobsChannel)

	return jobsChannel, jobLen
}
