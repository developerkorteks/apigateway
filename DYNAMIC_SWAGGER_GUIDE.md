# 🎯 Dynamic Swagger Categories Guide

Sistem ini membuat dropdown category di Swagger UI menjadi **100% dinamis** berdasarkan kategori yang ditambahkan via dashboard!

## 🚀 Fitur Utama

✅ **Auto-Update Dropdown**: Category dropdown di Swagger otomatis update  
✅ **Real-time Sync**: Sinkronisasi langsung dengan database  
✅ **Multi-Access**: Tersedia di 2 endpoint Swagger  
✅ **Auto-Refresh**: Refresh otomatis setiap 5 menit  
✅ **Fallback Safe**: Fallback ke kategori default jika error  

## 📍 Endpoint Swagger

### 1. Standard Swagger UI
```
http://localhost:8080/swagger/index.html
```
- Swagger UI standar dengan JavaScript injection
- Category dropdown otomatis terupdate

### 2. Custom Dynamic Swagger UI  
```
http://localhost:8080/swagger-ui
```
- Custom UI dengan visual indicator kategori aktif
- Menampilkan kategori yang loaded
- Real-time category updates

## 🎛️ Cara Kerja Sistem

### 1. Database Integration
```sql
-- Sistem membaca dari tabel categories
SELECT name FROM categories WHERE is_active = TRUE ORDER BY name
```

### 2. API Endpoint
```bash
# Endpoint untuk mendapatkan kategori dinamis
GET /api/categories/names

# Response:
{
  "status": "success", 
  "data": ["anime", "korean-drama", "donghua", "film", "all"]
}
```

### 3. JavaScript Auto-Injection
```javascript
// Auto-update dropdown saat Swagger load
window.DYNAMIC_CATEGORIES = ["anime", "korean-drama", "donghua", "film", "all"];

// Update semua dropdown category
updateCategoryDropdowns();
```

## 🎯 Demo Workflow

### Step 1: Tambah Kategori via Dashboard
```bash
# Via dashboard atau API
POST /dashboard/categories
{
  "name": "donghua",
  "description": "Chinese Animation"
}
```

### Step 2: Kategori Otomatis Muncul di Swagger
- Buka: `http://localhost:8080/swagger-ui`
- Lihat dropdown category di endpoint `/search`
- **Donghua** otomatis muncul dalam dropdown! 🎉

### Step 3: Test API dengan Kategori Baru
```bash
# Test search dengan kategori donghua
curl "http://localhost:8080/api/v1/search?q=demon&category=donghua"
```

## 📊 Visual Comparison

### ❌ Before (Static)
```
Category Dropdown:
┌─────────────────┐
│ anime           │
│ korean-drama    │  
│ all             │
└─────────────────┘
```

### ✅ After (Dynamic)
```
Category Dropdown:
┌─────────────────┐
│ anime           │
│ donghua         │ ← NEW!
│ film            │ ← NEW!
│ korean-drama    │
│ manhwa          │ ← NEW!
│ all             │
└─────────────────┘
```

## 🔧 Technical Implementation

### 1. Database Layer
```go
// pkg/database/database.go
func (db *DB) GetCategoryNames() ([]string, error) {
    // Query active categories from database
    rows, err := db.Query("SELECT name FROM categories WHERE is_active = TRUE ORDER BY name")
    // ... return category names + "all"
}
```

### 2. Service Layer  
```go
// internal/service/api_service.go
func (s *APIService) GetCategoryNames() ([]string, error) {
    return s.db.GetCategoryNames()
}
```

### 3. Handler Layer
```go
// internal/api/handlers/dashboard_handler.go
func (h *DashboardHandler) GetCategoryNames(c *gin.Context) {
    categoryNames, err := h.apiService.GetCategoryNames()
    // Return JSON response
}
```

### 4. Custom Swagger Handler
```go
// internal/api/handlers/swagger_handler.go
func (h *SwaggerHandler) ServeSwaggerUI(c *gin.Context) {
    categories, err := h.apiService.GetCategoryNames()
    // Inject categories into custom HTML template
}
```

