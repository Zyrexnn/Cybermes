---
name: business-logic-vuln
description: >-
  Entry P1 category router for business logic testing. Use when workflow abuse,
  race conditions, pricing flaws, or multi-step state attacks matter more than
  parser-level input injection.
---

# Business Logic Router

This is the routing entry point for business-logic and state-machine issues.

## When to Use

- The target involves coupons, inventory, payment, approvals, quotas, invites, trials, or state transitions
- The issue is not parser-level; it is about when checks happen and which business conditions are checked
- You suspect race conditions, workflow bypass, price tampering, negative values, stacked discounts, or multi-step flaws

## Skill Map

- [Business Logic Vulnerabilities](../business-logic-vulnerabilities/SKILL.md)

## Recommended Flow

1. First map key business states and one-time actions
2. Then check for check-then-act windows, sequence dependencies, or missing cross-step authorization
3. If the chain depends on APIs, uploads, or object permissions, return to the corresponding router skill to complete the path

## Related Categories

- [api-sec](../api-sec/SKILL.md)
- [auth-sec](../auth-sec/SKILL.md)
- [file-access-vuln](../file-access-vuln/SKILL.md)