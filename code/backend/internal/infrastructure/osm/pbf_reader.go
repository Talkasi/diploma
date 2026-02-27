package osm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/qedus/osmpbf"

	"logistics/internal/usecase/importer"
)

// PBFReader реализует двухпроходный парсинг OSM PBF.
type PBFReader struct{}

func NewPBFReader() *PBFReader {
	return &PBFReader{}
}

type wayData struct {
	id      int64
	nodeIDs []int64
	tags    map[string]string
	oneway  importer.Oneway
}

// Parse выполняет два прохода по файлу: ways -> nodeIDs, затем nodes.
func (r *PBFReader) Parse(ctx context.Context, params importer.ParseParams) (importer.ParsedData, error) {
	var result importer.ParsedData

	if params.PBFPath == "" {
		return result, fmt.Errorf("не указан путь к PBF")
	}
	if params.HighwayProfile == "" {
		params.HighwayProfile = "car"
	}

	profile := importer.DefaultCarProfile()

	// Первый проход: собираем подходящие ways и нужные узлы.
	parse1Start := time.Now()
	ways, neededNodes, stats1, err := r.passWays(ctx, params, profile)
	result.Stats.WaysRead = stats1.WaysRead
	result.Stats.WaysAccepted = stats1.WaysAccepted
	result.Stats.DurationParse1 = time.Since(parse1Start)
	if err != nil {
		return result, err
	}

	// Второй проход: читаем координаты нужных узлов.
	parse2Start := time.Now()
	nodes, stats2, err := r.passNodes(ctx, params.PBFPath, neededNodes)
	result.Stats.NodesNeeded = stats2.NodesNeeded
	result.Stats.NodesFound = stats2.NodesFound
	result.Stats.DurationParse2 = time.Since(parse2Start)
	if err != nil {
		return result, err
	}

	// Пост-обработка: строим рёбра.
	buildStart := time.Now()
	edges, buildStats := r.buildEdges(params.BBox, ways, nodes)
	result.Stats.EdgesBuilt = buildStats.EdgesBuilt
	result.Stats.SkippedEdgesMissingNode = buildStats.SkippedEdgesMissingNode
	result.Stats.SkippedEdgesOutOfBBox = buildStats.SkippedEdgesOutOfBBox
	result.Stats.SkippedEdgesZeroLength = buildStats.SkippedEdgesZeroLength
	result.Stats.DurationBuild = time.Since(buildStart)

	result.Nodes = nodes
	result.Edges = edges

	// Итоговые счётчики.
	result.Stats.NodesInserted = len(nodes)
	result.Stats.EdgesInserted = len(edges)

	return result, nil
}

// passWays читает все ways и отбирает подходящие.
func (r *PBFReader) passWays(ctx context.Context, params importer.ParseParams, profile importer.HighwayProfile) ([]wayData, map[int64]struct{}, importer.ParseStats, error) {
	var stats importer.ParseStats
	ways := make([]wayData, 0)
	neededNodes := make(map[int64]struct{})

	dec, closer, err := newDecoder(params.PBFPath)
	if err != nil {
		return nil, nil, stats, err
	}
	defer closer()

	for {
		if ctx.Err() != nil {
			return nil, nil, stats, ctx.Err()
		}
		v, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, nil, stats, fmt.Errorf("ошибка чтения PBF (ways): %w", err)
		}
		way, ok := v.(*osmpbf.Way)
		if !ok {
			continue
		}
		stats.WaysRead++
		if params.LimitWays > 0 && stats.WaysRead > params.LimitWays {
			break
		}
		tags := convertTags(way.Tags)
		if !importer.AcceptHighway(profile, tags) {
			continue
		}
		stats.WaysAccepted++
		ow := importer.ParseOneway(tags["oneway"])
		wd := wayData{id: way.ID, tags: tags, oneway: ow, nodeIDs: make([]int64, len(way.NodeIDs))}
		copy(wd.nodeIDs, way.NodeIDs)
		ways = append(ways, wd)
		for _, nid := range way.NodeIDs {
			neededNodes[nid] = struct{}{}
		}
	}

	return ways, neededNodes, stats, nil
}

