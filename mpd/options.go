package mpd

// Options configures MPD read/write behavior for optional extensions.
type Options struct {
	// ContentSteering, when non-nil, is merged onto the MPD after ReadFromStringWithOptions
	// decodes XML (see ApplyContentSteeringOptions), so Write encodes steering elements.
	ContentSteering *ContentSteeringOptions
}

// ContentSteeringOptions holds header- or policy-derived values used when rendering
// attributed BaseURL and ContentSteering elements (ETSI TS 103 998).
type ContentSteeringOptions struct {
	// SteeringURI is the steering server or manifest URI (character data of the
	// ContentSteering element).
	SteeringURI string

	// DefaultServiceLocation is ContentSteering@defaultServiceLocation: the preferred
	// service location until steering resolves.
	DefaultServiceLocation ServiceLocation

	// SteeringBaseURLs lists BaseURL elements with serviceLocation (one CDN root per location).
	SteeringBaseURLs []SteeredBaseURL

	// QueryBeforeStart is ContentSteering@queryBeforeStart when non-nil.
	QueryBeforeStart *bool

	// ClientRequirement is ContentSteering@clientRequirement when non-nil.
	ClientRequirement *bool
}
