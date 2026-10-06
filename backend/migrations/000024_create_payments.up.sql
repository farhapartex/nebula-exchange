CREATE TYPE payment_status AS ENUM ('OPEN', 'PAID', 'EXPIRED', 'FAILED', 'REFUNDED', 'DISPUTED');

CREATE TABLE payments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    plan_id TEXT NOT NULL REFERENCES plans (id) ON DELETE RESTRICT,
    status payment_status NOT NULL,
    currency CHAR(3) NOT NULL,
    subtotal_cents BIGINT NOT NULL,
    discount_percent INTEGER NOT NULL,
    discount_cents BIGINT NOT NULL,
    total_cents BIGINT NOT NULL,
    stripe_checkout_session_id TEXT,
    stripe_payment_intent_id TEXT,
    checkout_url TEXT,
    checkout_expires_at TIMESTAMPTZ NOT NULL,
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT payments_stripe_checkout_session_id_unique UNIQUE (stripe_checkout_session_id),
    CONSTRAINT payments_stripe_payment_intent_id_unique UNIQUE (stripe_payment_intent_id),
    CONSTRAINT payments_amounts_valid CHECK (
        subtotal_cents > 0
        AND discount_cents >= 0
        AND discount_percent BETWEEN 0 AND 90
        AND total_cents > 0
        AND total_cents = subtotal_cents - discount_cents
    ),
    CONSTRAINT payments_paid_at_present CHECK (status NOT IN ('PAID', 'REFUNDED', 'DISPUTED') OR paid_at IS NOT NULL),
    CONSTRAINT payments_refunded_at_present CHECK (status NOT IN ('REFUNDED', 'DISPUTED') OR refunded_at IS NOT NULL)
);

CREATE UNIQUE INDEX payments_one_open_checkout_per_user ON payments (user_id) WHERE status = 'OPEN';
CREATE INDEX payments_user_paid_at_index ON payments (user_id, paid_at DESC, id DESC) WHERE paid_at IS NOT NULL;
CREATE INDEX payments_plan_id_index ON payments (plan_id);

CREATE TABLE payment_chapters (
    payment_id UUID NOT NULL REFERENCES payments (id) ON DELETE CASCADE,
    chapter_id TEXT NOT NULL REFERENCES chapters (id) ON DELETE RESTRICT,
    chapter_number INTEGER NOT NULL,
    chapter_title TEXT NOT NULL,
    price_cents BIGINT NOT NULL CHECK (price_cents > 0),
    PRIMARY KEY (payment_id, chapter_id)
);

CREATE INDEX payment_chapters_chapter_id_index ON payment_chapters (chapter_id);
