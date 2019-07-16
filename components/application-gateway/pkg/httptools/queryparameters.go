package httptools

import "net/url"

func SetQueryParameters(url *url.URL, queryParameters *map[string][]string) {
	if queryParameters == nil {
		return
	}

	reqQueryValues := url.Query()

	for customQueryParam, values := range *queryParameters {
		if _, ok := reqQueryValues[customQueryParam]; ok {
			continue
		}

		reqQueryValues[customQueryParam] = values
	}

	url.RawQuery = reqQueryValues.Encode()
}
