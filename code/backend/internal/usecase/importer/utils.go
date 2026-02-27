package importer

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var errBadBBox = errors.New("некорректный формат bbox, ожидается minLat,minLon,maxLat,maxLon")

// ParseBBox разбирает строку bbox.
func ParseBBox(raw string, pad float64) (BBox, error) {
	parts := strings.Split(raw, ",")
	if len(parts) != 4 {
		return BBox{}, errBadBBox
	}
	vals := make([]float64, 4)
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return BBox{}, errBadBBox
		}
		vals[i] = v
	}
	if vals[0] >= vals[2] || vals[1] >= vals[3] {
		return BBox{}, errBadBBox
	}
	return BBox{MinLat: vals[0], MinLon: vals[1], MaxLat: vals[2], MaxLon: vals[3], Pad: pad}, nil
}

// ParseOneway определяет направление движения.
func ParseOneway(tag string) Oneway {
	switch strings.ToLower(tag) {
	case "yes", "true", "1", "forward":
		return OnewayForward
	case "-1", "reverse", "backward":
		return OnewayBackward
	default:
		return OnewayBoth
	}
}

// ParseMaxspeed пытается выделить числовое значение км/ч.
func ParseMaxspeed(raw string) (val *float64, rawOut string, parseErr bool) {
	cleaned := strings.ToLower(strings.TrimSpace(raw))
	if cleaned == "" {
		return nil, "", false
	}
	// Берём первое число в строке.
	re := regexp.MustCompile(`([0-9]+)`)
	match := re.FindStringSubmatch(cleaned)
	if len(match) < 2 {
		return nil, raw, true
	}
	num, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return nil, raw, true
	}
	return &num, raw, false
}

// ParseLanes возвращает количество полос.
func ParseLanes(raw string) (val *int, rawOut string, parseErr bool) {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return nil, "", false
	}
	num, err := strconv.Atoi(cleaned)
	if err != nil {
		return nil, raw, true
	}
	return &num, raw, false
}

// ToBoolTag интерпретирует yes/no теги.
func ToBoolTag(raw string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "yes", "true", "1":
		return true, true
	case "no", "false", "0":
		return false, true
	default:
		return false, false
	}
}

// HaversineDistance считает расстояние между двумя координатами в метрах.
func HaversineDistance(a, b Coordinate) float64 {
	const earthRadius = 6371000.0
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLon := (b.Lon - a.Lon) * math.Pi / 180

	sinLat := math.Sin(dLat / 2)
	sinLon := math.Sin(dLon / 2)

	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLon*sinLon
	c := 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
	return earthRadius * c
}
