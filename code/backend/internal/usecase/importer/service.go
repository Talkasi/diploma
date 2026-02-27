package importer

import (
	"context"
	"fmt"
	"time"
)

// Service запускает импорт OSM PBF в БД.
type Service struct {
	osmReader OSMReader
	cityRepo  CityRepository
	graphRepo GraphRepository
}

// NewService создаёт сервис импорта.
func NewService(reader OSMReader, cityRepo CityRepository, graphRepo GraphRepository) *Service {
	return &Service{osmReader: reader, cityRepo: cityRepo, graphRepo: graphRepo}
}

// Import выполняет полный цикл импорта.
func (s *Service) Import(ctx context.Context, params ImportParams) (ImportStats, error) {
	var total ImportStats
	if params.PBFPath == "" {
		return total, fmt.Errorf("обязателен флаг --pbf")
	}
	if params.City == "" {
		return total, fmt.Errorf("обязателен флаг --city")
	}

	parsed, err := s.osmReader.Parse(ctx, params.ParseParams)
	total.DurationParse1 = parsed.Stats.DurationParse1
	total.DurationParse2 = parsed.Stats.DurationParse2
	total.DurationBuild = parsed.Stats.DurationBuild
	total.Parse = parsed.Stats
	if err != nil {
		return total, err
	}
	total.DurationParse = parsed.Stats.DurationParse1 + parsed.Stats.DurationParse2 + parsed.Stats.DurationBuild

	if params.DryRun {
		return total, nil
	}

	startWrite := time.Now()
	cityID, err := s.cityRepo.Upsert(ctx, params.City, params.Source, params.BBox)
	if err != nil {
		return total, fmt.Errorf("upsert города: %w", err)
	}

	writeRes, err := s.graphRepo.ReplaceCityGraph(ctx, cityID, mapToSlice(parsed.Nodes), parsed.Edges)
	total.Write = writeRes
	total.DurationWrite = time.Since(startWrite)
	if err != nil {
		return total, fmt.Errorf("сохранение графа: %w", err)
	}

	total.Parse.NodesInserted = writeRes.NodesInserted
	total.Parse.EdgesInserted = writeRes.EdgesInserted
	return total, nil
}

func mapToSlice(m map[int64]Node) []Node {
	res := make([]Node, 0, len(m))
	for _, n := range m {
		res = append(res, n)
	}
	return res
}
