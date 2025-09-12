CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    channel VARCHAR(100) NOT NULL,       -- misal: "admin-chat", "room-123"
    user_id VARCHAR(50) NOT NULL,        -- ID user (bisa UUID atau string dari app)
    message TEXT NOT NULL,               -- isi pesan
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now() -- kapan pesan dikirim
);

CREATE INDEX idx_messages_channel_created_at
ON messages (channel, created_at);
