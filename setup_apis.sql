-- Setup script untuk menambahkan API sources ke database
-- Berdasarkan struktur proyek yang ada

-- Update API sources dengan URL yang benar
UPDATE api_sources SET base_url = 'http://localhost:8081' WHERE source_name = 'multiplescrape';
UPDATE api_sources SET base_url = 'http://localhost:8082' WHERE source_name = 'winbutv';

-- Tambahkan API sources untuk semua endpoints
INSERT OR IGNORE INTO api_sources (endpoint_id, source_name, base_url, priority, is_primary) 
SELECT e.id, 'multiplescrape', 'http://localhost:8081', 1, 1
FROM endpoints e;

INSERT OR IGNORE INTO api_sources (endpoint_id, source_name, base_url, priority, is_primary) 
SELECT e.id, 'winbutv', 'http://localhost:8082', 2, 1
FROM endpoints e;

-- Tambahkan fallback APIs (contoh)
INSERT OR IGNORE INTO fallback_apis (api_source_id, fallback_url, priority) 
SELECT a.id, 'http://localhost:8083' || e.path, 1
FROM api_sources a
JOIN endpoints e ON a.endpoint_id = e.id
WHERE a.source_name = 'multiplescrape';

INSERT OR IGNORE INTO fallback_apis (api_source_id, fallback_url, priority) 
SELECT a.id, 'http://localhost:8084' || e.path, 1
FROM api_sources a
JOIN endpoints e ON a.endpoint_id = e.id
WHERE a.source_name = 'winbutv';

-- Verifikasi data
SELECT 'Categories:' as info;
SELECT * FROM categories;

SELECT 'Endpoints:' as info;
SELECT * FROM endpoints;

SELECT 'API Sources:' as info;
SELECT a.*, e.path as endpoint_path FROM api_sources a
JOIN endpoints e ON a.endpoint_id = e.id;

SELECT 'Fallback APIs:' as info;
SELECT f.*, a.source_name, e.path as endpoint_path FROM fallback_apis f
JOIN api_sources a ON f.api_source_id = a.id
JOIN endpoints e ON a.endpoint_id = e.id;