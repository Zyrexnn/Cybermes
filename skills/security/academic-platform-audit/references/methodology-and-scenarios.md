# Academic Platform Security: Methodology & Audit Scenarios

This reference document provides standardized, anonymized testing scenarios illustrating how to assess examination and CBT platforms without impacting production availability or sensitive student data.

---

## Scenario A: Premature Question & Key Disclosure in Pre-fetched JSON

### Context & Symptom
Modern Single Page Applications (SPAs) frequently pre-fetch test batteries during the loading screen to optimize client responsiveness during offline or low-connectivity test periods. 

### Vulnerable Design Pattern
The backend endpoint `GET /api/v1/exams/{id}/questions` returns an array of question objects where each object directly embeds the correct answer flag or explanation:
```json
{
  "exam_id": 4021,
  "questions": [
    {
      "id": 101,
      "text": "What is the primary function of an operating system kernel?",
      "options": [
        {"id": "A", "text": "Rendering user interfaces", "is_correct": false},
        {"id": "B", "text": "Managing system resources and hardware", "is_correct": true},
        {"id": "C", "text": "Compiling source code", "is_correct": false}
      ],
      "solution_notes": "The kernel is the core component managing hardware interactions."
    }
  ]
}
```

### Diagnostic Procedure
1. Intercept network responses in the browser developer tools (Network tab) or through a test script:
   ```bash
   python3 scripts/audit_exam_state_machine.py --url https://exam.example.edu/api/v1/exams/4021/questions --token <TEST_SESSION_TOKEN>
   ```
2. Check if keys matching `is_correct`, `solution_notes`, or `kunci_jawaban` appear in the response while the exam session is in status `IN_PROGRESS`.

### Secure Implementation (Defense)
Use separate Data Transfer Objects (DTOs) for active testing vs post-exam review:
```python
# Secure Active Question DTO
class ActiveQuestionResponse(BaseModel):
    id: int
    text: str
    options: List[OptionDisplayDTO]  # Only contains id and text, NO is_correct
```

---

## Scenario B: Client-Side Timer Manipulation vs Server-Side Enforcement

### Context & Symptom
In some client-centric designs, the remaining time is tracked solely in a frontend JavaScript state (e.g. `state.remainingSeconds`). When the user finishes, the client sends:
```json
POST /api/v1/exams/4021/submit
{
  "answers": {"101": "B"},
  "time_spent_seconds": 1800,
  "client_status": "COMPLETED_ON_TIME"
}
```

### Diagnostic Procedure
1. Record the official test end time according to the server schedule.
2. After the official end time has elapsed by more than 5 minutes, transmit a test answer submission using curl or Postman.
3. Observe if the server responds with:
   - `400 Bad Request` or `409 Conflict` (Exam session expired) -> **SECURE**.
   - `200 OK` (Accepted and scored) -> **VULNERABLE (Late Submission Accepted)**.

### Secure Implementation (Defense)
Enforce time boundaries strictly on the database/backend:
```sql
UPDATE exam_sessions 
SET status = 'SUBMITTED', submitted_at = NOW() 
WHERE id = :session_id 
  AND user_id = :auth_user_id 
  AND status = 'IN_PROGRESS' 
  AND NOW() <= (started_at + INTERVAL '60 minutes');
```
If 0 rows are updated, reject the submission as expired.
