CREATE TABLE currencies (
    code char(3) PRIMARY KEY,
    numeric_code smallint NOT NULL UNIQUE,
    name text,
    latin_name text NOT NULL,
    minor_units smallint NOT NULL,
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (code ~ '^[A-Z]{3}$'),
    CHECK (numeric_code > 0),
    CHECK (minor_units >= 0)
);

CREATE TABLE fx_sources (
    code text PRIMARY KEY,
    type text NOT NULL,
    name text NOT NULL,
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (code <> ''),
    CHECK (type IN ('cbr', 'quik', 'moex', 'broker'))
);

CREATE TABLE fx_rates (
    id bigserial PRIMARY KEY,
    source_code text NOT NULL REFERENCES fx_sources(code),
    base_code char(3) NOT NULL REFERENCES currencies(code),
    quote_code char(3) NOT NULL REFERENCES currencies(code),
    value numeric(38, 18) NOT NULL,
    observed_at timestamptz NOT NULL,
    valid_from timestamptz NOT NULL,
    valid_to timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (base_code <> quote_code),
    CHECK (value > 0),
    CHECK (valid_to IS NULL OR valid_to > valid_from),
    UNIQUE (source_code, base_code, quote_code, observed_at)
);

CREATE INDEX fx_rates_pair_actual_idx
    ON fx_rates (base_code, quote_code, valid_from, valid_to);

CREATE INDEX fx_rates_source_observed_idx
    ON fx_rates (source_code, observed_at DESC);

CREATE TABLE source_priority_rules (
    id bigserial PRIMARY KEY,
    base_code char(3) REFERENCES currencies(code),
    quote_code char(3) REFERENCES currencies(code),
    source_code text NOT NULL REFERENCES fx_sources(code),
    priority smallint NOT NULL,
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (priority > 0),
    CHECK (base_code IS NULL OR quote_code IS NULL OR base_code <> quote_code)
);

CREATE UNIQUE INDEX source_priority_rules_scope_idx
    ON source_priority_rules (
        COALESCE(base_code, '*'),
        COALESCE(quote_code, '*'),
        source_code
    );
