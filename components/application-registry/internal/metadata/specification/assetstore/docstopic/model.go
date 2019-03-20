package docstopic

type SpecEntry struct {
	Url string
	Key string
}

type Entry struct {
	Id            string
	DisplayName   string
	Description   string
	ApiSpec       *SpecEntry
	EventsSpec    *SpecEntry
	Documentation *SpecEntry
	Labels        map[string]string
}
