---
name: api-testing
description: Evaluates REST, GraphQL, and JSON-RPC APIs for authorization issues (BOLA/IDOR), schema exposure, and business logic flaws.
---

# API Testing Skill

## Purpose
Analyze candidate API endpoints identified during web enumeration for API security risks (OWASP API Security Top 10).

## Workflow
1. **Catalog API Endpoints**:
   Aggregate candidate endpoints from `/workspace/recon/<target>/api_candidates.txt`.
2. **Schema & Documentation Discovery**:
   Inspect common API documentation paths:
   - `/swagger.json`, `/openapi.json`, `/api-docs`
   - `/graphql` (check schema introspection query)
   - `/v1/`, `/v2/` versioning discrepancies
3. **Analyze API Attack Surfaces**:
   - **Broken Object Level Authorization (BOLA/IDOR)**: Inspect numeric or GUID parameters in URI paths (`/api/users/101`).
   - **Broken Authentication**: Test missing `Authorization` or `Bearer` tokens on sensitive resources.
   - **Mass Assignment**: Check if object update payloads accept unexpected attributes (e.g., `role: "admin"`, `is_verified: true`).
   - **Excessive Data Exposure**: Compare frontend display vs full raw API JSON response objects.
4. **Preserve Request/Response Logs**:
   Store all curl commands and raw JSON responses in `/workspace/output/api_tests/<target>/`.

## Safety Guidelines
- Do not perform automated volumetric fuzzing without explicit scope authorization.
- Test authorization boundaries using two distinct test accounts (Account A vs Account B), never modifying target data without authorization.
