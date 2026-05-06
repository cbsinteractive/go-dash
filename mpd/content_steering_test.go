package mpd

import (
	"strings"
	"testing"

	. "github.com/cbsinteractive/go-dash/v3/helpers/ptrs"
	"github.com/cbsinteractive/go-dash/v3/helpers/require"
)

func TestReadFromStringWithOptionsNilMatchesReadFromString(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" type="static" mediaPresentationDuration="PT10S" minBufferTime="PT2S">
  <Period>
    <BaseURL>https://origin.example/</BaseURL>
  </Period>
</MPD>`
	m1, err1 := ReadFromString(xml)
	require.NoError(t, err1)
	m2, err2 := ReadFromStringWithOptions(xml, nil)
	require.NoError(t, err2)
	require.EqualStringSlice(t, BaseURLsToStrings(m1.Periods[0].BaseURL), BaseURLsToStrings(m2.Periods[0].BaseURL))
}

func TestReadFromStringWithOptionsNonNilOptsParsesContentSteering(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" type="static" mediaPresentationDuration="PT10S" minBufferTime="PT2S">
  <BaseURL serviceLocation="fastly">https://fastly.example/out/v1/x/</BaseURL>
  <BaseURL serviceLocation="akamai">https://akamai.example/out/v1/x/</BaseURL>
  <ContentSteering defaultServiceLocation="fastly">https://steer.example/api/v1/steer</ContentSteering>
  <Period id="0">
    <AdaptationSet mimeType="video/mp4">
      <Representation id="v1" bandwidth="1000000" width="1280" height="720" codecs="avc1.64001F">
        <BaseURL>seg/</BaseURL>
      </Representation>
    </AdaptationSet>
  </Period>
</MPD>`
	opts := &Options{ContentSteering: &ContentSteeringOptions{}}
	m, err := ReadFromStringWithOptions(xml, opts)
	require.NoError(t, err)
	if m.ContentSteering == nil {
		t.Fatal("expected ContentSteering")
	}
	if strings.TrimSpace(m.ContentSteering.URI) != "https://steer.example/api/v1/steer" {
		t.Fatalf("steering URI %q", m.ContentSteering.URI)
	}
	if m.ContentSteering.DefaultServiceLocation == nil || *m.ContentSteering.DefaultServiceLocation != "fastly" {
		t.Fatalf("defaultServiceLocation %+v", m.ContentSteering.DefaultServiceLocation)
	}
	if len(m.BaseURL) != 2 {
		t.Fatalf("base URLs: %d", len(m.BaseURL))
	}
	if m.BaseURL[0].ServiceLocation != "fastly" || m.BaseURL[0].Value != "https://fastly.example/out/v1/x/" {
		t.Fatalf("first BaseURL %+v", m.BaseURL[0])
	}
	if m.BaseURL[1].ServiceLocation != "akamai" {
		t.Fatalf("second BaseURL %+v", m.BaseURL[1])
	}
}

func TestContentSteeringRoundTripWriteRead(t *testing.T) {
	m := NewMPD(DASH_PROFILE_ONDEMAND, "PT6M16S", "PT1.97S")
	m.BaseURL = []BaseURLValue{
		{ServiceLocation: "fastly", Value: "https://f.example/path/"},
		{ServiceLocation: "akamai", Value: "https://a.example/path/"},
	}
	m.ContentSteering = &ContentSteering{
		DefaultServiceLocation: Strptr("fastly"),
		URI:                    "https://steer.example/steer",
	}
	out, err := m.WriteToString()
	require.NoError(t, err)
	m2, err := ReadFromStringWithOptions(out, &Options{ContentSteering: &ContentSteeringOptions{}})
	require.NoError(t, err)
	if m2.ContentSteering == nil || strings.TrimSpace(m2.ContentSteering.URI) != m.ContentSteering.URI {
		t.Fatalf("ContentSteering after round trip: %+v", m2.ContentSteering)
	}
	if len(m2.BaseURL) != 2 {
		t.Fatal(len(m2.BaseURL))
	}
}

