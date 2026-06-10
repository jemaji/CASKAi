"""Formatea un mensaje Conventional Commit. Recurso de la skill commit-helper."""

TYPES = {"feat", "fix", "chore", "docs", "refactor", "test"}


def format_commit(commit_type: str, scope: str, summary: str, body: str = "") -> str:
    if commit_type not in TYPES:
        raise ValueError(f"tipo inválido: {commit_type}")
    head = f"{commit_type}({scope}): {summary}" if scope else f"{commit_type}: {summary}"
    return f"{head}\n\n{body}".strip()
