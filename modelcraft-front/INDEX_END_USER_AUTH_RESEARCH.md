# End-User Authentication Research - Complete Documentation Index

## 📚 Documentation Files

All files have been created in the `modelcraft-front/` directory:

### 1. **START HERE** → [RESEARCH_SUMMARY.md](RESEARCH_SUMMARY.md)
   - **What:** Executive summary of findings
   - **Read Time:** 5 minutes
   - **Contains:**
     - What we found (5 sections with ✅ checks)
     - The Problem (current broken state)
     - The Solution (pattern to implement)
     - Key Insights + Next Steps

### 2. [END_USER_AUTH_QUICK_REFERENCE.md](END_USER_AUTH_QUICK_REFERENCE.md)
   - **What:** Practical quick lookup guide
   - **Read Time:** 5 minutes
   - **Best For:** Implementation, debugging, testing
   - **Contains:**
     - File locations table
     - Storage layer details
     - Token refresh flow (code examples)
     - Apollo client patterns
     - Endpoints & error codes
     - Testing scenarios

### 3. [END_USER_AUTH_ANALYSIS.md](END_USER_AUTH_ANALYSIS.md)
   - **What:** Deep architectural analysis
   - **Read Time:** 15 minutes
   - **Best For:** Understanding system design
   - **Contains:**
     - Complete token storage breakdown
     - Refresh mechanism (BFF + client-side)
     - Apollo client integration patterns
     - Middleware protection details
     - Types & interfaces
     - Endpoint summary

### 4. [END_USER_AUTH_FLOW.txt](END_USER_AUTH_FLOW.txt)
   - **What:** Visual ASCII diagrams
   - **Read Time:** 10 minutes (skim as needed)
   - **Best For:** Visual learners
   - **Contains:**
     - 5-phase flow diagrams (login → refresh → graphql → logout)
     - Storage layer architecture diagrams
     - Concurrent refresh pattern with timeline
     - All in ASCII art (no tools needed)

### 5. [END_USER_AUTH_FILE_TREE.txt](END_USER_AUTH_FILE_TREE.txt)
   - **What:** File structure + dependency map
   - **Read Time:** 5 minutes
   - **Best For:** Navigation, finding files
   - **Contains:**
     - Complete directory tree with annotations
     - Data flow diagrams (ASCII)
     - Import dependency tree
     - Each file's purpose and key functions

---

## 🎯 Reading Path by Role

### If you're an **Implementer** (fixing the bug)
1. Read: RESEARCH_SUMMARY.md (THE PROBLEM section)
2. Ref: END_USER_AUTH_QUICK_REFERENCE.md (Apollo Client Pattern)
3. Ref: END_USER_AUTH_ANALYSIS.md (4. Apollo Client Integration)

### If you're a **Designer/Architect** (understanding the system)
1. Read: RESEARCH_SUMMARY.md (all sections)
2. Read: END_USER_AUTH_ANALYSIS.md (all)
3. Skim: END_USER_AUTH_FLOW.txt (visual overview)

### If you're **Debugging** (something is broken)
1. Ref: END_USER_AUTH_QUICK_REFERENCE.md (Error Handling section)
2. Ref: END_USER_AUTH_FLOW.txt (find your phase)
3. Ref: END_USER_AUTH_FILE_TREE.txt (locate files)

### If you're **Testing** (writing tests)
1. Ref: END_USER_AUTH_QUICK_REFERENCE.md (Testing Refresh Scenario)
2. Ref: END_USER_AUTH_FLOW.txt (Refresh Concurrency section)
3. Ref: Files:
   - `src/mocks/handlers/end-user/auth-handlers.ts` (existing test mocks)

### If you're **Learning** (first time here)
1. Start: RESEARCH_SUMMARY.md
2. Then: END_USER_AUTH_FLOW.txt (look at PHASE diagrams)
3. Deep: END_USER_AUTH_ANALYSIS.md (section 1 & 2)
4. Reference: Everything else as needed

---

## 🔍 Quick File Finder

**Looking for this?** → **Find it here:**

