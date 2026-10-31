"""Pure routing decisions for the flow, kept separate so they are easy to test."""

from __future__ import annotations

PASSED, RETRY, GIVE_UP, APPROVED = "checks_passed", "retry", "give_up", "approved"


def after_verify(*, has_changes: bool, checks_passed: bool, attempt: int, max_attempts: int) -> str:
    if has_changes and checks_passed:
        return PASSED
    return RETRY if attempt < max_attempts else GIVE_UP


def after_review(*, approved: bool, attempt: int, max_attempts: int) -> str:
    if approved:
        return APPROVED
    return RETRY if attempt < max_attempts else GIVE_UP
