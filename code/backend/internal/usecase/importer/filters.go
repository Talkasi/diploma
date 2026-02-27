package importer

import "strings"

// HighwayProfile описывает допущенные и исключенные типы дорог.
type HighwayProfile struct {
	Allowed        map[string]struct{}
	Excluded       map[string]struct{}
	IncludeService bool
}

func DefaultCarProfile() HighwayProfile {
	allowed := []string{
		"motorway", "trunk", "primary", "secondary", "tertiary",
		"unclassified", "residential", "living_street",
		"motorway_link", "trunk_link", "primary_link", "secondary_link", "tertiary_link",
	}
	// service и track исключаем по умолчанию; service можно включить через IncludeService.
	excluded := []string{
		"footway", "pedestrian", "path", "steps", "cycleway", "bridleway", "corridor",
		"construction", "proposed", "platform", "track",
	}
	return HighwayProfile{Allowed: toSet(allowed), Excluded: toSet(excluded), IncludeService: false}
}

func toSet(items []string) map[string]struct{} {
	m := make(map[string]struct{}, len(items))
	for _, v := range items {
		m[v] = struct{}{}
	}
	return m
}

// AcceptHighway решает, оставлять ли way для профиля "car".
func AcceptHighway(profile HighwayProfile, tags map[string]string) bool {
	hw, ok := tags["highway"]
	if !ok {
		return false
	}
	hw = strings.ToLower(hw)
	if _, deny := profile.Excluded[hw]; deny {
		return false
	}
	if hw == "service" {
		if !profile.IncludeService {
			return false
		}
		// Консервативный подход: разрешаем только подъездные/парковочные проезды.
		if s, ok := tags["service"]; ok {
			sv := strings.ToLower(s)
			if sv == "driveway" || sv == "parking_aisle" {
				return true
			}
		}
		return false
	}
	if _, ok := profile.Allowed[hw]; ok {
		return true
	}
	return false
}

// BuildCriteria формирует нормализованные критерии из тегов way.
func BuildCriteria(tags map[string]string) map[string]any {
	res := make(map[string]any)

	if v, ok := tags["maxspeed"]; ok {
		val, raw, errFlag := ParseMaxspeed(v)
		if val != nil {
			res["maxspeed_kph"] = *val
		} else if raw != "" {
			res["maxspeed_raw"] = raw
		}
		if errFlag {
			res["maxspeed_parse_error"] = true
		}
	}

	if v, ok := tags["lanes"]; ok {
		val, raw, errFlag := ParseLanes(v)
		if val != nil {
			res["lanes"] = *val
		} else if raw != "" {
			res["lanes_raw"] = raw
		}
		if errFlag {
			res["lanes_parse_error"] = true
		}
	}

	if v, ok := tags["surface"]; ok {
		res["surface"] = strings.ToLower(v)
	}

	if v, ok := tags["lit"]; ok {
		if b, ok := ToBoolTag(v); ok {
			res["lit"] = b
		} else {
			res["lit_raw"] = v
		}
	}

	if v, ok := tags["bridge"]; ok {
		if b, ok := ToBoolTag(v); ok {
			res["bridge"] = b
		}
	}

	if v, ok := tags["tunnel"]; ok {
		if b, ok := ToBoolTag(v); ok {
			res["tunnel"] = b
		}
	}

	if v, ok := tags["access"]; ok {
		res["access"] = strings.ToLower(v)
	}

	res["incident_count"] = 0
	return res
}
