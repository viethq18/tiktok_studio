-- ============================================================
-- TikTok Carousel AI Platform — core schema (blueprint §58-§72)
-- ============================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- §59 Users
CREATE TABLE users (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email            TEXT NOT NULL,
    name             TEXT NOT NULL DEFAULT '',
    avatar_url       TEXT NOT NULL DEFAULT '',
    provider         TEXT NOT NULL,
    provider_user_id TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_key ON users (lower(email));
CREATE UNIQUE INDEX users_provider_key ON users (provider, provider_user_id);

-- §162 Server-side sessions
CREATE TABLE sessions (
    id           TEXT PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX sessions_user_idx ON sessions (user_id);

-- §60 Projects
CREATE TABLE projects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    niche       TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    language    TEXT NOT NULL DEFAULT 'vi',
    status      TEXT NOT NULL DEFAULT 'active',
    brand_json  JSONB NOT NULL DEFAULT '{}'::jsonb,   -- §39 brand kit
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX projects_user_idx ON projects (user_id, created_at DESC) WHERE deleted_at IS NULL;

-- §61 Project AI context
CREATE TABLE project_ai_contexts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    version      INT NOT NULL,
    context_json JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, version)
);

-- §62 Project (carousel input) schemas
CREATE TABLE project_schemas (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    version             INT NOT NULL,
    schema_json         JSONB NOT NULL,
    ui_json             JSONB NOT NULL DEFAULT '{}'::jsonb,
    context_version     INT NOT NULL DEFAULT 1,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, version)
);

-- §63 Carousels
CREATE TABLE carousels (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title         TEXT NOT NULL DEFAULT 'Untitled carousel',
    status        TEXT NOT NULL DEFAULT 'draft', -- draft|generating|ready|archived|failed
    platform      TEXT NOT NULL DEFAULT 'tiktok',
    canvas_ratio  TEXT NOT NULL DEFAULT '4:5',
    canvas_width  INT NOT NULL DEFAULT 1080,
    canvas_height INT NOT NULL DEFAULT 1350,
    formula_id    TEXT NOT NULL DEFAULT '',
    thumbnail_key TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);
CREATE INDEX carousels_project_idx ON carousels (project_id, created_at DESC) WHERE deleted_at IS NULL;

-- §64 Carousel inputs
CREATE TABLE carousel_inputs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    carousel_id    UUID NOT NULL REFERENCES carousels(id) ON DELETE CASCADE,
    schema_version INT NOT NULL,
    input_json     JSONB NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX carousel_inputs_carousel_idx ON carousel_inputs (carousel_id);

-- §65 Carousel content
CREATE TABLE carousel_content (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    carousel_id  UUID NOT NULL REFERENCES carousels(id) ON DELETE CASCADE,
    version      INT NOT NULL,
    content_json JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (carousel_id, version)
);

-- §66 Carousel designs
CREATE TABLE carousel_designs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    carousel_id UUID NOT NULL REFERENCES carousels(id) ON DELETE CASCADE,
    version     INT NOT NULL,
    design_json JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (carousel_id, version)
);

-- §44 Carousel versions (snapshots)
CREATE TABLE carousel_versions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    carousel_id UUID NOT NULL REFERENCES carousels(id) ON DELETE CASCADE,
    version     INT NOT NULL,
    label       TEXT NOT NULL DEFAULT '',
    design_json JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (carousel_id, version)
);

-- §67 Assets
CREATE TABLE assets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id    UUID REFERENCES projects(id) ON DELETE CASCADE,
    carousel_id   UUID REFERENCES carousels(id) ON DELETE CASCADE,
    source        TEXT NOT NULL,            -- unsplash|upload|generated
    source_id     TEXT NOT NULL DEFAULT '',
    storage_key   TEXT NOT NULL,
    mime_type     TEXT NOT NULL,
    width         INT NOT NULL DEFAULT 0,
    height        INT NOT NULL DEFAULT 0,
    file_size     BIGINT NOT NULL DEFAULT 0,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb, -- §156 unsplash attribution
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX assets_user_idx ON assets (user_id, created_at DESC);
CREATE INDEX assets_carousel_idx ON assets (carousel_id);

CREATE TABLE carousel_assets (
    carousel_id UUID NOT NULL REFERENCES carousels(id) ON DELETE CASCADE,
    asset_id    UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    slide_id    TEXT NOT NULL DEFAULT '',
    role        TEXT NOT NULL DEFAULT 'background',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (carousel_id, asset_id, slide_id)
);

