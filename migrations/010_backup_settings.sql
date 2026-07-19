-- +goose Up
-- Настройки резервного копирования и общего S3-подключения (Cloudflare R2).
-- Секреты S3 (access_key_id, secret_access_key) шифруются приложением (enc:v1:).
INSERT INTO settings (key, value) VALUES
    ('BACKUP', '{"enabled":false,"type":"full","interval_hours":24,"keep":7,"storage":"local"}'),
    ('S3', '{"endpoint":"","region":"auto","bucket":"","access_key_id":"","secret_access_key":"","use_path_style":false}')
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM settings WHERE key IN ('BACKUP', 'S3');
