---
description: Prepare a pull request from local changes (lint, test, commit, draft PR)
---

# Prepare Pull Request

Prepare local changes for a pull request — lint, test, commit, and draft the PR description.

CRITICAL RULE: NEVER perform any write action on GitHub without explicit user permission. This includes but is not limited to: submitting PR reviews, posting comments, creating/merging pull requests, pushing commits, creating branches, or creating issues. Always draft the content locally and present it to the user for review and approval BEFORE publishing anything to GitHub.

## When to Use

- When the user wants to submit their work as a PR
- Invoked via `/prepare-pr`

// turbo-all

## Steps

### 1. Assess the current state

```bash
git status
git diff HEAD --stat
git log --oneline -5
```

Identify:
- Current branch name
- Base branch (usually `master` or `main`)
- What files changed and rough scope

If on `master`/`main`, ask the user what branch name to use and create it.

### 2. Run `/review-changes`

Before proceeding, run the `/review-changes` workflow which covers lint, build, tests, documentation checks, and a full code review. Fix any critical or suggested issues before continuing.

### 3. Stage and commit

When updating a PR **prefer amending** the existing commit on the branch unless:
- The last commit is from another contributor (e.g., pushing to someone else's PR branch)
- There are multiple logically distinct changes that warrant separate commits

Check the current branch history first:

```bash
git log --oneline main..HEAD
```

#### If the branch already has a commit and the changes belong to the same logical unit:

Propose amending:

```bash
git add -A
git commit --amend
```

#### If creating a new commit:

Propose a commit message following [Conventional Commits](https://www.conventionalcommits.org/) style:

```
<type>(<scope>): <short summary>

<body — what and why, not how>
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`, `ci`

#### If the branch has multiple commits that don't make sense individually:

Warn the user and propose squashing:

> ⚠️ This branch has N commits. Some of them may not make sense individually. Would you like to squash them into a single commit?

```bash
git rebase -i main
```

**Do NOT commit or amend without explicit user approval of the message and strategy.**

### 4. Push

```bash
git push origin <branch-name>
```

If the branch doesn't exist upstream yet:

```bash
git push -u origin <branch-name>
```

### 5. Draft the PR

Generate a PR title and description. Structure:

```markdown
## Title
<type>(<scope>): <short summary>

## Description

### What
Brief description of the change.

### Why
Motivation or issue reference.

### Changes
- Bullet list of key changes grouped by component
- Focus on what a reviewer needs to know

### Testing
- What was tested and how
- Any manual verification steps
```

**Present the draft to the user for review. Do NOT create the PR on GitHub until explicitly approved.**

### 6. Create PR (after approval only)

Only after the user explicitly approves the PR content, use the GitHub MCP tools to create the PR. Remind the user:

> Ready to create the PR on GitHub. Shall I proceed?

## Important Notes

- **Never push or create PRs without user approval** — this is a hard rule
- If changes span multiple concerns, suggest splitting into separate PRs
- Always run the full pre-flight (lint + build + test) before proposing a commit
- If the branch already has a PR open, offer to update it instead of creating a new one
- Check for any CI/CD implications (e.g., new env vars, config changes)