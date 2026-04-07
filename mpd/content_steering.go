package mpd

// ServiceLocation is the value of BaseURL@serviceLocation and
// ContentSteering@defaultServiceLocation (ETSI TS 103 998).
type ServiceLocation string

// SteeredBaseURL is one attributed BaseURL for content steering: a URI and its
// serviceLocation label.
type SteeredBaseURL struct {
	Location ServiceLocation
	Value    string
}

// BaseURLValue is one DASH BaseURL element. ServiceLocation is used for
// DASH content steering (ETSI TS 103 998 / DASH-IF); omit for plain BaseURLs.
type BaseURLValue struct {
	ServiceLocation string `xml:"serviceLocation,attr,omitempty"`
	Value           string `xml:",chardata"`
}

// ContentSteering is the MPD-level ContentSteering element (steering server URI
// as character data; defaultServiceLocation names the preferred CDN at startup).
type ContentSteering struct {
	DefaultServiceLocation *string `xml:"defaultServiceLocation,attr,omitempty"`
	QueryBeforeStart       *bool   `xml:"queryBeforeStart,attr,omitempty"`
	ClientRequirement      *bool   `xml:"clientRequirement,attr,omitempty"`
	URI                    string  `xml:",chardata"`
}

// StringsToBaseURLs converts plain URL strings to BaseURLValue entries.
func StringsToBaseURLs(ss []string) []BaseURLValue {
	out := make([]BaseURLValue, len(ss))
	for i, s := range ss {
		out[i] = BaseURLValue{Value: s}
	}
	return out
}

// BaseURLsToStrings returns each BaseURL element's text (ignores serviceLocation).
func BaseURLsToStrings(bs []BaseURLValue) []string {
	out := make([]string, len(bs))
	for i, b := range bs {
		out[i] = b.Value
	}
	return out
}