-- §68 Research sources
CREATE TABLE research_sources (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    UUID REFERENCES projects(id) ON DELETE CASCADE,
    carousel_id   UUID REFERENCES carousels(id) ON DELETE CASCADE,
    url           TEXT NOT NULL DEFAULT '',
    title         TEXT NOT NULL DEFAULT '',
    domain        TEXT NOT NULL DEFAULT '',
    source_type   TEXT NOT NULL DEFAULT 'web',
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX research_sources_carousel_idx ON research_sources (carousel_id);

-- §69 + §158 Generation jobs (with checkpointing)
CREATE TABLE generation_jobs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    carousel_id         UUID REFERENCES carousels(id) ON DELETE CASCADE,
    type                TEXT NOT NULL DEFAULT 'carousel_generation',
    status              TEXT NOT NULL DEFAULT 'queued',
    progress            INT NOT NULL DEFAULT 0,
    current_step        TEXT NOT NULL DEFAULT '',
    last_completed_step TEXT NOT NULL DEFAULT '',
    step_outputs        JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempt             INT NOT NULL DEFAULT 0,
    idempotency_key     TEXT,
    error_code          TEXT NOT NULL DEFAULT '',
    error_message       TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at          TIMESTAMPTZ,
    completed_at        TIMESTAMPTZ
);
CREATE INDEX generation_jobs_carousel_idx ON generation_jobs (carousel_id);
CREATE UNIQUE INDEX generation_jobs_idem_key ON generation_jobs (user_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

-- §70 AI generation audit
CREATE TABLE ai_generations (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id         UUID REFERENCES generation_jobs(id) ON DELETE SET NULL,
    user_id        UUID REFERENCES users(id) ON DELETE SET NULL,
    provider       TEXT NOT NULL,
    model          TEXT NOT NULL,
    task_type      TEXT NOT NULL,
    prompt_version TEXT NOT NULL DEFAULT 'v1',
    input_json     JSONB NOT NULL DEFAULT '{}'::jsonb,
    output_json    JSONB NOT NULL DEFAULT '{}'::jsonb,
    token_usage    JSONB NOT NULL DEFAULT '{}'::jsonb,
    latency_ms     INT NOT NULL DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'ok',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ai_generations_job_idx ON ai_generations (job_id);

-- §71 Image searches (cacheable)
CREATE TABLE image_searches (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    carousel_id  UUID REFERENCES carousels(id) ON DELETE CASCADE,
    slide_id     TEXT NOT NULL DEFAULT '',
    query        TEXT NOT NULL,
    provider     TEXT NOT NULL,
    page         INT NOT NULL DEFAULT 1,
    results_json JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX image_searches_carousel_idx ON image_searches (carousel_id, slide_id);

-- §72 Export jobs + exports
CREATE TABLE export_jobs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    carousel_id   UUID NOT NULL REFERENCES carousels(id) ON DELETE CASCADE,
    format        TEXT NOT NULL DEFAULT 'png_zip',
    status        TEXT NOT NULL DEFAULT 'queued',
    progress      INT NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    storage_key   TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at  TIMESTAMPTZ
);
CREATE INDEX export_jobs_carousel_idx ON export_jobs (carousel_id, created_at DESC);

-- §40 Font registry
CREATE TABLE font_registry (
    id            TEXT PRIMARY KEY,
    family        TEXT NOT NULL,
    css_stack     TEXT NOT NULL,
    weights       INT[] NOT NULL DEFAULT '{400,700}',
    vietnamese    BOOLEAN NOT NULL DEFAULT true,  -- §157 glyph coverage
    google_font   BOOLEAN NOT NULL DEFAULT true,
    enabled       BOOLEAN NOT NULL DEFAULT true
);

-- §118 Formula registry
CREATE TABLE formula_registry (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    structure       TEXT[] NOT NULL,
    recommended_for TEXT NOT NULL DEFAULT '',
    enabled         BOOLEAN NOT NULL DEFAULT true
);

-- §115 Analytics events / §89 usage
CREATE TABLE usage_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    project_id    UUID,
    carousel_id   UUID,
    event         TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX usage_events_user_idx ON usage_events (user_id, created_at DESC);

-- ---------- seeds ----------
INSERT INTO font_registry (id, family, css_stack, weights, vietnamese) VALUES
    ('inter',    'Inter',            'Inter, system-ui, sans-serif',      '{400,600,700,800}', true),
    ('roboto',   'Roboto',           'Roboto, system-ui, sans-serif',     '{400,500,700,900}', true),
    ('montserrat','Montserrat',      'Montserrat, system-ui, sans-serif', '{400,600,700,800}', true),
    ('nunito',   'Nunito',           'Nunito, system-ui, sans-serif',     '{400,700}',         true),
    ('playfair', 'Playfair Display', '"Playfair Display", Georgia, serif','{400,700,900}',     true),
    ('be-vietnam','Be Vietnam Pro',  '"Be Vietnam Pro", system-ui, sans-serif','{400,600,700,800}', true);

INSERT INTO formula_registry (id, name, description, structure, recommended_for) VALUES
    ('problem_solution', 'Problem → Solution', 'Nêu vấn đề rồi đưa giải pháp.', '{hook,problem,solution,proof,cta}', 'pain-point topics'),
    ('listicle',         'Listicle',           'Danh sách đánh số các điểm.',   '{hook,item,item,item,cta}',        'numbered topics'),
    ('myth_fact',        'Myth → Fact',        'Bóc trần hiểu lầm phổ biến.',   '{hook,myth,fact,myth,cta}',        'misconception topics'),
    ('how_to',           'How-to',             'Hướng dẫn từng bước.',          '{hook,step,step,step,cta}',        'tutorial topics'),
    ('mistake_fix',      'Mistake → Correction','Sai lầm và cách sửa.',         '{hook,mistake,correction,mistake,cta}', 'common mistakes'),
    ('before_after',     'Before → After',     'So sánh trước và sau.',         '{hook,before,after,takeaway,cta}',  'transformation topics'),
    ('question_answer',  'Question → Answer',  'Đặt câu hỏi rồi trả lời.',      '{hook,question,answer,answer,cta}', 'FAQ topics'),
    ('story_lesson',     'Story → Lesson',     'Kể chuyện rút ra bài học.',     '{hook,story,turn,lesson,cta}',      'narrative topics');
