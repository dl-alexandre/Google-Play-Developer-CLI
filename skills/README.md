# gpd agent skills pack

ASC-style skill packs for AI agents operating **gpd** (Google Play Developer CLI).

Each skill is a directory with a `SKILL.md` (YAML frontmatter + operational guide). Flags and examples match live `gpd <cmd> --help` at packaging time — prefer re-checking help if your binary is newer.

## Skills

| Skill | Path | Use for |
| --- | --- | --- |
| **gpd-auth** | [`gpd-auth/SKILL.md`](gpd-auth/SKILL.md) | Service accounts, profiles (`list` / `switch` / `delete` / `logout --name`), `doctor`, `check` |
| **gpd-release** | [`gpd-release/SKILL.md`](gpd-release/SKILL.md) | `validate`, `publish play`, upload/release/rollout/promote/halt/rollback |
| **gpd-reviews-vitals** | [`gpd-reviews-vitals/SKILL.md`](gpd-reviews-vitals/SKILL.md) | `reviews list` / `reply`, `vitals crashes` / `anrs` / query |

## Install / wire into an agent

Keep it simple — no special installer required:

1. **Point the agent at the skill directories** (repo checkout or copied tree):

   ```text
   skills/gpd-auth
   skills/gpd-release
   skills/gpd-reviews-vitals
   ```

2. **Copy into your agent’s skills root** if it only loads a fixed path (examples):

   ```bash
   # from the gpd repo root
   cp -R skills/gpd-auth skills/gpd-release skills/gpd-reviews-vitals \
     /path/to/your-agent/skills/
   ```

3. **Symlink** when you want updates to track the repo:

   ```bash
   ln -s "$(pwd)/skills/gpd-auth" /path/to/your-agent/skills/gpd-auth
   ln -s "$(pwd)/skills/gpd-release" /path/to/your-agent/skills/gpd-release
   ln -s "$(pwd)/skills/gpd-reviews-vitals" /path/to/your-agent/skills/gpd-reviews-vitals
   ```

If your toolchain supports a skills CLI (e.g. `npx skills` or vendor-specific import), point it at this `skills/` directory or at individual `gpd-*` folders — packaging is plain markdown + frontmatter.

## Agent operating rules (all skills)

- Prefer **`--output json`** (and avoid interactive assumptions).
- Always pass **`--package`** for package-scoped commands.
- Prefer **`--dry-run`** before mutating publish/reply/halt/rollback operations.
- Do **not invent flags** — run `gpd <command> --help`.
- Auth before write ops: see **gpd-auth**.

## Related in-repo docs

- Main README — [AI Agent Integration](../README.md#ai-agent-integration)
- [docs/COMMANDS.md](../docs/COMMANDS.md) — generated full command reference
- [docs/auth-parity-guide.md](../docs/auth-parity-guide.md)
- [docs/examples/release-workflow.md](../docs/examples/release-workflow.md)
- `gpd --help` on your installed binary
