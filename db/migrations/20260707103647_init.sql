-- +goose Up
CREATE TABLE tasks (
	id UUID PRIMARY KEY,
	proc_status VARCHAR(10) NOT NULL CHECK (proc_status IN ('pending', 'done', 'failed')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE results (
	id BIGSERIAL PRIMARY KEY,
	task_id UUID NOT NULL REFERENCES tasks(id),
	url TEXT NOT NULL,
	status_code SMALLINT NOT NULL DEFAULT 0,
	duration_ms INT	NOT NULL DEFAULT 0,
	error TEXT
);

CREATE INDEX idx_results_on_task ON results(task_id);


-- +goose Down
DROP TABLE results;
DROP TABLE tasks;
