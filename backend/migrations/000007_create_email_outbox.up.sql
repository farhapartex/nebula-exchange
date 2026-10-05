CREATE TYPE email_status AS ENUM ('PENDING', 'SENT', 'FAILED');

CREATE TABLE email_outbox (
    id UUID PRIMARY KEY,
    template TEXT NOT NULL,
    recipient_email CITEXT NOT NULL,
    recipient_name TEXT NOT NULL DEFAULT '',
    subject TEXT NOT NULL,
    html_body TEXT NOT NULL,
    text_body TEXT NOT NULL,
    status email_status NOT NULL DEFAULT 'PENDING',
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 10,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    claimed_until TIMESTAMPTZ,
    last_error TEXT,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT email_outbox_attempts_not_negative CHECK (attempts >= 0),
    CONSTRAINT email_outbox_max_attempts_positive CHECK (max_attempts > 0),
    CONSTRAINT email_outbox_sent_has_time CHECK ((status = 'SENT') = (sent_at IS NOT NULL))
);

CREATE INDEX email_outbox_due_index ON email_outbox (next_attempt_at) WHERE status = 'PENDING';
