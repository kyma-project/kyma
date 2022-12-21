package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageList(t *testing.T) {
	testList := NewTaggedImageList()

	testImages := map[string]string{
		"image1": "tag1",
		"image2": "tag2",
	}

	t.Run("Add images", func(t *testing.T) {
		for k, v := range testImages {
			testList.AddImageWithTag(k, v)
		}

		t.Log("Add images to the list")
		require.ElementsMatch(t, []string{"image1", "image2"}, testList.ListImages())
	})

	t.Run("Check image with tags", func(t *testing.T) {
		t.Log("check if image:tag exists")
		require.True(t, testList.HasImageWithTag("image1", "tag1"))

		t.Log("fail if image exists but not the tag")
		require.False(t, testList.HasImageWithTag("image2", "tag3"))

		t.Log("fail if checking unrelated image/tag")
		require.False(t, testList.HasImageWithTag("image1", "tag2"))

		t.Log("fail if image doesn't exists")
		require.False(t, testList.HasImageWithTag("image3", "tag2"))

	})
}
