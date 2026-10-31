"""Optional tracing to Arize Phoenix. Turn on with TRACING=phoenix in automation/.env.

When on, every LLM call, agent step and tool call (via OpenInference auto-instrumentation)
plus dev-crew's own stages (plan, implement attempt N, verify, review...) appear as one
trace per run at http://localhost:6006. When off, everything here is a no-op.
"""

from __future__ import annotations

import json
import os
import urllib.error
import urllib.request
from contextlib import contextmanager
from typing import Any, Iterator

from opentelemetry import trace
from opentelemetry.trace import Status, StatusCode

TRACER = trace.get_tracer("dev_crew")
_provider = None

DEFAULT_ENDPOINT = "http://localhost:6006"
DEFAULT_PROJECT = "dev-crew"


def _endpoint() -> str:
    return os.getenv("PHOENIX_COLLECTOR_ENDPOINT", DEFAULT_ENDPOINT).rstrip("/")


def _reachable(url: str, timeout: float = 2.0) -> bool:
    try:
        urllib.request.urlopen(url, timeout=timeout)  # noqa: S310 - local URL from config
        return True
    except urllib.error.HTTPError:
        return True  # server answered
    except (urllib.error.URLError, OSError):
        return False


def setup_tracing() -> str | None:
    """Configure Phoenix if TRACING=phoenix. Returns the UI URL, or None when off."""
    mode = os.getenv("TRACING", "").strip().lower()
    if mode in ("", "off", "none", "false", "0"):
        return None
    if mode != "phoenix":
        raise SystemExit(f"Unknown TRACING={mode!r}. Use TRACING=phoenix or leave it empty.")
    try:
        from phoenix.otel import register
    except ImportError as exc:
        raise SystemExit("Tracing needs the extra packages: run `uv sync --extra tracing`.") from exc

    base = _endpoint()
    if not _reachable(base):
        print(f"[dev-crew] WARNING: Phoenix is not reachable at {base}. Start it with "
              "`docker compose -f phoenix/docker-compose.yml up -d`. Traces from this run will be lost.")
    global _provider
    _provider = register(
        endpoint=f"{base}/v1/traces",
        project_name=os.getenv("PHOENIX_PROJECT_NAME", DEFAULT_PROJECT),
        batch=True,
        auto_instrument=True,  # crewAI, Anthropic, OpenAI, LiteLLM instrumentors if installed
        verbose=False,
    )
    return base


def shutdown_tracing() -> None:
    if _provider is not None:
        _provider.force_flush()
        _provider.shutdown()


def _text(value: Any, limit: int = 20_000) -> str:
    if not isinstance(value, str):
        value = json.dumps(value, default=str, indent=2)
    return value if len(value) <= limit else value[:limit] + "\n...[truncated]"


@contextmanager
def stage(name: str, input: Any = None, **attrs: Any) -> Iterator[Any]:
    """A dev-crew stage span, shown in Phoenix as a CHAIN with input/output."""
    with TRACER.start_as_current_span(f"dev-crew.{name}") as span:
        span.set_attribute("openinference.span.kind", "CHAIN")
        if input is not None:
            span.set_attribute("input.value", _text(input))
        for key, value in attrs.items():
            if value is not None:
                span.set_attribute(f"dev_crew.{key}", value if isinstance(value, (str, bool, int, float)) else str(value))
        yield span


def set_output(span: Any, output: Any, ok: bool = True) -> None:
    span.set_attribute("output.value", _text(output))
    span.set_status(Status(StatusCode.OK if ok else StatusCode.ERROR))


class _ToolSpan:
    def __init__(self, span):
        self.span = span

    def set_attribute(self, key, value):
        self.span.set_attribute(key, value)

    def finish(self, result: Any) -> None:
        set_output(self.span, result, ok=not str(result).startswith("ERROR"))


@contextmanager
def tool_span(name: str, args: Any) -> Iterator[_ToolSpan]:
    """A TOOL span (shown with a wrench icon in Phoenix) for one dev-crew tool call."""
    with TRACER.start_as_current_span(f"tool.{name}") as span:
        span.set_attribute("openinference.span.kind", "TOOL")
        span.set_attribute("tool.name", name)
        span.set_attribute("input.value", _text(args))
        yield _ToolSpan(span)
