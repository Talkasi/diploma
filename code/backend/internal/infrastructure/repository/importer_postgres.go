package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"logistics/internal/usecase/importer"
)

// ImportRepository реализует сохранение результатов импорта в Postgres.
type ImportRepository struct {
	pool *pgxpool.Pool
}

func NewImportRepository(pool *pgxpool.Pool) *ImportRepository {
	return &ImportRepository{pool: pool}
}

// Upsert вставляет или обновляет город, возвращая его id.
func (r *ImportRepository) Upsert(ctx context.Context, name string, source string, bbox importer.BBox) (int64, error) {
	const q = `
INSERT INTO cities (name, osm_source, bounds)
VALUES ($1, $2, ST_SetSRID(ST_MakeEnvelope($3, $4, $5, $6), 4326))
ON CONFLICT (name) DO UPDATE
    SET osm_source = EXCLUDED.osm_source,
        bounds = EXCLUDED.bounds
RETURNING id;
`
	row := r.pool.QueryRow(ctx, q, name, source, bbox.MinLon, bbox.MinLat, bbox.MaxLon, bbox.MaxLat)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("upsert city: %w", err)
	}
	return id, nil
}

// ReplaceCityGraph удаляет текущий граф города и вставляет новый в транзакции.
func (r *ImportRepository) ReplaceCityGraph(ctx context.Context, cityID int64, nodes []importer.Node, edges []importer.Edge) (importer.GraphWriteResult, error) {
	var res importer.GraphWriteResult

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return res, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "DELETE FROM edges WHERE city_id=$1", cityID); err != nil {
		return res, fmt.Errorf("delete edges: %w", err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM nodes WHERE city_id=$1", cityID); err != nil {
		return res, fmt.Errorf("delete nodes: %w", err)
	}

	if len(nodes) > 0 {
		if err := copyNodes(ctx, tx, cityID, nodes); err != nil {
			return res, err
		}
		res.NodesInserted = len(nodes)
	}

	if len(edges) > 0 {
		if err := copyEdges(ctx, tx, cityID, edges); err != nil {
			return res, err
		}
		res.EdgesInserted = len(edges)
	}

	if err := tx.Commit(ctx); err != nil {
		return res, fmt.Errorf("commit: %w", err)
	}
	return res, nil
}

func copyNodes(ctx context.Context, tx pgx.Tx, cityID int64, nodes []importer.Node) error {
	if _, err := tx.Exec(ctx, `CREATE TEMP TABLE tmp_nodes (city_id bigint, id bigint, lat double precision, lon double precision) ON COMMIT DROP`); err != nil {
		return fmt.Errorf("create temp nodes: %w", err)
	}

	rows := make([][]any, 0, len(nodes))
	for _, n := range nodes {
		rows = append(rows, []any{cityID, n.ID, n.Lat, n.Lon})
	}

	_, err := tx.CopyFrom(ctx, pgx.Identifier{"tmp_nodes"}, []string{"city_id", "id", "lat", "lon"}, pgx.CopyFromRows(rows))
	if err != nil {
		return fmt.Errorf("copy nodes: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO nodes (city_id, id, lat, lon, geom)
SELECT city_id, id, lat, lon,
       ST_SetSRID(ST_MakePoint(lon, lat), 4326)
FROM tmp_nodes;
`)
	if err != nil {
		return fmt.Errorf("insert nodes: %w", err)
	}
	return nil
}

func copyEdges(ctx context.Context, tx pgx.Tx, cityID int64, edges []importer.Edge) error {
	if _, err := tx.Exec(ctx, `CREATE TEMP TABLE tmp_edges (
        city_id bigint,
        from_node bigint,
        to_node bigint,
        osm_way_id bigint,
        from_lat double precision,
        from_lon double precision,
        to_lat double precision,
        to_lon double precision,
        length_m double precision,
        criteria jsonb,
        tags jsonb
    ) ON COMMIT DROP`); err != nil {
		return fmt.Errorf("create temp edges: %w", err)
	}

	rows := make([][]any, 0, len(edges))
	for _, e := range edges {
		criteriaJSON, _ := json.Marshal(e.Criteria)
		tagsJSON, _ := json.Marshal(e.Tags)
		rows = append(rows, []any{
			cityID,
			e.From,
			e.To,
			e.WayID,
			e.FromCoord.Lat,
			e.FromCoord.Lon,
			e.ToCoord.Lat,
			e.ToCoord.Lon,
			e.LengthM,
			criteriaJSON,
			tagsJSON,
		})
	}

	_, err := tx.CopyFrom(ctx, pgx.Identifier{"tmp_edges"}, []string{
		"city_id", "from_node", "to_node", "osm_way_id", "from_lat", "from_lon", "to_lat", "to_lon", "length_m", "criteria", "tags",
	}, pgx.CopyFromRows(rows))
	if err != nil {
		return fmt.Errorf("copy edges: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO edges (city_id, from_node, to_node, osm_way_id, geom, length_m, criteria, tags)
SELECT city_id, from_node, to_node, osm_way_id,
       ST_SetSRID(ST_MakeLine(ST_MakePoint(from_lon, from_lat), ST_MakePoint(to_lon, to_lat)), 4326),
       length_m, criteria, tags
FROM tmp_edges;
`)
	if err != nil {
		return fmt.Errorf("insert edges: %w", err)
	}
	return nil
}

var _ importer.CityRepository = (*ImportRepository)(nil)
var _ importer.GraphRepository = (*ImportRepository)(nil)
