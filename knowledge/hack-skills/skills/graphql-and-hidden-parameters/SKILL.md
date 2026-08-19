---
name: graphql-and-hidden-parameters
description: >-
  GraphQL and hidden parameter testing playbook. Use when exploring introspection, batching, undocumented fields, hidden parameters, schema abuse, and GraphQL authorization gaps.
---

# SKILL: GraphQL and Hidden Parameters — Introspection, Batching, and Undocumented Fields

> **AI LOAD INSTRUCTION**: Use this skill when GraphQL exists or when REST documentation suggests optional, deprecated, or undocumented fields. Focus on schema discovery, hidden parameter abuse, and batching as a force multiplier.

## 1. GRAPHQL FIRST PASS

```graphql
query { __typename }
query {
  __schema {
    types { name }
  }
}
```

If introspection is restricted, continue with:

- field suggestions and error-based discovery
- known type probes like `__type(name: "User")`
- JS and mobile bundle route extraction

## 2. HIGH-VALUE GRAPHQL TESTS

| Theme | Example |
|---|---|
| IDOR | `user(id: "victim")` |
| batching | array of login or object fetch operations |
| hidden fields | admin-only fields exposed in type definitions |
| nested authz gaps | related object fields with weaker checks |

## 3. HIDDEN PARAMETER DISCOVERY

Look for:

- fields present in admin docs but not public docs
- `additionalProperties` or permissive schemas
- frontend code using richer request bodies than visible UI controls
- mobile endpoints carrying role, org, feature-flag, or internal filter fields

## 4. NEXT ROUTING

- If hidden fields affect privilege: [api authorization and bola](../api-authorization-and-bola/SKILL.md)
- If GraphQL batching changes auth or rate behavior: [api auth and jwt abuse](../api-auth-and-jwt-abuse/SKILL.md)
- If endpoint discovery is incomplete: [api recon and docs](../api-recon-and-docs/SKILL.md)