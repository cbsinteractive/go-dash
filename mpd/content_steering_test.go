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
