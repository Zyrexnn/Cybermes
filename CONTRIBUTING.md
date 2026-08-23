# 🤝 Contributing to Cybermes

Thank you for your interest in contributing to **Cybermes**! 

To maintain high code quality, security standards, and reproducibility across autonomous offensive research workflows, all contributions must go through our standard Pull Request (PR) and code review process.

---

## 📜 Core Contribution Rules

1. **Mandatory Review Gate**:
   - Direct pushes to `main` are restricted.
   - All contributions must be submitted via a **Pull Request (PR)** from a fork or topic branch.
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

### 2. Create a Feature Branch
```bash
git checkout -b fix/your-fix-name
# or
git checkout -b feat/your-feature-name
```

### 3. Make and Verify Changes
* Ensure Python scripts compile and pass linting:
  ```bash
  python3 -m py_compile tools/<your_script>.py
  ```
* Ensure tools are self-contained and document their CLI arguments clearly.

### 4. Commit and Push
```bash
git add .
git commit -m "fix(tools): describe your changes concisely"
git push origin fix/your-fix-name
```

### 5. Open a Pull Request
* Open a PR against the `main` branch of `Zyrexnn/Cybermes`.
* Fill out the automated PR template completely.
* Wait for maintainer review. Respond to feedback or requested changes promptly.

---

## 👥 Recognition
All accepted contributors will be credited in our release changelogs and documentation. Thank you for making Cybermes better!
