from dev_crew import routing
from dev_crew.checks import areas_for, run_checks
from dev_crew.gitops import changed_paths, create_worktree, full_diff, git
from dev_crew.settings import Check


def test_areas_for():
    assert areas_for(["server/main.go"]) == ["server"]
    assert areas_for(["client-vite/client-vite/src/a.ts", "ocr/ocr.go"]) == ["client", "server"]
    assert areas_for(["README.md"]) == []


def test_run_checks_only_changed_area(repo):
    checks = [
        Check("ok", "server", "server", "true"),
        Check("fmt", "server", "server", "echo bad.go", fail_on_output=True),
        Check("client", "client", "client-vite/client-vite", "false"),
    ]
    (repo / "server/internal/demo/new.go").write_text("package demo\n")
    report = run_checks(repo, checks)
    assert report.areas == ["server"]
    assert [r.passed for r in report.results] == [True, False]
    assert "bad.go" in report.failures()


def test_worktree_isolated_and_diff(repo, tmp_path):
    (repo / "client-vite/client-vite/node_modules").mkdir()
    wt = create_worktree(repo, tmp_path / "wts", "crew/test-1")
    assert git("branch", "--show-current", cwd=wt).strip() == "crew/test-1"
    assert (wt / "client-vite/client-vite/node_modules").is_symlink()
    assert changed_paths(wt) == []  # symlinked node_modules is ignored
    (wt / "server/internal/demo/extra.go").write_text("package demo\n// new\n")
    assert changed_paths(wt) == ["server/internal/demo/extra.go"]
    assert "+// new" in full_diff(wt)
    assert changed_paths(repo) == []  # main checkout untouched


def test_routing():
    assert routing.after_verify(has_changes=True, checks_passed=True, attempt=1, max_attempts=3) == routing.PASSED
    assert routing.after_verify(has_changes=False, checks_passed=True, attempt=1, max_attempts=3) == routing.RETRY
    assert routing.after_verify(has_changes=True, checks_passed=False, attempt=3, max_attempts=3) == routing.GIVE_UP
    assert routing.after_review(approved=False, attempt=1, max_attempts=2) == routing.RETRY
    assert routing.after_review(approved=True, attempt=2, max_attempts=2) == routing.APPROVED
