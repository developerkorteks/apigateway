-- Fix database configuration untuk mengatasi masalah duplikasi konten
-- Masalah: gomunime dan winbutv menggunakan URL yang sama (localhost:8002)

-- Update URL yang benar untuk setiap source
UPDATE api_sources SET base_url = 'http://localhost:8001' WHERE source_name = 'gomunime';
UPDATE api_sources SET base_url = 'http://localhost:8002' WHERE source_name = 'winbutv';
UPDATE api_sources SET base_url = 'http://128.199.109.211:8182' WHERE source_name = 'samehadaku';

-- Verifikasi perubahan
SELECT 'Updated API Sources for anime-terbaru:' as info;
SELECT a.id, a.source_name, a.base_url, a.priority, a.is_primary, a.is_active 
FROM api_sources a 
JOIN endpoints e ON a.endpoint_id = e.id 
WHERE e.path = '/api/v1/anime-terbaru' 
ORDER BY a.priority, a.id;

-- Cek semua sources
SELECT 'All API Sources:' as info;
SELECT a.id, a.source_name, a.base_url, a.priority, a.is_primary, a.is_active, e.path as endpoint_path
FROM api_sources a 
JOIN endpoints e ON a.endpoint_id = e.id 
ORDER BY e.path, a.priority, a.id;