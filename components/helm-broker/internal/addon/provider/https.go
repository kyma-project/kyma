package provider

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	getter "github.com/hashicorp/go-getter"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"k8s.io/apimachinery/pkg/util/rand"
)

// defaultClient setup as global variable to allow sharing the same client across different HTTPGetter instances
// same as in core library http.DefaultClient.
var defaultClient = cleanhttp.DefaultClient()

// HTTPGetter provides functionality for loading addon from any HTTP/HTTPS repository serving static files.
type HTTPGetter struct {
	underlying *getter.HttpGetter

	dst     string
	idxURL  *url.URL
	repoURL *url.URL
}

// NewHTTP returns new instance of HTTPGetter
func NewHTTP(idxAddr, dst string) (RepositoryGetter, error) {
	repoAddr := baseOfURL(idxAddr)

	repoURL, err := url.Parse(repoAddr)
	if err != nil {
		return nil, err
	}

	idxURL, err := url.Parse(idxAddr)
	if err != nil {
		return nil, err
	}

	return &HTTPGetter{
		idxURL:  idxURL,
		repoURL: repoURL,
		dst:     path.Join(dst, rand.String(10)),
		underlying: &getter.HttpGetter{
			Client: defaultClient,
		},
	}, nil
}

// Cleanup removes directory where content was downloaded
func (h *HTTPGetter) Cleanup() error {
	return os.RemoveAll(h.dst)
}

// IndexReader returns index reader
func (h *HTTPGetter) IndexReader() (io.ReadCloser, error) {
	savePath := path.Join(h.dst, "index")
	if err := h.underlying.GetFile(savePath, h.idxURL); err != nil {
		return nil, err
	}

	return os.Open(savePath)
}

// AddonLoadInfo returns information how to load addon
func (h *HTTPGetter) AddonLoadInfo(name addon.Name, version addon.Version) (LoadType, string, error) {
	rawURL := h.AddonDocURL(name, version)
	u, err := url.Parse(rawURL)
	if err != nil {
		return UnknownLoadType, "", err
	}

	savePath := path.Join(h.dst, rand.String(10))
	if err := h.underlying.GetFile(savePath, u); err != nil {
		return UnknownLoadType, "", err
	}

	return ArchiveLoadType, savePath, nil
}

// AddonDocURL returns url for addon documentation
func (h *HTTPGetter) AddonDocURL(name addon.Name, version addon.Version) string {
	return fmt.Sprintf("%s%s-%s.tgz", h.repoURL, name, version)
}

func baseOfURL(fullURL string) string {
	return strings.TrimRight(fullURL, path.Base(fullURL))
}
