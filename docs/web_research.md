# Web research in the Cybermes CLI

Fresh Cybermes CLI installations connect Hermes Agent to
[Parallel Search MCP](https://docs.parallel.ai/integrations/mcp/search-mcp).
It provides web search in Fast mode and page fetching without an account or API key.
Use it to look up public vendor documentation and remediation references:

```bash
./cybermes "Use Parallel to find public documentation on secure HTTP cookies, then fetch a relevant page and cite its URL."
```

Hermes discovers `web_search` and `web_fetch` over Streamable HTTP and exposes
them as `mcp__parallel__web_search` and `mcp__parallel__web_fetch`. Results, including
source URLs and excerpts, return to the agent's conversation. This connection
belongs to the Hermes CLI workflow; installing the Cybermes Go MCP server in an
external AI client does not add it to that client.

## Existing installations

Setup preserves an existing `.hermes/config.yaml`. To enable Parallel there,
update the Python dependencies and add this entry under the existing
`mcp_servers` mapping in `.hermes/config.yaml`:

```bash
python -m pip install -r requirements.txt
```

```yaml
mcp_servers:
  parallel:
    url: https://search.parallel.ai/mcp
    headers:
      User-Agent: Cybermes (https://github.com/Zyrexnn/Cybermes)
    connect_timeout: 15
    timeout: 30
    tools:
      resources: false
      prompts: false
      include:
        - web_search
        - web_fetch
```

Keep other servers and settings when merging the entry, then restart the CLI.
In Docker, edit the mounted Hermes configuration instead. Search queries and
requested URLs are sent to Parallel; use public research terms and keep
assessment secrets out of requests. The anonymous service has usage limits;
connection failures and tool errors are returned by Hermes.

## Controls

To stop the connection, add `enabled: false` under `mcp_servers.parallel` and
restart Hermes. You can also exclude the `mcp-parallel` toolset with Hermes's
toolset controls. Existing platform permissions and approval settings still apply.

This entry does not change Hermes's native `web.backend`, `web.search_backend`,
`web.extract_backend`, provider credentials, or existing MCP server settings.
Configured native providers remain available under their own tools. Parallel is
a separately named MCP connection; it is not a silent fallback for a failed
explicit provider choice.

## Contributor check

After installing the Python dependencies, run
`python tools/test_parallel_search_mcp.py`. The deterministic tests use the real
Hermes config loader, MCP client, tool dispatcher and a scripted agent loop with an HTTP fixture.
They make no live search or model requests; timeout and interrupt tests hold
the session RPC pending to check cancellation.