| What | File | Reference |
|------|------|-----------|
| Token storage implementation | `src/shared/stores/end-user-auth-store.ts` | ANALYSIS.md §1, QUICK_REF.md §1 |
| Refresh function | `src/api-client/end-user/end-user-auth-client.ts` | ANALYSIS.md §3, QUICK_REF.md §2 |
| Tenant pattern (reference) | `src/api-client/apollo/clients.ts` lines 46-74 | ANALYSIS.md §4, SUMMARY.md §3 |
| End-user Apollo clients (broken) | `src/api-client/apollo/clients.ts` lines 144-264 | ANALYSIS.md §4, SUMMARY.md THE PROBLEM |
| BFF refresh endpoint | `src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts` | ANALYSIS.md §2, SUMMARY.md §5 |
| Proxy helper | `src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts` | ANALYSIS.md §2, FILE_TREE.txt |
| Type definitions | `src/types/end-user-auth.ts` | ANALYSIS.md §6, QUICK_REF.md §8 |
| Route protection | `src/middleware.ts` | ANALYSIS.md §5, FILE_TREE.txt |
| Test mocks | `src/mocks/handlers/end-user/auth-handlers.ts` | FILE_TREE.txt |

---

## 💡 Key Concepts Explained

### Token Expiry Buffer
```
Actual Expiry: 60 minutes
Warning Window: -5 minutes
Refresh Triggered: At 55 minute mark
Reason: Prevents queries using expired token
```
See: QUICK_REF.md §2, FLOW.txt §PHASE 2

### Concurrency Protection
```
Multiple queries check expiry simultaneously
→ Multiple calls to refreshEndUserAccessToken()
→ BUT: Only 1 backend request via shared Promise
```
See: FLOW.txt §REFRESH CONCURRENCY PATTERN, ANALYSIS.md end

### HttpOnly Cookie Strategy
```
Why HttpOnly?
├─ XSS Protection: JavaScript can't read it
├─ CSRF Protection: SameSite=Strict prevents cross-site
├─ Automatic: Sent with credentials: include
└─ Rotation: Handled by backend on each refresh
```
See: QUICK_REF.md §1, FLOW.txt §STORAGE LAYER

### BFF Proxy Pattern
```
Browser → BFF (reads cookie from request)
       ↓ (extracts refreshToken from cookie)
       → Injects into request body
       → Proxies to Go backend
       ← Receives Set-Cookie (rotated token)
       → Transparently passes back to browser
```
See: ANALYSIS.md §2, FILE_TREE.txt §REFRESH

---

## 🚀 Implementation Checklist

To implement end-user token refresh in Apollo:

- [ ] Read: RESEARCH_SUMMARY.md § THE SOLUTION
- [ ] Ref: QUICK_REF.md § Apollo Client Pattern (both examples)
- [ ] Find: `src/api-client/apollo/clients.ts`
- [ ] Update: `createEndUserScopedClient()` (remove token param, add refresh logic)
- [ ] Update: `createEndUserOrgScopedClient()` (same pattern)
- [ ] Update: `createEndUserModelRuntimeClient()` (same pattern)
- [ ] Import: `getEndUserToken`, `refreshEndUserAccessToken` from end-user-auth-client.ts
- [ ] Import: `useEndUserAuthStore` for isTokenExpired() check
- [ ] Test: Manual token expiry scenario (use browser console)
- [ ] Verify: No queries fail after 1 hour

See: SUMMARY.md § Next Steps

---

## 🎓 Learning Materials

### Understand How Tokens Work
1. FLOW.txt § PHASE 1 (Login) → See how tokens first arrive
2. FLOW.txt § PHASE 2 (Silent Refresh) → See refresh flow
3. ANALYSIS.md § 1 (END-USER TOKEN STORAGE) → How they're stored

### Understand Apollo Integration
1. QUICK_REF.md § Apollo Client Pattern
2. ANALYSIS.md § 4 (APOLLO CLIENT INTEGRATION)
3. FILE_TREE.txt § IMPORT TREE → See what needs importing

### Understand Error Scenarios
1. QUICK_REF.md § Error Handling
2. FLOW.txt § PHASE 4 (Token Expiry During Operation)
3. Search for `INVALID_REFRESH_TOKEN` in ANALYSIS.md

### Understand Cookie Security
1. FLOW.txt § STORAGE LAYER ARCHITECTURE
2. QUICK_REF.md § Storage Layer § 2. HttpOnly Cookie
3. ANALYSIS.md § 2 (EXISTING REFRESH MECHANISM) § Cookie Handling

---

## 📋 File Checklist

All source files referenced in this documentation:

**Authentication Store:**
- [x] `src/shared/stores/end-user-auth-store.ts` (51 lines)

**Refresh Logic:**
- [x] `src/api-client/end-user/end-user-auth-client.ts` (174 lines)

