import pytest

from dev_crew.sandbox import Sandbox, SandboxError
from dev_crew.tools import EditFileTool, ListFilesTool, ReadFileTool, SearchCodeTool, WriteFileTool


def test_rejects_escape_and_secrets(repo):
    sb = Sandbox(repo)
    with pytest.raises(SandboxError):
        sb.resolve("../outside.txt")
    with pytest.raises(SandboxError):
        sb.resolve("server/.env")
    with pytest.raises(SandboxError):
        sb.resolve(".git/config")
    assert sb.resolve("server/internal/demo/service.go").is_file()


def test_locked_files_are_read_only(repo):
    sb = Sandbox(repo)
    assert sb.resolve("server/go.mod").is_file()
    with pytest.raises(SandboxError):
        sb.resolve("server/go.mod", write=True)
    with pytest.raises(SandboxError):
        sb.resolve("automation/anything.py", write=True)


def test_list_hides_secrets(repo):
    out = ListFilesTool(sandbox=Sandbox(repo))._run("server")
    assert "server/internal/demo/service.go" in out
    assert ".env" not in out


def test_read_and_search(repo):
    sb = Sandbox(repo)
    assert "    3  func Ok()" in ReadFileTool(sandbox=sb)._run("server/internal/demo/service.go")
    assert "service.go:3" in SearchCodeTool(sandbox=sb)._run(r"func \w+", "server", "*.go")


def test_edit_requires_unique_match(repo):
    sb = Sandbox(repo)
    edit = EditFileTool(sandbox=sb)
    path = "server/internal/demo/service.go"
    assert edit._run(path, "return true", "return false") == f"Edited {path}"
    assert "not found" in edit._run(path, "return true", "x")
    WriteFileTool(sandbox=sb)._run("server/internal/demo/dup.go", "a\na\n")
    assert "appears 2 times" in edit._run("server/internal/demo/dup.go", "a", "b")


def test_write_blocks_secrets_and_md_only(repo):
    sb = Sandbox(repo)
    assert WriteFileTool(sandbox=sb)._run("server/.env", "X=1").startswith("ERROR")
    md_writer = WriteFileTool(sandbox=sb, allowed_suffixes=(".md",))
    assert md_writer._run("server/x.go", "package x").startswith("ERROR")
    assert md_writer._run("server/NOTES.md", "# hi").startswith("Created")
