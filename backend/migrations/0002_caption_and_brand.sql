-- Caption + hashtags for the TikTok post that carries the carousel.
-- One editable row per carousel: the creator owns the final text, the AI only
-- proposes it.
CREATE TABLE carousel_captions (
    carousel_id  UUID PRIMARY KEY REFERENCES carousels(id) ON DELETE CASCADE,
    caption      TEXT NOT NULL DEFAULT '',
    hashtags     TEXT[] NOT NULL DEFAULT '{}',
    edited       BOOLEAN NOT NULL DEFAULT false,
    generated_at TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Brand kit extras (website, @handle, logo, display rules) live in the existing
-- projects.brand_json document, so no column changes are needed here. This
-- backfills the display defaults for projects created before the feature.
UPDATE projects
SET brand_json = brand_json || jsonb_build_object(
    'display', jsonb_build_object(
        'handle',  jsonb_build_object('scope', 'all',        'position', 'bottom_center'),
        'logo',    jsonb_build_object('scope', 'first_last', 'position', 'top_left'),
        'website', jsonb_build_object('scope', 'last',       'position', 'bottom_center')
    ))
WHERE NOT (brand_json ? 'display');
