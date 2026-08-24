# 🤝 Contributing to Cybermes

Thank you for your interest in contributing to **Cybermes**! 

To maintain high code quality, security standards, and reproducibility across autonomous offensive research workflows, all contributions must go through our standard Pull Request (PR) and code review process.

---

## 📜 Core Contribution Rules

1. **Mandatory Review Gate & Branching Strategy**:
   - Direct pushes to `main` and `dev` are restricted.
   - **`dev` branch**: Primary development branch. All Pull Requests (PRs) should be branched from and targeted against `dev`.
   - **`main` branch**: Production-ready, stable releases only. `dev` is merged into `main` after complete testing and tagging.
   - Every PR requires review and explicit approval by the repository maintainer ([@Zyrexnn](https://github.com/Zyrexnn)) before merging.

2. **Zero Sensitive Data / Sanitization Standard**:
   - **NEVER** include real target domains, live IP addresses, production credentials, internal tokens, or actual client assessment data.
   - Always use RFC 2606 (`example.com`, `target.example.org`) and RFC 5737 (`198.51.100.x`, `203.0.113.x`) documentation standards for test cases and playbooks.

3. **Conventional Commits**:
   - Use industry-standard commit conventions:
     - `feat(...)`: New tools, skills, or framework capabilities
     - `fix(...)`: Bug fixes
     - `docs(...)`: Documentation or skill reference updates
     - `refactor(...)`: Code refactoring without behavioral changes
     - `chore(...)`: Maintenance, dependency, or configuration updates

4. **Non-Destructive Testing Principle**:
   - PoCs, scripts, and modules must follow safe, rate-controlled, and deterministic testing principles.

---

## 🛠️ Step-by-Step Workflow

### 1. Fork and Clone
```bash
# Fork the repository on GitHub, then clone your fork:
git clone https://github.com/<YOUR_USERNAME>/Cybermes.git
cd Cybermes
```

### 2. Create a Feature Branch from `dev`
```bash
# Ensure you are based on the latest dev branch
git checkout dev
git pull origin dev

# Create your feature/fix branch
git checkout -b feat/your-feature-name
# or
git checkout -b fix/your-fix-name
```

### 3. Make and Verify Changes
* Run Go test suite for core packages:
  ```bash
  go test ./pkg/... -v
  ```
* Ensure all Go CLI tools compile cleanly:
  ```bash
  go build ./cmd/...
  ```
* Ensure Python scripts compile and pass linting:
  ```bash
  python3 -m py_compile tools/<your_script>.py
  ```
* On Windows, test compatibility:
  ```powershell
  python tools\windows_compat_check.py
  ```
* Ensure tools are self-contained and document their CLI arguments clearly.

### 4. Commit and Push
```bash
git add .
git commit -m "feat(tools): add new reconnaissance module"
git push origin feat/your-feature-name
```

### 5. Open a Pull Request
* Open a PR against the **`dev`** branch of `Zyrexnn/Cybermes`.
* Fill out the automated PR template completely.
* You are welcome to add yourself to [`CONTRIBUTORS.md`](CONTRIBUTORS.md) as part of your PR.
* Wait for maintainer review. Respond to feedback or requested changes promptly.

---

## 👥 Recognition
All accepted contributors are listed in [`CONTRIBUTORS.md`](CONTRIBUTORS.md) and credited in release changelogs. Thank you for making Cybermes better!
