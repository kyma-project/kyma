package certificates

import "regexp"

func GetOrganization(subject string) string {
	return getRegexMatch("O=([^,]+)", subject)
}

func GetOrganizationalUnit(subject string) string {
	return getRegexMatch("OU=([^,]+)", subject)
}

func GetCountry(subject string) string {
	return getRegexMatch("C=([^,]+)", subject)
}

func GetProvince(subject string) string {
	return getRegexMatch("ST=([^,]+)", subject)
}

func GetLocality(subject string) string {
	return getRegexMatch("L=([^,]+)", subject)
}

func GetCommonName(subject string) string {
	return getRegexMatch("CN=([^,]+)", subject)
}

func getRegexMatch(regex, text string) string {
	cnRegex := regexp.MustCompile(regex)
	matches := cnRegex.FindStringSubmatch(text)

	if len(matches) != 2 {
		return ""
	}

	return matches[1]
}
