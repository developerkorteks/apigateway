# 🌐 **DYNAMIC SWAGGER CONFIGURATION**

**Feature**: Dynamic host detection for all swagger documentation  
**Benefit**: Automatic adaptation to any deployment domain  

---

## 🎯 **IMPLEMENTATION OVERVIEW**

### **Problem Solved:**
- ❌ **Before**: Hardcoded `localhost:8002` in swagger docs
- ✅ **After**: Dynamic host detection (`yourdomain.com`, `api.example.com`, etc.)

### **Benefits:**
- ✅ **Easy Deployment**: Works on any domain without code changes
- ✅ **HTTPS Support**: Automatic scheme detection (http/https)
- ✅ **Reverse Proxy Ready**: Supports X-Forwarded-Proto headers
- ✅ **Load Balancer Compatible**: Works with any deployment setup

---

## 🔧 **IMPLEMENTATION DETAILS**

### **1. FastAPI (Port 8000)**
```python
# Dynamic OpenAPI generation
def custom_openapi(request: Request):
    scheme = request.url.scheme
    host = request.headers.get("host", str(request.url.netloc))
    
    # Check forwarded headers
    forwarded_proto = request.headers.get("x-forwarded-proto")
    if forwarded_proto:
        scheme = forwarded_proto
    
    server_url = f"{scheme}://{host}"
    # Generate OpenAPI with dynamic server URL
```

### **2. WinbuTV (Port 8002)**  
```go
// Dynamic swagger.json endpoint
r.GET("/swagger/doc.json", func(c *gin.Context) {
    host := c.Request.Host
    scheme := "http"
    
    // HTTPS detection
    if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
        scheme = "https"
    }
    
    // Replace host dynamically in swagger JSON
    doc = strings.Replace(doc, `"host": "localhost:8002"`, `"host": "`+host+`"`, 1)
})
```

### **3. MultipleScrape (Port 8001)**
```go
// Enhanced dynamic host detection
func getDynamicSwaggerJSON(c *gin.Context) string {
    host := c.Request.Host
    scheme := "http"
    
    // Multiple HTTPS detection methods
    if c.Request.TLS != nil || 
       c.GetHeader("X-Forwarded-Proto") == "https" ||
       c.GetHeader("X-Forwarded-Ssl") == "on" {
        scheme = "https"
    }
    
    // Dynamic replacement in swagger JSON
    swaggerJSON = strings.Replace(swaggerJSON, `"host": "localhost:8001"`, `"host": "`+host+`"`, 1)
}
```

---

## 🌍 **DEPLOYMENT SCENARIOS**

### **Local Development:**
- `http://localhost:8000/docs` → Works ✅
- `http://localhost:8001/swagger/index.html` → Works ✅  
- `http://localhost:8002/swagger/index.html` → Works ✅

### **Production Domain:**
- `https://api.yourdomain.com/docs` → Auto-detects ✅
- `https://anime-api.example.com/swagger/index.html` → Auto-detects ✅
- `http://myserver.com:3000/swagger/` → Auto-detects ✅

### **Behind Reverse Proxy/Load Balancer:**
```nginx
# Nginx config example
location /api/ {
    proxy_pass http://backend;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Host $host;
}
```
**Result**: Swagger automatically detects `https://yourdomain.com` ✅

---

## 🚀 **TESTING COMMANDS**

### **Local Testing:**
```bash
# Test all swagger endpoints
curl http://localhost:8000/docs                    # FastAPI
curl http://localhost:8001/swagger/index.html      # MultipleScrape  
curl http://localhost:8002/swagger/index.html      # WinbuTV

# Test dynamic OpenAPI JSON
curl http://localhost:8000/api/v1/openapi.json | jq '.servers[0].url'
curl http://localhost:8001/swagger/doc.json | jq '.host'
curl http://localhost:8002/swagger/doc.json | jq '.host'
```

### **Production Testing:**
```bash
# Will show your actual domain
curl https://yourdomain.com/api/v1/openapi.json | jq '.servers[0].url'
# Expected: "https://yourdomain.com"
```

---

## ✅ **VERIFICATION CHECKLIST**

### **Development Environment:**
- [ ] FastAPI swagger shows `http://localhost:8000`
- [ ] MultipleScrape swagger shows `http://localhost:8001`  
- [ ] WinbuTV swagger shows `http://localhost:8002`

### **Production Environment:**
- [ ] All swagger UIs load with correct domain
- [ ] API calls work from swagger UI
- [ ] HTTPS detection working properly
- [ ] Reverse proxy headers recognized

---

## 🏆 **BENEFITS ACHIEVED**

### **For Developers:**
- ✅ **No Config Changes**: Deploy anywhere without modifying swagger settings
- ✅ **Easy Testing**: Swagger UI works immediately on any domain
- ✅ **HTTPS Ready**: Automatic secure protocol detection

### **For DevOps:**
- ✅ **Domain Agnostic**: Same code works on dev/staging/production  
- ✅ **Load Balancer Ready**: Proper header handling for proxy setups
- ✅ **Container Friendly**: No hardcoded URLs to change

### **For End Users:**
- ✅ **Always Accessible**: Swagger docs work on current domain
- ✅ **Secure**: HTTPS automatically detected and configured
- ✅ **Fast**: No manual configuration needed

---

**🎯 Result: Production-ready swagger documentation that adapts to any deployment environment!** 🌐