**Apollo Clients:**
- [x] `src/api-client/apollo/clients.ts` (301 lines)

**BFF Routes:**
- [x] `src/app/api/bff/org/[orgName]/end-user/auth/refresh/route.ts` (44 lines)
- [x] `src/app/api/bff/org/[orgName]/end-user/auth/_proxy.ts` (100 lines)

**Types:**
- [x] `src/types/end-user-auth.ts` (176 lines)

**Middleware:**
- [x] `src/middleware.ts` (104 lines)

**Total: 950 lines of code analyzed**

---

## 🔗 Cross-References

**Find mentions of:**

| Concept | Found In |
|---------|----------|
| `useEndUserAuthStore` | ANALYSIS §1, QUICK_REF §1, FILE_TREE |
| `refreshEndUserAccessToken()` | ANALYSIS §3, QUICK_REF §2, SUMMARY §2, FLOW |
| `createEndUserScopedClient()` | ANALYSIS §4, QUICK_REF §3, SUMMARY §PROBLEM, FILE_TREE |
| `isTokenExpired()` | ANALYSIS §1, QUICK_REF §2, FLOW §PHASE 2 |
| `mc_enduser_refresh_token` | ANALYSIS §1, QUICK_REF §1, FILE_TREE, FLOW |
| BFF refresh route | ANALYSIS §2, SUMMARY §5, FILE_TREE, FLOW |
| Tenant pattern | SUMMARY §3, QUICK_REF §3, ANALYSIS §4 |
| X-Action header | ANALYSIS §4, FILE_TREE |
| Concurrency protection | ANALYSIS end, QUICK_REF §2, FLOW §REFRESH |

---

## 📞 Questions? Check Here

**Q: Where is the token stored?**
A: QUICK_REF §1, ANALYSIS §1, FLOW § STORAGE LAYER

**Q: Is there a refresh mechanism?**
A: YES! SUMMARY §2, ANALYSIS §2, QUICK_REF §2

**Q: What's broken with the current code?**
A: SUMMARY § THE PROBLEM, QUICK_REF §3 (Current State)

**Q: How do I fix it?**
A: SUMMARY § THE SOLUTION, IMPLEMENTATION CHECKLIST

**Q: What's the tenant pattern?**
A: SUMMARY §3, ANALYSIS §4, QUICK_REF §3 (Tenant Pattern)

**Q: How does refresh work?**
A: FLOW § PHASE 2, ANALYSIS §2, QUICK_REF §2

**Q: What endpoints exist?**
A: QUICK_REF § Endpoints, ANALYSIS § Endpoint Summary

**Q: How are cookies handled?**
A: FLOW § STORAGE LAYER, ANALYSIS §2 § Proxy Helper

**Q: What's the concurrency pattern?**
A: FLOW § REFRESH CONCURRENCY PATTERN, ANALYSIS end

---

## 📖 Document Statistics

| Document | Size | Read Time | Type |
|----------|------|-----------|------|
| RESEARCH_SUMMARY.md | 11.1 KB | 5 min | Executive Summary |
| END_USER_AUTH_ANALYSIS.md | 11.7 KB | 15 min | Deep Dive |
| END_USER_AUTH_QUICK_REFERENCE.md | 8.2 KB | 5 min | Quick Lookup |
| END_USER_AUTH_FLOW.txt | 27.8 KB | 10 min | Visual Diagrams |
| END_USER_AUTH_FILE_TREE.txt | 11.6 KB | 5 min | Navigation |
| **TOTAL** | **70.4 KB** | **~40 min** | Full Package |

**Optimal reading strategy:** 
- Executive: SUMMARY (5 min)
- Implementer: SUMMARY + QUICK_REF (10 min)
- Deep dive: ANALYSIS + FLOW (25 min)
- Complete: All (40 min)

---

## ✅ What You Now Understand

After reading these documents, you know:

- ✅ Tokens are stored in Zustand store + HttpOnly cookie (dual storage)
- ✅ Refresh mechanism exists: BFF endpoint + client-side function
- ✅ Concurrency is protected (shared promise pattern)
- ✅ Tenant pattern is the template to replicate
- ✅ Apollo clients need fixing (3 functions)
- ✅ HttpOnly cookies handle rotation securely
- ✅ 5-minute warning buffer prevents expiry race conditions
- ✅ All infrastructure is already in place (just needs Apollo integration)
- ✅ Next steps are clear (1 file to modify)

**Ready to implement!** 🚀

