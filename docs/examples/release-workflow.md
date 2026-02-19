# Release and Submission Workflow Mapping

This guide maps App Store Connect (ASC) submit and versions workflows to Google Play release workflows in gpd. Play uses track-based releases and staged rollouts rather than a single submit transaction, so the commands below focus on the closest operational equivalents.

## ASC to gpd mapping

| ASC workflow | gpd command | Notes |
| --- | --- | --- |
| `asc submit create` | `gpd publish release` | Create or update a track release |
| `asc submit status` | `gpd publish status` | Track status and release state |
| `asc submit cancel` | `gpd publish halt` | Halt an in-progress production rollout |
| `asc versions release` | `gpd publish release` | Release on a target track |
| `asc versions phased-release update` | `gpd publish rollout` | Update rollout percentage |
| `asc versions phased-release delete` | `gpd publish rollback` | Roll back to a prior version |
| `asc versions promotions create` | `gpd publish promote` | Promote release between tracks |

## Common workflows

### Draft and promote to production

1. Upload the build.
2. Create a draft release on internal or beta.
3. Promote to production when ready.

```bash
gpd publish upload app.aab --package com.example.app
gpd publish release --package com.example.app --track internal --status draft --version-code 123
gpd publish promote --package com.example.app --from-track internal --to-track production
```

### Start a staged rollout

```bash
gpd publish release --package com.example.app --track production --status inProgress --version-code 123
gpd publish rollout --package com.example.app --track production --percentage 10
```

### Increase rollout percentage

```bash
gpd publish rollout --package com.example.app --track production --percentage 50
```

### Halt a rollout (ASC submit cancel)

```bash
gpd publish halt --package com.example.app --track production --confirm
```

### Roll back to a previous version (ASC phased-release delete)

```bash
gpd publish rollback --package com.example.app --track production --confirm --version-code 122
```

### Check release status

```bash
gpd publish status --package com.example.app --track production
```

## Decision paths

- Use `gpd publish promote` when moving a tested release across tracks.
- Use `gpd publish rollout` when you need a staged rollout adjustment.
- Use `gpd publish halt` to stop a rollout without rolling back.
- Use `gpd publish rollback` to return to a prior version when needed.
