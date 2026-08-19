---
name: saml-sso-assertion-attacks
description: >-
  SAML SSO assertion attack playbook. Use when testing signature validation, assertion wrapping, audience restrictions, ACS handling, XML trust boundaries, and enterprise SSO flaws.
---

# SKILL: SAML SSO and Assertion Attacks — Signature Validation, Binding, and Trust Confusion

> **AI LOAD INSTRUCTION**: Use this skill when the target uses SAML-based SSO and you need to validate assertion trust: signature coverage, audience and recipient checks, ACS handling, XML parsing weaknesses, and IdP/SP confusion.

## 1. WHEN TO LOAD THIS SKILL

Load when:

- Enterprise SSO uses SAML requests or responses
- You see `SAMLRequest`, `SAMLResponse`, XML assertions, or ACS endpoints
- Login flows involve an external IdP and browser POST/redirect binding

## 2. HIGH-VALUE MISCONFIGURATION CHECKS

| Theme | What to Check |
|---|---|
| signature validation | unsigned assertion accepted, wrong node signed, signature wrapping |
| audience and recipient | weak `Audience`, `Recipient`, `Destination`, or ACS validation |
| issuer trust | wrong IdP accepted or multi-tenant issuer confusion |
| replay and freshness | missing `InResponseTo`, weak `NotBefore` / `NotOnOrAfter` enforcement |
| account mapping | email-only binding, case folding, unverified attributes |
| XML parser behavior | XXE-like parser issues or unsafe transforms around SAML documents |

## 3. QUICK TRIAGE

1. Capture one full login round trip.
2. Inspect which XML nodes are signed and which attributes drive account binding.
3. Compare SP-initiated and IdP-initiated flows.
4. Test replay, altered attributes, and assertion placement confusion.

## 4. RELATED ROUTES

- XML parser attack depth: [xxe xml external entity](../xxe-xml-external-entity/SKILL.md)
- OAuth or OIDC SSO alternatives: [oauth oidc misconfiguration](../oauth-oidc-misconfiguration/SKILL.md)
- Auth boundary issues after SSO: [authbypass authentication flaws](../authbypass-authentication-flaws/SKILL.md)