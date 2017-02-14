-- +migrate Up
CREATE TABLE pipeline_invoker_type (id BIGSERIAL PRIMARY KEY, name TEXT, author TEXT, description TEXT, identifier TEXT, type JSONB);
select create_unique_index('pipeline_invoker_type', 'IDX_PIPELINE_INVOKER_TYPE_NAME', 'name');

-- +migrate Down
DROP TABLE pipeline_invoker_type;