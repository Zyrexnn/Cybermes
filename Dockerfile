FROM ubuntu:24.04

ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive
ENV PYTHONUNBUFFERED=1
ENV PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true
ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/google-chrome-stable
ENV CHROME_BIN=/usr/bin/google-chrome-stable
ENV PATH="/opt/hermes-venv/bin:/usr/local/bin:/workspace/tools/bin:${PATH}"
ENV HERMES_HOME=/root/.hermes

WORKDIR /workspace

# 1. Essential runtime packages, network tools, and Chrome/Playwright dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    wget \
    gnupg \
    golang-go \
    ripgrep \
    nmap \
    python3 \
    python3-pip \
    python3-venv \
    ca-certificates \
    libglib2.0-0 \
    libnss3 \
    libnspr4 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libxkbcommon0 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxrandr2 \
    libgbm1 \
    libpango-1.0-0 \
    libcairo2 \
    libasound2t64 \
    fonts-liberation \
    xdg-utils \
    acl \
    && if [ "$TARGETARCH" = "arm64" ]; then \
         apt-get install -y --no-install-recommends chromium-browser && \
         ln -sf /usr/bin/chromium-browser /usr/bin/google-chrome-stable; \
       else \
         wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | gpg --dearmor -o /usr/share/keyrings/google-chrome.gpg && \
         echo "deb [arch=amd64 signed-by=/usr/share/keyrings/google-chrome.gpg] http://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google-chrome.list && \
         apt-get update && \
         apt-get install -y --no-install-recommends google-chrome-stable; \
       fi \
    && rm -rf /var/lib/apt/lists/*

# 2. Node.js LTS & MCP servers
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get install -y --no-install-recommends nodejs && \
    npm install -g @modelcontextprotocol/server-puppeteer @modelcontextprotocol/server-filesystem && \
    npm cache clean --force && \
    rm -rf /var/lib/apt/lists/*

# 3. Copy pre-compiled security tools from host
COPY tools/bin/* /usr/local/bin/
RUN chmod +x /usr/local/bin/* 2>/dev/null || true

# 4. Install Hermes Agent, Playwright & Python security tools
RUN python3 -m venv /opt/hermes-venv && \
    /opt/hermes-venv/bin/pip install --no-cache-dir hermes-agent playwright pyyaml requests python-telegram-bot arjun rich markdown jinja2 mcp-server-fetch && \
    printf '#!/usr/bin/env bash\nexec /opt/hermes-venv/bin/python3 "$@"\n' > /usr/local/bin/python3 && chmod +x /usr/local/bin/python3 && \
    printf '#!/usr/bin/env bash\nexec /opt/hermes-venv/bin/pip "$@"\n' > /usr/local/bin/pip && chmod +x /usr/local/bin/pip && \
    ln -sf /opt/hermes-venv/bin/hermes /usr/local/bin/hermes && \
    ln -sf /opt/hermes-venv/bin/mcp-server-fetch /usr/local/bin/mcp-server-fetch && \
    /opt/hermes-venv/bin/playwright install chromium 2>/dev/null || true

# 5. Copy Workspace, skills, and configuration
COPY . /workspace/

# Build Go native tools from source inside container (ensuring arch compatibility)
RUN go build -ldflags="-s -w" -o /usr/local/bin/smart_pipe /workspace/cmd/smart_pipe && \
    go build -ldflags="-s -w" -o /usr/local/bin/secret_scan /workspace/cmd/secret_scan && \
    go build -ldflags="-s -w" -o /usr/local/bin/search_knowledge /workspace/cmd/search_knowledge && \
    go build -ldflags="-s -w" -o /usr/local/bin/aggregate_reports /workspace/cmd/aggregate_reports && \
    chmod +x /usr/local/bin/smart_pipe /usr/local/bin/secret_scan /usr/local/bin/search_knowledge /usr/local/bin/aggregate_reports

# Setup Hermes configuration directories
RUN mkdir -p /root/.hermes/skills /workspace/reports /workspace/recon /workspace/targets /workspace/output /workspace/logs && \
    cp -r /workspace/skills/* /root/.hermes/skills/ 2>/dev/null || true && \
    cp /workspace/.hermes/config.yaml /root/.hermes/config.yaml 2>/dev/null || true && \
    cp /workspace/.hermes/auth.json /root/.hermes/auth.json 2>/dev/null || true && \
    cp /workspace/.hermes/.env /root/.hermes/.env 2>/dev/null || true && \
    chmod +x /workspace/entrypoint.sh

ENTRYPOINT ["/workspace/entrypoint.sh"]
