-- +migrate Up
CREATE TABLE pipeline_invoker_type (id BIGSERIAL PRIMARY KEY, name TEXT, author TEXT, description TEXT, identifier TEXT, type JSONB);
select create_unique_index('pipeline_invoker_type', 'IDX_PIPELINE_INVOKER_TYPE_NAME', 'name');

CREATE TABLE pipeline_invoker (uuid TEXT PRIMARY KEY, application_id BIGINT, pipeline_id BIGINT,  environment_id BIGINT, pipeline_invoker_type_id BIGINT);
select create_foreign_key('FK_PIPELINE_INVOKER_PIPELINE', 'pipeline_invoker', 'pipeline', 'pipeline_id', 'id');
select create_foreign_key('FK_PIPELINE_INVOKER_APPLICATION', 'pipeline_invoker', 'application', 'application_id', 'id');
select create_foreign_key('FK_PIPELINE_INVOKER_ENVIRONMENT', 'pipeline_invoker', 'environment', 'environment_id', 'id');
select create_foreign_key('FK_PIPELINE_INVOKER_TYPE', 'pipeline_invoker', 'pipeline_invoker_type', 'pipeline_invoker_type_id', 'id');

-- +migrate Down
DROP TABLE pipeline_invoker;
DROP TABLE pipeline_invoker_type;