### 5. Frontend JavaScript
```javascript
// Auto-update category dropdowns
function updateCategoryDropdowns() {
    const categorySelects = document.querySelectorAll('select[data-param-name="category"]');
    categorySelects.forEach(select => {
        // Clear and repopulate with dynamic categories
    });
}
```

## 🎨 Custom Swagger UI Features

### Visual Category Indicator
```html
<div class="dynamic-category-info">
    <strong>🎯 Dynamic Categories Loaded:</strong>
    <div class="category-list">
        <span class="category-tag">anime</span>
        <span class="category-tag">donghua</span>
        <span class="category-tag">film</span>
        <span class="category-tag">korean-drama</span>
        <span class="category-tag">all</span>
    </div>
</div>
```

### Auto-Refresh Mechanism
```javascript
// Auto-refresh every 5 minutes
setInterval(async () => {
    const response = await fetch('/api/categories/names');
    const data = await response.json();
    if (data.status === 'success') {
        window.DYNAMIC_CATEGORIES = data.data;
        updateCategoryDropdowns();
    }
}, 5 * 60 * 1000);
```

## 🧪 Testing Scenarios

### Test 1: Add New Category
```bash
# 1. Add category via dashboard
curl -X POST http://localhost:8080/dashboard/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"donghua","description":"Chinese Animation"}'

# 2. Check if appears in API
curl http://localhost:8080/api/categories/names

# 3. Refresh Swagger UI - should see "donghua" in dropdown
```

### Test 2: Real-time Update
```bash
# 1. Open Swagger UI in browser
# 2. Add category via dashboard in another tab  
# 3. Wait max 5 minutes or refresh page
# 4. New category appears in dropdown automatically
```

### Test 3: API Usage with New Category
```bash
# Use new category in API calls
curl "http://localhost:8080/api/v1/search?q=test&category=donghua"
curl "http://localhost:8080/api/v1/anime-terbaru?category=film"
```

## 🎯 Benefits Achieved

### For Developers
✅ **No Code Changes**: Add categories without touching Swagger annotations  
✅ **Auto-Documentation**: Swagger always reflects current system state  
✅ **Type Safety**: Categories validated against database  

### For Users  
✅ **Intuitive UI**: Dropdown shows exactly what's available  
✅ **Real-time Updates**: No need to restart or redeploy  
✅ **Visual Feedback**: See active categories at a glance  

### For Operations
✅ **Zero Downtime**: Add categories without service restart  
✅ **Consistent State**: UI always matches backend capabilities  
✅ **Easy Scaling**: Support unlimited content types  

## 🚀 Advanced Usage

### Bulk Category Addition
```bash
# Add multiple categories at once
curl -X POST http://localhost:8080/dashboard/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"donghua","description":"Chinese Animation"}'

curl -X POST http://localhost:8080/dashboard/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"manhwa","description":"Korean Webtoons"}'

curl -X POST http://localhost:8080/dashboard/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"film","description":"Movies"}'
```

### Category Management
```bash
# List all categories
curl http://localhost:8080/dashboard/categories

# Update category
curl -X PUT http://localhost:8080/dashboard/categories/3 \
  -H "Content-Type: application/json" \
  -d '{"name":"donghua","description":"Chinese Animation & Donghua"}'

# Deactivate category (removes from dropdown)
curl -X DELETE http://localhost:8080/dashboard/categories/3
```

## 🎉 Result

**Sekarang dropdown category di Swagger 100% dinamis!**

- ✅ Tambah kategori via dashboard → Langsung muncul di Swagger
- ✅ Support unlimited content types (anime, drakor, film, donghua, manhwa, dll)
- ✅ Real-time updates tanpa restart aplikasi
- ✅ Visual feedback untuk user
- ✅ Backward compatible dengan sistem lama

**Perfect untuk scaling ke berbagai jenis content!** 🚀