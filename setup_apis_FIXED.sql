-- CORRECTED Setup script untuk menambahkan API sources ke database
-- Updated: 2026-02-13 - Fixed self-referencing URLs

-- IMPORTANT: Use production URLs, NOT api-gateway (self-reference)

-- Clear old potentially wrong data
DELETE FROM api_sources WHERE base_url LIKE '%api-gateway%';

-- Add winbutv for anime category
INSERT OR IGNORE INTO api_sources (endpoint_id, source_name, base_url, priority, is_primary, is_active) 
SELECT e.id, 'winbutv', 'https://winbu-tv.humanmade.my.id', 1, 1, 1
FROM endpoints e
JOIN categories c ON e.category_id = c.id
WHERE c.name = 'anime';

-- Add winbutv_drakor for drakor category  
INSERT OR IGNORE INTO api_sources (endpoint_id, source_name, base_url, priority, is_primary, is_active)
SELECT e.id, 'winbutv_drakor', 'https://winbu-tv.humanmade.my.id', 1, 1, 1
FROM endpoints e
JOIN categories c ON e.category_id = c.id
WHERE c.name = 'drakor';

-- Verifikasi data
SELECT 'API Sources Summary:' as info;
SELECT 
    c.name as category,
    s.source_name,
    s.base_url,
    COUNT(*) as endpoint_count
FROM api_sources s
JOIN endpoints e ON s.endpoint_id = e.id
JOIN categories c ON e.category_id = c.id
GROUP BY c.name, s.source_name, s.base_url;

-- Check for self-referencing (should be 0)
SELECT 'Self-reference check (should be 0):' as check;
SELECT COUNT(*) as self_ref_count 
FROM api_sources 
WHERE base_url LIKE '%api-gateway%';
