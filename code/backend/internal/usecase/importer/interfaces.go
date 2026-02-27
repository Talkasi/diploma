package importer

import (
	"context"

	"logistics/internal/domain"
)

// CityRepository описывает работу с сущностью города.
type CityRepository interface {
	Upsert(ctx context.Context, name string, source string, bbox BBox) (int64, error)
}

// GraphRepository отвечает за атомарную замену графа города.
type GraphRepository interface {
	ReplaceCityGraph(ctx context.Context, cityID int64, nodes []Node, edges []Edge) (GraphWriteResult, error)
}

// OSMReader выполняет парсинг PBF и строит граф в памяти.
type OSMReader interface {
	Parse(ctx context.Context, params ParseParams) (ParsedData, error)
}

// CityService предоставляет доступ к городам (для API), оставлен для совместимости.
// Здесь не используется, но сохраняем зависимость чистой архитектуры.
var _ domain.City
