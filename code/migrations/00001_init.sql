CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE cities (
    id         bigserial PRIMARY KEY,
    name       text        NOT NULL UNIQUE,
    osm_source text        NULL,
    bounds     geometry(Polygon, 4326) NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE nodes (
    id      bigint NOT NULL,
    city_id bigint NOT NULL REFERENCES cities(id) ON DELETE CASCADE,
    lat     double precision NOT NULL,
    lon     double precision NOT NULL,
    geom    geometry(Point, 4326) NOT NULL,
    PRIMARY KEY (city_id, id)
);

CREATE TABLE edges (
    id        bigserial PRIMARY KEY,
    city_id   bigint NOT NULL REFERENCES cities(id) ON DELETE CASCADE,
    from_node bigint NOT NULL,
    to_node   bigint NOT NULL,
    osm_way_id bigint NULL,
    geom      geometry(LineString, 4326) NULL,
    length_m  double precision NOT NULL,
    criteria  jsonb NOT NULL DEFAULT '{}'::jsonb,
    tags      jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_nodes_geom ON nodes USING gist (geom);
CREATE INDEX IF NOT EXISTS idx_edges_geom ON edges USING gist (geom);
CREATE INDEX IF NOT EXISTS idx_edges_from ON edges USING btree (city_id, from_node);
CREATE INDEX IF NOT EXISTS idx_edges_to   ON edges USING btree (city_id, to_node);

-- DROP TABLE IF EXISTS edges;
-- DROP TABLE IF EXISTS nodes;
-- DROP TABLE IF EXISTS cities;