func TestContentSteeringOptionalAttributesRoundTrip(t *testing.T) {
	m := NewMPD(DASH_PROFILE_ONDEMAND, "PT1S", "PT1S")
	m.ContentSteering = &ContentSteering{
		DefaultServiceLocation: Strptr("c"),
		QueryBeforeStart:       Boolptr(true),
		ClientRequirement:      Boolptr(false),
		URI:                    "https://steer/x",
	}
	out, err := m.WriteToString()
	require.NoError(t, err)
	m2, err := ReadFromString(out)
	require.NoError(t, err)
	require.NotNil(t, m2.ContentSteering)
	if m2.ContentSteering.QueryBeforeStart == nil || !*m2.ContentSteering.QueryBeforeStart {
		t.Fatal("queryBeforeStart")
	}
	if m2.ContentSteering.ClientRequirement == nil || *m2.ContentSteering.ClientRequirement {
		t.Fatal("clientRequirement")
	}
}

func TestStringsToBaseURLsAndBack(t *testing.T) {
	ss := []string{"a", "b"}
	b := StringsToBaseURLs(ss)
	require.EqualStringSlice(t, ss, BaseURLsToStrings(b))
}

// ISO/IEC 23009-1 Amd 1 (DASH-MPD.xsd) requires ContentSteering to be the last
// child of MPD. Strict XSD validators reject manifests that emit it earlier.
func TestContentSteeringEmittedAfterPeriodAndUTCTiming(t *testing.T) {
	m := NewDynamicMPD(DASH_PROFILE_LIVE, "1970-01-01T00:00:00Z", "PT2S",
		AttrMediaPresentationDuration("PT10S"))
	m.BaseURL = []BaseURLValue{
		{ServiceLocation: "fastly", Value: "https://f.example/x/"},
	}
	m.ContentSteering = &ContentSteering{
		DefaultServiceLocation: Strptr("fastly"),
		URI:                    "https://steer.example/steer",
	}
	out, err := m.WriteToString()
	require.NoError(t, err)

	baseURLIdx := strings.Index(out, "<BaseURL")
	periodOpenIdx := strings.Index(out, "<Period")
	periodCloseIdx := strings.LastIndex(out, "</Period>")
	utcIdx := strings.Index(out, "<UTCTiming")
	csIdx := strings.Index(out, "<ContentSteering")
	if baseURLIdx < 0 || periodOpenIdx < 0 || periodCloseIdx < 0 || utcIdx < 0 || csIdx < 0 {
		t.Fatalf("missing expected element in output: %s", out)
	}
	if !(baseURLIdx < periodOpenIdx) {
		t.Fatalf("BaseURL must come before Period: %s", out)
	}
	if !(csIdx > periodCloseIdx) {
		t.Fatalf("ContentSteering must come after </Period>: %s", out)
	}
	if !(csIdx > utcIdx) {
		t.Fatalf("ContentSteering must come after UTCTiming: %s", out)
	}
}

func TestReadFromStringWithOptionsAppliesSteeringOptionsForWrite(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" type="static" mediaPresentationDuration="PT10S" minBufferTime="PT2S">
  <Period id="0"><AdaptationSet mimeType="video/mp4">
    <Representation id="v1" bandwidth="1000000" codecs="avc1.64001F"><BaseURL>v/</BaseURL></Representation>
  </AdaptationSet></Period>
</MPD>`
	m, err := ReadFromStringWithOptions(xml, &Options{ContentSteering: &ContentSteeringOptions{
		SteeringURI:            "https://steer.example/steer",
		DefaultServiceLocation: "cdn-a",
		SteeringBaseURLs: []SteeredBaseURL{
			{Location: "cdn-a", Value: "https://cdn-a.example/out/"},
			{Location: "cdn-b", Value: "https://cdn-b.example/out/"},
		},
	}})
	require.NoError(t, err)
	out, err := m.WriteToString()
	require.NoError(t, err)
	if !strings.Contains(out, "<ContentSteering") || !strings.Contains(out, "https://steer.example/steer") {
		t.Fatalf("missing ContentSteering in output: %s", out)
	}
	if !strings.Contains(out, `serviceLocation="cdn-a"`) || !strings.Contains(out, "https://cdn-a.example/out/") {
		t.Fatalf("missing steered BaseURL in output: %s", out)
	}
}
