---
name: content_writer
description: Draft, edit, and format content items from the publishing queue. Takes a content item number or title, reads the source material, writes or edits it into publishable form for the target platform (Dev.to, HN, Reddit, Twitter, GitHub PR, gist), verifies no personal data leaks, commits to the docs repo, and displays the formatted output for posting. Use when drafting a blog post, preparing today's content, editing research into a Dev.to article, or formatting content for a specific platform.
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Grep
  - Glob
  - Agent
  - mcp__github__create_or_update_file
  - mcp__github__create_pull_request
  - mcp__github__push_files
---

# Content Writer

Draft and format content items from `content-calendar.json` for publishing.

## Setup

```bash
source ~/hairglasses-studio/.env && export GH_TOKEN="$GITHUB_ORG_ADMIN_PAT"
```

## Default loop

1. **Identify item**: Parse `$ARGUMENTS` for item number (`#8`), title keyword, or `today` (next unposted item for current time slot). Read `content-calendar.json` to get item details.

2. **Determine effort level and process**:

   **R (Ready to post):**
   - Read source file from `source_file` path
   - Verify numbers are current: cross-reference mcpkit (`go list ./... | wc -l` = 72 packages), ralphglasses (222 MCP tools), dotfiles (157 shaders)
   - Fix any stale numbers in the content
   - Copy to `docs/content/published/<date>-<slug>.md`

   **E (Edit from research):**
   - Read the research source file (typically 400-1,500 lines)
   - Extract the 5-8 key insights or patterns
   - Write a 800-2,000 word blog post: hook → problem → solution → code examples → takeaway
   - Add platform-specific frontmatter
   - Save to `docs/content/published/<date>-<slug>.md`

   **W (Write from scratch):**
   - Read resume achievements at `~/hairglasses-studio/jobb/resumes/drafts-2026/achievement-portfolio.md`
   - Read relevant public repo code for working examples
   - Write anonymized 1,000-2,500 word post: NO employer names, focus on patterns and numbers
   - Save to `docs/content/published/<date>-<slug>.md`

3. **Platform formatting**:

   **Dev.to:**
   ```yaml
   ---
   title: "Title Here"
   published: true
   tags: [go, mcp, ai, devtools]
   cover_image: https://placeholder.example.com/cover.png
   ---
   ```

   **Hacker News:** Plain text only. 2-space indent for code blocks. Asterisks for *italic*. No markdown links in body. Keep under 2,000 chars for the text field.

   **Reddit:** Markdown. Match subreddit tone:
   - r/golang: technical, "here's what I built and why"
   - r/unixporn: visual, include neofetch + screenshots
   - r/LocalLLaMA: practical, "here's what this enables for local LLMs"

   **Twitter/X:** Thread format. 280 chars per tweet. Number each tweet (1/N). End with repo link.

   **GitHub PR:** Follow target repo's CONTRIBUTING.md format exactly. Alphabetical placement.

4. **Security verification** (MANDATORY before commit):
   ```bash
   grep -rn '/home/hg\|mixellburk\|mitch@\|ghp_\|gho_\|sk-\|API_KEY\|SECRET' <output_file>
   ```
   If any match: STOP. Remove the personal data. Re-verify. Never commit content with personal data.

5. **Commit and push**:
   ```bash
   cd ~/hairglasses-studio/docs
   git add content/published/<date>-<slug>.md strategy/launch-content/content-calendar.json
   git commit -m "content: draft <title> for <platform>"
   git push origin main
   ```

6. **Update calendar state**: Set item status to "drafted" in `content-calendar.json`.

7. **Display output**: Show the formatted content for the user to copy-paste to the platform. For GitHub PRs, offer to create the PR directly via `mcp__github__create_pull_request`.

## Anonymization rules

- Never name employers (Galileo AI, Primer.ai, Capella Space, Google — use "a production platform" or "a previous role")
- Use `hairglasses-studio` org name and public repo names freely
- Resume numbers (832K lines, 1,860 tools, 23 workers) are OK — they describe the pattern, not the employer
- No personal emails, no home paths, no API keys
