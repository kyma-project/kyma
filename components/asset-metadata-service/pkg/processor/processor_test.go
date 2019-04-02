package processor_test

import (
	"context"
	"errors"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/extractor/automock"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/processor"
	"testing"
	"time"

	fautomock "github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader/automock"
	"github.com/onsi/gomega"
)

func TestProcessor_Do(t *testing.T) {
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

		files := []processor.Job{
			{
				FilePath: "test/test1.yaml",
				File:     mock1,
			},
			{
				FilePath: "test/test2.yaml",
				File:     mock2,
			},
		}

		expectedResult := []processor.ResultSuccess{
			{
				FilePath: "test/test1.yaml",
				Output: map[string]interface{}{
					"foo": "bar",
					"bar": 3,
				},
			},
			{
				FilePath: "test/test2.yaml",
				Output: map[string]interface{}{
					"foo": 32,
					"bar": "test.example.com",
				},
			},
		}

		timeout, err := time.ParseDuration("10h")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		jobCh, jobCount := fixJobCh(files)

		extractorMock := new(automock.Extractor)
		extractorMock.On("ReadMetadata", mock1).Return(map[string]interface{}{
			"foo": "bar",
			"bar": 3,
		}, nil).Once()
		extractorMock.On("ReadMetadata", mock2).Return(map[string]interface{}{
			"foo": 32,
			"bar": "test.example.com",
		}, nil).Once()
		defer extractorMock.AssertExpectations(t)

		e := processor.New(func(job processor.Job) (interface{}, error) {
			return extractorMock.ReadMetadata(job.File)
		}, 5, timeout)

		// When
		res, errs := e.Do(context.TODO(), jobCh, jobCount)

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

		files := []processor.Job{
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

		extractorMock := new(automock.Extractor)
		extractorMock.On("ReadMetadata", mock1).Return(nil, testErr).Once()
		extractorMock.On("ReadMetadata", mock2).Return(nil, testErr).Once()
		defer extractorMock.AssertExpectations(t)

		e := processor.New(func(job processor.Job) (interface{}, error) {
			return extractorMock.ReadMetadata(job.File)
		}, 5, timeout)

		// When
		_, errs := e.Do(context.TODO(), jobCh, jobCount)

		// Then
		g.Expect(errs).To(gomega.HaveLen(2))

		for _, err := range errs {
			g.Expect(err.Error.Error()).To(gomega.ContainSubstring("Test error"))
		}
	})
}

func TestProcessor_PopulateErrors(t *testing.T) {
	t.Run("Errors", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		elem1 := processor.ResultError{
			Error: errors.New("Test 1"),
		}
		elem2 := processor.ResultError{
			FilePath: "test/path.js",
			Error:    errors.New("Test 2"),
		}

		errCh := make(chan *processor.ResultError, 3)
		errCh <- &elem1
		errCh <- &elem2
		errCh <- nil
		close(errCh)

		e := processor.Processor{}

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

		errCh := make(chan *processor.ResultError)
		close(errCh)

		e := processor.Processor{}

		// When
		errs := e.PopulateErrors(errCh)

		// Then
		g.Expect(errs).To(gomega.BeEmpty())
	})
}

func TestProcessor_PopulateResults(t *testing.T) {
	t.Run("Results", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		res1 := processor.ResultSuccess{
			FilePath: "test.yaml",
		}
		res2 := processor.ResultSuccess{
			FilePath: "test2.yaml",
		}

		resultsCh := make(chan *processor.ResultSuccess, 3)
		resultsCh <- &res1
		resultsCh <- &res2
		resultsCh <- nil
		close(resultsCh)

		e := processor.Processor{}

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

		resultsCh := make(chan *processor.ResultSuccess, 3)
		close(resultsCh)

		e := processor.Processor{}

		// When
		res := e.PopulateResults(resultsCh)

		// Then
		g.Expect(res).To(gomega.BeEmpty())
	})

}

func fixJobCh(jobs []processor.Job) (chan processor.Job, int) {
	jobLen := len(jobs)

	jobsChannel := make(chan processor.Job, jobLen)
	for _, job := range jobs {
		jobsChannel <- job
	}
	close(jobsChannel)

	return jobsChannel, jobLen
}
