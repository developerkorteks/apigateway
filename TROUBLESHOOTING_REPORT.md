# 🔍 Troubleshooting Report: Anime-Detail Endpoint

**Date:** 2026-02-13  
**Issue:** `/api/v1/anime-detail` endpoint returning 503 error  
**Status:** ✅ **RESOLVED**

---

## 📋 Problem Summary

Endpoint `/api/v1/anime-detail` gagal mengambil data dari API eksternal, selalu return:
- HTTP 503 Service Unavailable
- Error: "all API sources failed"
- Time: ~2 seconds (timeout)

---

## 🔬 Troubleshooting Process

### Testing Steps Performed:
1. ✅ Verified external API works (winbu-tv.humanmade.my.id)
2. ✅ Checked database configuration
3. ✅ Added debug logging to track request flow
4. ✅ Tested URL building logic
5. ✅ Identified timeout issue
6. ✅ Fixed and verified

---

## 🐛 Root Causes Found

### **Issue #1: Incorrect Base URL in Database**

**Problem:**  
Database stored self-referencing URL causing infinite loop/timeout

```
❌ WRONG: https://api-gateway.humanmade.my.id
✅ CORRECT: https://winbu-tv.humanmade.my.id
```

**Impact:**  
Gateway called itself instead of external API → timeout after 2s

**Fix:**
```sql
UPDATE api_sources 
SET base_url = 'https://winbu-tv.humanmade.my.id'
WHERE base_url = 'https://api-gateway.humanmade.my.id';
```

---

### **Issue #2: Timeout Too Short**

**Problem:**  
Bruteforce timeout = 2 seconds per source  
External API needs 3-5 seconds to respond

```go
// BEFORE (TOO SHORT):
case <-time.After(time.Duration(len(allSources)) * time.Second * 2):

// AFTER (FIXED):
case <-time.After(10 * time.Second): // Fixed 10 second timeout
```

**Impact:**  
Even with correct URL, timeout killed request before response arrived

**Fix:**  
Changed to fixed 10-second timeout in `internal/service/api_service.go`

---

## ✅ Solution Implemented

### 1. Database Update
- Updated all 7 api_sources entries (ID 90-96)
- Changed from api-gateway to winbu-tv URL
- Verified with SQL queries

### 2. Code Changes
- File: `internal/service/api_service.go`
- Function: `bruteforceDetailSources()`
- Line: ~1680
- Change: Timeout from dynamic (2s × sources) to fixed 10s

### 3. Testing Results
```
✅ HTTP Status: 200 OK
✅ Response Time: 3.65s
✅ Data: Complete anime details received
✅ Title: "Android wa Keiken Ninzuu ni Hairimasu ka??"
```

---

## 📊 Before vs After

| Metric | Before | After |
|--------|--------|-------|
| HTTP Status | 503 | **200** ✅ |
| Response Time | 2s (timeout) | 3.65s (success) ✅ |
| Data Received | null | **Full anime data** ✅ |
| Base URL | api-gateway (wrong) | **winbu-tv** ✅ |
| Timeout | 2s | **10s** ✅ |

---

## 🎯 Lessons Learned

1. **Always check database configuration first** - Wrong URL was root cause #1
2. **Timeout values matter** - API response time must be considered
3. **Add comprehensive logging** - Debug logs helped identify issues quickly
4. **Test manually first** - curl tests confirmed external API works
5. **Step-by-step troubleshooting** - Systematic approach found both issues

---

## 🔧 Files Modified

1. `data.db` - Updated api_sources table
2. `internal/service/api_service.go` - Increased timeout

---

## ✅ Verification

Test command:
```bash
curl "http://localhost:58080/api/v1/anime-detail?anime_slug=android-wa-keiken-ninzuu-ni-hairimasu-ka&category=anime"
```

Expected: 200 OK with anime details  
**Result: ✅ WORKING**

---

**Report Generated:** 2026-02-13  
**Total Troubleshooting Time:** ~12 iterations  
**Final Status:** ✅ **RESOLVED**
