package importer

import (
	"math"
	"testing"
)

func TestParseBBox(t *testing.T) {
	bbox, err := ParseBBox("10,20,30,40", 0.1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bbox.MinLat != 10 || bbox.MinLon != 20 || bbox.MaxLat != 30 || bbox.MaxLon != 40 || bbox.Pad != 0.1 {
		t.Fatalf("unexpected bbox: %+v", bbox)
	}

	if _, err := ParseBBox("bad", 0); err == nil {
		t.Fatalf("expected error for bad input")
	}
}

func TestParseOneway(t *testing.T) {
	cases := map[string]Oneway{
		"yes":     OnewayForward,
		"1":       OnewayForward,
		"true":    OnewayForward,
		"-1":      OnewayBackward,
		"reverse": OnewayBackward,
		"":        OnewayBoth,
	}
	for in, want := range cases {
		if got := ParseOneway(in); got != want {
			t.Fatalf("%s => %v, want %v", in, got, want)
		}
	}
}

func TestAcceptHighway(t *testing.T) {
	profile := DefaultCarProfile()

	tags := map[string]string{"highway": "primary"}
	if !AcceptHighway(profile, tags) {
		t.Fatalf("primary should be accepted")
	}

	tags = map[string]string{"highway": "footway"}
	if AcceptHighway(profile, tags) {
		t.Fatalf("footway should be excluded")
	}

	tags = map[string]string{"highway": "service", "service": "driveway"}
	if AcceptHighway(profile, tags) {
		t.Fatalf("service without IncludeService should be excluded")
	}
}

func TestParseMaxspeed(t *testing.T) {
	val, raw, errFlag := ParseMaxspeed("60")
	if errFlag || val == nil || *val != 60 || raw != "60" {
		t.Fatalf("unexpected parse: val=%v raw=%s err=%v", val, raw, errFlag)
	}

	_, raw, errFlag = ParseMaxspeed("fast")
	if !errFlag || raw != "fast" {
		t.Fatalf("expected parse error")
	}
}

func TestParseLanes(t *testing.T) {
	val, raw, errFlag := ParseLanes("2")
	if errFlag || val == nil || *val != 2 || raw != "2" {
		t.Fatalf("unexpected parse lanes")
	}

	_, _, errFlag = ParseLanes("many")
	if !errFlag {
		t.Fatalf("expected parse error")
	}
}

func TestHaversineDistance(t *testing.T) {
	a := Coordinate{Lat: 0, Lon: 0}
	b := Coordinate{Lat: 0, Lon: 1}
	d := HaversineDistance(a, b)
	if math.Abs(d-111319.5) > 1000 {
		t.Fatalf("unexpected distance: %f", d)
	}
}
