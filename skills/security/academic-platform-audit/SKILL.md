---
name: academic-platform-audit
description: Security assessment methodology for academic portals, Computer-Based Testing (CBT), and online examination platforms. Focuses on OWASP WSTG Business Logic testing, Question Bank information disclosure, server-side timer integrity, IDOR on submissions, and defensive mitigations.
---

# Academic & Examination Platform Security Assessment Playbook

## 1. Overview & Scope
This skill provides a structured methodology for auditing Computer-Based Testing (CBT), Learning Management Systems (LMS), and online examination portals. These applications frequently manage high-stakes evaluation data, student records, and timed workflows that depend heavily on state machine integrity.

### Primary Vulnerability Classes (OWASP WSTG)
- **WSTG-ATHZ-04**: Insecure Direct Object References (IDOR / BOLA) on question banks and student answers.
- **WSTG-INFO-05**: Sensitive Data Exposure in Single Page Application (SPA) client bundles and un-rendered JSON states.
- **WSTG-BUSL-01/02**: Business Logic Flaws & State Machine bypasses in multi-stage exam workflows.
- **WSTG-BUSL-08**: Client-Side Timing & Clock Manipulation vs Server-Side Epoch Timestamp enforcement.

---

## 2. Assessment Methodology

### Phase 1: Passive Reconnaissance & State Analysis
1. **Frontend Architecture Fingerprinting**:
   - Identify whether the portal is a Single Page Application (React, Vue, Next.js, Angular) or traditional server-rendered template.
   - Inspect JavaScript bundles (`/static/js/*.js` or `_next/static/chunks/*.js`) for unminified state models or route schemas.
2. **Payload & Model Review**:
   - Check if exam questions fetched from `/api/v1/session/{id}` or GraphQL endpoints include premature fields (`correct_option`, `answer_key`, `explanation`, `solution`).
   - Run the Cybermes diagnostic tool:
     ```bash
     python3 scripts/audit_exam_state_machine.py --url https://exam-portal.example.edu/api/v1/session/101 --token <SESSION_TOKEN>
     ```

### Phase 2: State Machine & Workflow Verification (WSTG-BUSL-02)
Standard online exam lifecycle:
`[Registration] -> [Verification] -> [Exam Start] -> [Answering] -> [Submission] -> [Scoring]`

1. **Step-Skipping Tests**:
   - Attempt direct POST/GET to `/api/v1/exam/{id}/submit` before invoking `/api/v1/exam/{id}/start`.
   - Attempt to access results endpoint `/api/v1/exam/{id}/review` while the exam status is still `ACTIVE`.
2. **Re-submission & Concurrency (Race Conditions)**:
   - Test if sending simultaneous answer submission requests allows overwriting final scores or duplicate grading executions.

### Phase 3: Timer & Deadline Integrity (WSTG-BUSL-08)
1. **Client Clock vs Server Timestamp**:
   - Check if the timer is enforced client-side (`remaining_time` decremented in JavaScript `setInterval`) or tracked server-side with an immutable `deadline_epoch`.
   - Tamper with local machine time or intercept submission payloads modifying `duration` or `client_submitted_at` parameters.
2. **Late-Submission Acceptance**:
   - Submit an answer payload 5-10 minutes after the scheduled test window expires.
   - Confirm whether the server strictly rejects late entries with HTTP 403 / 409 or silently accepts them.

### Phase 4: Access Control & IDOR Testing (WSTG-ATHZ-04)
1. **Cross-Account Session Isolation**:
   - Setup two authorized test accounts (Student A and Student B).
   - In Student A's active session, replace `session_id` or `student_id` with Student B's identifier in `POST /api/v1/exam/answer`.
   - Verify that the server validates session ownership against the authenticated JWT / session cookie and returns 403 Forbidden.

---

## 3. Remediation & Defensive Hardening
1. **Server-Side Authorization & Ownership**:
   - Bind every question-fetch and answer-submission strictly to the authenticated user ID extracted from server-validated tokens.
   - Never accept `user_id` or `student_id` from client request bodies.
2. **Data Model Stripping**:
   - Separate question entities from answer keys at the database/DTO layer.
   - Never serialize `correct_answer` or `explanation` into the client-facing question delivery API until the exam period has formally concluded.
3. **Immutable Server-Side Timer**:
   - Store `started_at` and calculate `expires_at = started_at + duration` on the server database.
   - When a submission arrives, compare `current_server_time <= expires_at + grace_period` (e.g. 5 seconds for network latency). Discard late submissions automatically.
