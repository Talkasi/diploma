package importer

import "time"

// BBox описывает прямоугольную область.
type BBox struct {
	MinLat float64
	MinLon float64
	MaxLat float64
	MaxLon float64
	Pad    float64
}

// Contains проверяет, лежит ли точка внутри bbox с учётом padding.
func (b BBox) Contains(lat, lon float64) bool {
	minLat := b.MinLat - b.Pad
	minLon := b.MinLon - b.Pad
	maxLat := b.MaxLat + b.Pad
	maxLon := b.MaxLon + b.Pad
	return lat > minLat && lat < maxLat && lon > minLon && lon < maxLon
}

// ParseParams управляет парсером PBF.
type ParseParams struct {
	PBFPath        string
	City           string
	BBox           BBox
	Source         string
	LimitWays      int
	HighwayProfile string
}

// ImportParams параметры запуска импортера.
type ImportParams struct {
	ParseParams
	DryRun bool
}

// Node описывает узел графа.
type Node struct {
	ID  int64
	Lat float64
	Lon float64
}

// Coordinate хранит широту и долготу.
type Coordinate struct {
	Lat float64
	Lon float64
}

// Edge описывает ориентированное ребро.
type Edge struct {
	From      int64
	To        int64
	WayID     int64
	FromCoord Coordinate
	ToCoord   Coordinate
	LengthM   float64
	Criteria  map[string]any
	Tags      map[string]string
}

// ParsedData результат парсинга PBF.
type ParsedData struct {
	Nodes map[int64]Node
	Edges []Edge
	Stats ParseStats
}

// GraphWriteResult отражает записанные элементы.
type GraphWriteResult struct {
	NodesInserted int
	EdgesInserted int
}

// ImportStats агрегирует статистику всех стадий.
type ImportStats struct {
	Parse          ParseStats
	Write          GraphWriteResult
	DurationParse1 time.Duration
	DurationParse2 time.Duration
	DurationBuild  time.Duration
	DurationWrite  time.Duration
	DurationParse  time.Duration
}

// ParseStats собирает счётчики парсера.
type ParseStats struct {
	WaysRead                int
	WaysAccepted            int
	NodesNeeded             int
	NodesFound              int
	NodesInserted           int
	EdgesBuilt              int
	EdgesInserted           int
	SkippedEdgesMissingNode int
	SkippedEdgesOutOfBBox   int
	SkippedEdgesZeroLength  int
	DurationParse1          time.Duration
	DurationParse2          time.Duration
	DurationBuild           time.Duration
}

// Oneway описывает направление движения.
type Oneway int

const (
	OnewayBoth Oneway = iota
	OnewayForward
	OnewayBackward
)
