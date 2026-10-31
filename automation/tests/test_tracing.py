"""Tracing: dev-crew stage spans and crewAI spans land in one trace per run."""

import pytest
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import SimpleSpanProcessor
from opentelemetry.sdk.trace.export.in_memory_span_exporter import InMemorySpanExporter

from dev_crew import tracing

EXPORTER = InMemorySpanExporter()


@pytest.fixture(scope="module", autouse=True)
def provider():
    tp = TracerProvider()
    tp.add_span_processor(SimpleSpanProcessor(EXPORTER))
    trace.set_tracer_provider(tp)  # global provider can only be set once per process
    yield tp


@pytest.fixture(autouse=True)
def clear():
    EXPORTER.clear()


def test_setup_is_off_by_default(monkeypatch):
    monkeypatch.delenv("TRACING", raising=False)
    assert tracing.setup_tracing() is None
    monkeypatch.setenv("TRACING", "bogus")
    with pytest.raises(SystemExit):
        tracing.setup_tracing()


def test_flow_stages_form_one_trace(repo, tmp_path):
    from test_flow import _run

    flow, _ = _run(repo, tmp_path, max_attempts=4)
    spans = EXPORTER.get_finished_spans()
    names = [s.name for s in spans]
    for expected in ["dev-crew.run", "dev-crew.plan", "dev-crew.implement #1", "dev-crew.verify #1",
                     "dev-crew.implement #3", "dev-crew.review #2", "dev-crew.review #3", "dev-crew.docs"]:
        assert expected in names, names
    assert len({s.context.trace_id for s in spans}) == 1  # everything nested under the run
    root = next(s for s in spans if s.name == "dev-crew.run")
    assert root.attributes["dev_crew.outcome"] == "ready"
    verify1 = next(s for s in spans if s.name == "dev-crew.verify #1")
    assert verify1.attributes["dev_crew.route"] == "retry"
    assert verify1.status.status_code.name == "ERROR"
    assert all(s.attributes.get("openinference.span.kind") == "CHAIN" for s in spans if s.name.startswith("dev-crew."))


def test_crewai_spans_are_captured():
    """A real crewAI crew with a fake LLM is instrumented by OpenInference."""
    pytest.importorskip("openinference.instrumentation.crewai")
    from crewai import Agent, Crew, Task
    from crewai.llms.base_llm import BaseLLM
    from openinference.instrumentation.crewai import CrewAIInstrumentor

    class FakeLLM(BaseLLM):
        def call(self, messages, tools=None, callbacks=None, available_functions=None,
                 from_task=None, from_agent=None, response_model=None):
            return "Thought: done\nFinal Answer: hello from fake llm"

        def supports_function_calling(self) -> bool:
            return False

    CrewAIInstrumentor().instrument(skip_dep_check=True)
    try:
        agent = Agent(role="tester", goal="say hello", backstory="b", llm=FakeLLM(model="fake"))
        task = Task(description="say hello", expected_output="a greeting", agent=agent)
        with tracing.stage("run"):
            out = Crew(agents=[agent], tasks=[task]).kickoff()
    finally:
        CrewAIInstrumentor().uninstrument()
    assert "hello" in out.raw
    spans = EXPORTER.get_finished_spans()
    crew_spans = [s for s in spans if not s.name.startswith("dev-crew.")]
    assert crew_spans, [s.name for s in spans]
    assert len({s.context.trace_id for s in spans}) == 1


def test_tool_calls_become_tool_spans(repo):
    from dev_crew.sandbox import Sandbox
    from dev_crew.tools import EditFileTool, ReadFileTool

    sb = Sandbox(repo)
    ReadFileTool(sandbox=sb)._run(path="server/internal/demo/service.go")
    EditFileTool(sandbox=sb)._run(path="server/internal/demo/service.go", old_text="nope", new_text="x")
    spans = {s.name: s for s in EXPORTER.get_finished_spans()}
    assert spans["tool.read_file"].attributes["openinference.span.kind"] == "TOOL"
    assert "service.go" in spans["tool.read_file"].attributes["input.value"]
    assert spans["tool.edit_file"].status.status_code.name == "ERROR"  # tool returned ERROR