// passNodes читает только нужные узлы.
func (r *PBFReader) passNodes(ctx context.Context, path string, needed map[int64]struct{}) (map[int64]importer.Node, importer.ParseStats, error) {
	var stats importer.ParseStats
	nodes := make(map[int64]importer.Node, len(needed))
	stats.NodesNeeded = len(needed)

	dec, closer, err := newDecoder(path)
	if err != nil {
		return nil, stats, err
	}
	defer closer()

	for {
		if ctx.Err() != nil {
			return nil, stats, ctx.Err()
		}
		v, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, stats, fmt.Errorf("ошибка чтения PBF (nodes): %w", err)
		}
		node, ok := v.(*osmpbf.Node)
		if !ok {
			continue
		}
		if _, want := needed[node.ID]; !want {
			continue
		}
		nodes[node.ID] = importer.Node{ID: node.ID, Lat: node.Lat, Lon: node.Lon}
		stats.NodesFound++
	}

	return nodes, stats, nil
}

// buildEdges создаёт рёбра между последовательными узлами.
func (r *PBFReader) buildEdges(bbox importer.BBox, ways []wayData, nodes map[int64]importer.Node) ([]importer.Edge, importer.ParseStats) {
	var stats importer.ParseStats
	edges := make([]importer.Edge, 0)

	for _, w := range ways {
		for i := 0; i < len(w.nodeIDs)-1; i++ {
			fromID := w.nodeIDs[i]
			toID := w.nodeIDs[i+1]
			fromNode, okFrom := nodes[fromID]
			toNode, okTo := nodes[toID]
			if !okFrom || !okTo {
				stats.SkippedEdgesMissingNode++
				continue
			}
			if !bbox.Contains(fromNode.Lat, fromNode.Lon) || !bbox.Contains(toNode.Lat, toNode.Lon) {
				stats.SkippedEdgesOutOfBBox++
				continue
			}
			length := importer.HaversineDistance(
				importer.Coordinate{Lat: fromNode.Lat, Lon: fromNode.Lon},
				importer.Coordinate{Lat: toNode.Lat, Lon: toNode.Lon},
			)
			if length <= 0 {
				stats.SkippedEdgesZeroLength++
				continue
			}

			stats.EdgesBuilt++
			// Строим направления.
			switch w.oneway {
			case importer.OnewayForward:
				edges = append(edges, makeEdge(fromNode, toNode, w, length))
			case importer.OnewayBackward:
				edges = append(edges, makeEdge(toNode, fromNode, w, length))
			default:
				edges = append(edges, makeEdge(fromNode, toNode, w, length))
				edges = append(edges, makeEdge(toNode, fromNode, w, length))
			}
		}
	}

	// Оставляем только используемые узлы.
	used := make(map[int64]struct{})
	for _, e := range edges {
		used[e.From] = struct{}{}
		used[e.To] = struct{}{}
	}
	for id := range nodes {
		if _, ok := used[id]; !ok {
			delete(nodes, id)
		}
	}

	return edges, stats
}

func makeEdge(from importer.Node, to importer.Node, w wayData, length float64) importer.Edge {
	return importer.Edge{
		From:      from.ID,
		To:        to.ID,
		WayID:     w.id,
		FromCoord: importer.Coordinate{Lat: from.Lat, Lon: from.Lon},
		ToCoord:   importer.Coordinate{Lat: to.Lat, Lon: to.Lon},
		LengthM:   length,
		Criteria:  importer.BuildCriteria(w.tags),
		Tags:      w.tags,
	}
}

func newDecoder(path string) (*osmpbf.Decoder, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("не удалось открыть PBF: %w", err)
	}
	dec := osmpbf.NewDecoder(f)
	dec.SetBufferSize(osmpbf.MaxBlobSize)
	if err := dec.Start(runtime.GOMAXPROCS(0)); err != nil {
		f.Close()
		return nil, nil, fmt.Errorf("ошибка старта декодера: %w", err)
	}
	return dec, func() { f.Close() }, nil
}

func convertTags(src map[string]string) map[string]string {
	if src == nil {
		return map[string]string{}
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
