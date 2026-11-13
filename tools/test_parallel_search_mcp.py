"""Exercise the shipped connection through Hermes, without network or model calls.

Run after installing requirements.txt: python tools/test_parallel_search_mcp.py
The HTTP fixture replaces only the wire transport; Hermes loads the real config,
discovers and dispatches tools, and builds the tool-result conversation message.
"""

import asyncio
import copy
import json
import os
from pathlib import Path
import tempfile
import threading
import unittest
from unittest.mock import patch

import httpx
import yaml


ROOT = Path(__file__).resolve().parent.parent
ENDPOINT = "https://search.parallel.ai/mcp"


class ParallelConnectionTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.home = tempfile.TemporaryDirectory()
        cls.environment = patch.dict(os.environ, {"HERMES_HOME": cls.home.name}, clear=True)
        cls.environment.start()
        cls.addClassCleanup(cls.environment.stop)
        cls.addClassCleanup(cls.home.cleanup)
        # Import after isolating credentials and saved Hermes state.
        from tools import mcp_tool
        from tools.registry import registry
        import model_tools
        from agent.tool_dispatch_helpers import make_tool_result_message

        cls.mcp = mcp_tool
        cls.registry = registry
        cls.dispatcher = model_tools
        cls.result_message = staticmethod(make_tool_result_message)

    def setUp(self):
        self.config = yaml.safe_load((ROOT / ".hermes/config.yaml.example").read_text())
        self.config_path = Path(self.home.name) / "config.yaml"
        self.config_path.write_text(yaml.safe_dump(self.config))
        self.requests = []
        self.mode = "success"
        self.started = threading.Event()
        self.cancelled = threading.Event()
        transport = httpx.MockTransport(self.respond)
        original_client = httpx.AsyncClient

        class FixtureClient(original_client):
            def __init__(self, *args, **kwargs):
                kwargs["transport"] = transport
                super().__init__(*args, **kwargs)

        self.client_patch = patch.object(httpx, "AsyncClient", FixtureClient)
        self.client_patch.start()
        self.addCleanup(self.client_patch.stop)
        self.addCleanup(self.mcp.shutdown_mcp_servers)
        self.addCleanup(self.remove_tools)

    def remove_tools(self):
        for name in self.registry.get_tool_names_for_toolset("mcp-parallel"):
            self.registry.deregister(name)

    def connect(self, **overrides):
        self.config["mcp_servers"]["parallel"].update(overrides)
        self.config_path.write_text(yaml.safe_dump(self.config))
        # Actual runtime config resolver. Avoid launching unrelated stdio servers.
        servers = self.mcp._load_mcp_config()
        return self.mcp.register_mcp_servers({"parallel": servers["parallel"]})

    async def respond(self, request):
        self.assertEqual(str(request.url), ENDPOINT)
        self.assertEqual(request.headers["user-agent"], "Cybermes (https://github.com/Zyrexnn/Cybermes)")
        self.assertNotIn("authorization", request.headers)
        if request.method == "HEAD":
            return httpx.Response(200, headers={"content-type": "application/json"})
        if request.method != "POST":
            return httpx.Response(405)
        body = json.loads(request.content)
        self.requests.append(body)
        if "id" not in body:
            return httpx.Response(202)
        method = body["method"]
        if method == "initialize":
            result = {"protocolVersion": body["params"]["protocolVersion"],
                      "capabilities": {"tools": {}},
                      "serverInfo": {"name": "parallel-fixture", "version": "1.0"}}
        elif method == "tools/list":
            result = {"tools": [
                {"name": name, "description": "Public documentation research",
                 "inputSchema": {"type": "object", "properties": {}}}
                for name in ("web_search", "web_fetch", "unrelated_tool")
            ]}
        elif method == "tools/call":
            if self.mode == "error":
                result = {"isError": True, "content": [{"type": "text", "text": "Search unavailable"}]}
            else:
                payload = {"results": [{"url": "https://example.com/reference",
                                        "title": "Cookie reference",
                                        "excerpts": ["Use Secure and HttpOnly cookie attributes."]}]}
                result = {"content": [{"type": "text", "text": json.dumps(payload)}]}
        else:
            result = {}
        return httpx.Response(200, json={"jsonrpc": "2.0", "id": body["id"], "result": result})

    def call(self, name="web_search", arguments=None):
        return self.dispatcher.handle_function_call(
            "mcp__parallel__" + name,
            arguments or {"objective": "Find cookie documentation", "search_queries": ["secure cookie attributes"]},
        )

    def test_search_fetch_and_result_consumption(self):
        names = self.connect()
        self.assertIn("mcp__parallel__web_search", names)
        self.assertNotIn("mcp__parallel__unrelated_tool", names)
        for name, args in [("web_search", {"objective": "Find cookie documentation",
                                            "search_queries": ["secure cookie attributes"]}),
                           ("web_fetch", {"urls": ["https://example.com/reference"]})]:
            result = self.call(name, args)
            message = self.result_message("mcp__parallel__" + name, result, "fixture-call")
            self.assertEqual(message["role"], "tool")
            self.assertIn("https://example.com/reference", message["content"])
            self.assertIn("HttpOnly", message["content"])
            sent = [r for r in self.requests if r["method"] == "tools/call"][-1]
            self.assertEqual(sent["params"]["name"], name)
            self.assertEqual(sent["params"]["arguments"], args)

    def test_scripted_agent_loop_consumes_search_result(self):
        from openai.types.chat import ChatCompletion
        from run_agent import AIAgent

        self.connect()
        agent = AIAgent(
            model="gpt-4o-mini", provider="openai", api_mode="chat_completions",
            api_key="fixture-key", base_url="https://example.com/v1",
            enabled_toolsets=["parallel"], max_iterations=3, tool_delay=0,
            quiet_mode=True, skip_context_files=True, skip_memory=True,
        )
        agent._disable_streaming = True
        turns = []

        def scripted_completion(kwargs):
            turns.append(kwargs["messages"])
            if len(turns) == 1:
                message = {"role": "assistant", "content": None, "tool_calls": [{
                    "id": "search-call", "type": "function", "function": {
                        "name": "mcp__parallel__web_search",
                        "arguments": json.dumps({"objective": "Find cookie documentation",
                                                 "search_queries": ["secure cookie attributes"]}),
                    }}]}
                finish = "tool_calls"
            else:
                results = [m for m in kwargs["messages"] if m.get("role") == "tool"]
                self.assertTrue(results)
                self.assertIn("https://example.com/reference", results[-1]["content"])
                self.assertIn("HttpOnly", results[-1]["content"])
                message = {"role": "assistant", "content": "Reference: https://example.com/reference"}
                finish = "stop"
            return ChatCompletion(id="fixture-completion", model="gpt-4o-mini", created=0,
                                  object="chat.completion", choices=[{
                                      "index": 0, "finish_reason": finish, "message": message}])

        with patch.object(agent, "_interruptible_api_call", side_effect=scripted_completion):
            result = agent.run_conversation("Find public documentation about secure cookies.")
        self.assertEqual(len(turns), 2)
        self.assertIn("https://example.com/reference", result["final_response"])

    def test_disabled_connection_makes_no_requests(self):
        self.assertEqual(self.connect(enabled=False), [])
        self.assertEqual(self.requests, [])

    def test_toolsets_and_permission_hook(self):
        self.connect()
        defs = self.dispatcher.get_tool_definitions(
            enabled_toolsets=["mcp-parallel"], quiet_mode=True, skip_tool_search_assembly=True)
        self.assertIn("mcp__parallel__web_search", [d["function"]["name"] for d in defs])
        restricted = self.dispatcher.get_tool_definitions(
            enabled_toolsets=["mcp-parallel"], disabled_toolsets=["mcp-parallel"],
            quiet_mode=True, skip_tool_search_assembly=True)
        self.assertNotIn("mcp__parallel__web_search", [d["function"]["name"] for d in restricted])
        before = len(self.requests)
        with patch("hermes_cli.plugins.resolve_pre_tool_block", return_value="Permission denied"):
            self.assertIn("Permission denied", self.call())
        self.assertEqual(len(self.requests), before)

    def test_default_cli_resolver_and_saved_mcp_restrictions(self):
        from hermes_cli.config import load_config
        from hermes_cli.tools_config import _get_platform_tools

        self.connect()
        config = load_config()
        selected = _get_platform_tools(config, "cli")
        self.assertIn("parallel", selected)
        definitions = self.dispatcher.get_tool_definitions(
            enabled_toolsets=sorted(selected), quiet_mode=True, skip_tool_search_assembly=True)
        self.assertIn("mcp__parallel__web_search", [d["function"]["name"] for d in definitions])
        for saved in (["no_mcp"], ["fetch"]):
            config["platform_toolsets"] = {"cli": saved}
            selected = _get_platform_tools(config, "cli")
            self.assertNotIn("parallel", selected)
            definitions = self.dispatcher.get_tool_definitions(
                enabled_toolsets=sorted(selected), quiet_mode=True, skip_tool_search_assembly=True)
            self.assertNotIn("mcp__parallel__web_search", [d["function"]["name"] for d in definitions])

    def test_service_error_reaches_dispatcher(self):
        self.connect()
        self.mode = "error"
        self.assertIn("Search unavailable", json.loads(self.call())["error"])

    async def pending_rpc(self, *args, **kwargs):
        # Isolate cancellation at the session boundary. The real SDK and wire
        # transport are exercised above; no slow remote server is required.
        self.started.set()
        try:
            await asyncio.Event().wait()
        except asyncio.CancelledError:
            self.cancelled.set()
            raise

    def test_timeout_cancels_rpc(self):
        self.connect(timeout=0.2)
        with patch.object(self.mcp._servers["parallel"].session, "call_tool", self.pending_rpc):
            self.assertIn("timed out", json.loads(self.call())["error"])
            self.assertTrue(self.cancelled.wait(2), "Timed-out RPC must be cancelled")

    def test_user_interrupt(self):
        self.connect()
        from tools.interrupt import set_interrupt
        caller_id = threading.get_ident()

        def interrupt_when_started():
            if self.started.wait(2):
                set_interrupt(True, caller_id)

        interrupter = threading.Thread(target=interrupt_when_started)
        interrupter.start()
        try:
            with patch.object(self.mcp._servers["parallel"].session, "call_tool", self.pending_rpc):
                self.assertIn("interrupted", json.loads(self.call())["error"])
                self.assertTrue(self.cancelled.wait(2), "Interrupted RPC must be cancelled")
        finally:
            interrupter.join(3)
            set_interrupt(False)

    def test_saved_provider_and_connection_override(self):
        from hermes_cli.config import load_config
        from tools.web_tools import _get_search_backend
        saved = copy.deepcopy(self.config)
        saved["web"] = {"backend": "tavily"}
        saved["mcp_servers"]["parallel"]["enabled"] = False
        saved["mcp_servers"]["parallel"]["url"] = "https://example.com/custom-mcp"
        self.config_path.write_text(yaml.safe_dump(saved))
        resolved = self.mcp._load_mcp_config()["parallel"]
        self.assertEqual(resolved["url"], "https://example.com/custom-mcp")
        self.assertFalse(resolved["enabled"])
        self.assertEqual(_get_search_backend(), "tavily")
        self.assertEqual(load_config()["web"]["backend"], "tavily")
        self.assertEqual(self.requests, [])

    def test_native_credentials_overrides_and_failure_stay_native(self):
        from hermes_cli.config import invalidate_env_cache
        from tools import web_tools
        from agent.web_search_registry import get_provider

        with patch.dict(os.environ, {"TAVILY_API_KEY": "fixture-key", "EXA_API_KEY": "fixture-key"}):
            invalidate_env_cache()
            self.assertEqual(web_tools._get_search_backend(), "tavily")
            self.config["web"] = {"backend": "tavily", "search_backend": "exa"}
            self.config_path.write_text(yaml.safe_dump(self.config))
            self.assertEqual(web_tools._get_search_backend(), "exa")
            provider = get_provider("exa")
            self.assertIsNotNone(provider)
            with patch.object(provider, "search", return_value={"error": "Configured provider failed"}):
                result = self.dispatcher.handle_function_call("web_search", {"query": "example documentation"})
                self.assertIn("Configured provider failed", result)
            self.assertEqual(self.requests, [])
        invalidate_env_cache()


if __name__ == "__main__":
    unittest.main(verbosity=2)
