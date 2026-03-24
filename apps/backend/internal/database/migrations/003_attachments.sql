CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    todo_id UUID NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    uploaded_by TEXT NOT NULL,
    download_key TEXT NOT NULL,
    file_size BIGINT,
    mime_type TEXT
);

CREATE INDEX idx_attachments_todo_id ON attachments(todo_id);
CREATE INDEX idx_attachments_uploaded_by ON attachments(uploaded_by);

CREATE TRIGGER update_attachments_updated_at
    BEFORE UPDATE ON attachments
    FOR EACH ROW
    EXECUTE FUNCTION trigger_update_updated_at_column();