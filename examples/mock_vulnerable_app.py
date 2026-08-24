"""
Cybermes Mock Vulnerable Web Target
Used for safe local validation of IDOR, information disclosure, and reporting pipelines.
"""

from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import urllib.parse

DOCUMENTS = {
    "101": {"id": "101", "owner": "researcher_user_a", "title": "User A Public Notes", "content": "Public draft"},
    "102": {"id": "102", "owner": "researcher_user_b", "title": "User B Confidential Report", "content": "CONFIDENTIAL_FINANCIAL_DATA_FLAG{IDOR_VALIDATED_SUCCESSFULLY}"}
}

class VulnerableHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        pass

    def do_GET(self):
        parsed = urllib.parse.urlparse(self.path)
        path = parsed.path
        query = urllib.parse.parse_qs(parsed.query)

        if path == "/":
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({
                "service": "Cybermes Mock Target API",
                "version": "1.0.0",
                "status": "healthy",
                "endpoints": [
                    "/api/health",
                    "/api/documents/<id>",
                    "/search?q=<keyword>"
                ]
            }).encode())

        elif path == "/api/health":
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(b'{"status": "ok"}')

        elif path.startswith("/api/documents/"):
            doc_id = path.split("/")[-1]
            if doc_id in DOCUMENTS:
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(DOCUMENTS[doc_id]).encode())
            else:
                self.send_response(404)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"error": "Document not found"}')

        elif path == "/search":
            q = query.get("q", [""])[0]
            self.send_response(200)
            self.send_header("Content-Type", "text/html")
            self.end_headers()
            self.wfile.write(f"<html><body><h1>Search Results for: {q}</h1><p>No results found.</p></body></html>".encode())

        else:
            self.send_response(404)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(b'{"error": "Not Found"}')

if __name__ == "__main__":
    server = HTTPServer(("127.0.0.1", 8888), VulnerableHandler)
    print("Mock Vulnerable Target listening on http://127.0.0.1:8888")
    server.serve_forever()
