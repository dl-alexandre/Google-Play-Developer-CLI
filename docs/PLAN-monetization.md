# Monetization Implementation Plan

**Status**: Not started
**Priority**: Medium — full IAP/subscription management
**Date**: 2026-03-01

## Background

All 30 monetization commands are stubs. These cover in-app products, subscriptions, base plans, and promotional offers — the full monetization lifecycle via the `Monetization` resource in the Android Publisher API v3.

All stubs in `internal/cli/kong_purchases_monetization.go`.

## Commands to Implement

### In-App Products (5 commands)

| Command | Line | API Method |
|---------|------|------------|
| `monetization products list` | 209 | `Inappproducts.List(packageName)` |
| `monetization products get` | 219 | `Inappproducts.Get(packageName, sku)` |
| `monetization products create` | 232 | `Inappproducts.Insert(packageName, product)` |
| `monetization products update` | 244 | `Inappproducts.Update(packageName, sku, product)` |
| `monetization products delete` | 254 | `Inappproducts.Delete(packageName, sku)` |

- `--sku` for product ID
- `create`/`update` accept `--price`, `--title`, `--description`, `--status` (active/inactive)
- `create` can accept `--from-json` for full product definition
- `list` supports `--max-results` and `--page-token`

### Subscriptions (9 commands)

| Command | Line | API Method |
|---------|------|------------|
| `monetization subscriptions list` | 284 | `Monetization.Subscriptions.List(packageName)` |
| `monetization subscriptions get` | 294 | `Monetization.Subscriptions.Get(packageName, productId)` |
| `monetization subscriptions create` | 305 | `Monetization.Subscriptions.Create(packageName, sub)` |
| `monetization subscriptions update` | 316 | `Monetization.Subscriptions.Update()` — full replace |
| `monetization subscriptions patch` | 329 | `Monetization.Subscriptions.Patch()` — partial update |
| `monetization subscriptions delete` | 340 | `Monetization.Subscriptions.Delete(packageName, productId)` |
| `monetization subscriptions archive` | 350 | `Monetization.Subscriptions.Archive()` — soft delete |
| `monetization subscriptions batch-get` | 360 | `Monetization.Subscriptions.BatchGet()` |
| `monetization subscriptions batch-update` | 370 | `Monetization.Subscriptions.BatchUpdate()` |

- `--product-id` for subscription ID
- `create` accepts `--from-json` or individual flags for listing details
- `patch` accepts `--update-mask` to specify which fields to update
- `archive` deactivates but preserves for existing subscribers
- Batch operations accept `--product-ids` (comma-separated)

### Base Plans (6 commands)

| Command | Line | API Method |
|---------|------|------------|
| `monetization base-plans activate` | 395 | `Monetization.Subscriptions.BasePlans.Activate()` |
| `monetization base-plans deactivate` | 406 | `Monetization.Subscriptions.BasePlans.Deactivate()` |
| `monetization base-plans delete` | 418 | `Monetization.Subscriptions.BasePlans.Delete()` |
| `monetization base-plans migrate-prices` | 431 | `Monetization.Subscriptions.BasePlans.MigratePrices()` |
| `monetization base-plans batch-migrate` | 442 | `Monetization.Subscriptions.BasePlans.BatchMigratePrices()` |
| `monetization base-plans batch-update-states` | 453 | `Monetization.Subscriptions.BasePlans.BatchUpdateStates()` |

- `--product-id` and `--base-plan-id` required
- `activate`/`deactivate` toggle availability
- `migrate-prices` accepts `--regional-price-migrations` JSON
- Batch operations accept multiple base plan IDs

### Offers (9 commands + capabilities)

| Command | Line | API Method |
|---------|------|------------|
| `monetization offers create` | 483 | `Monetization.Subscriptions.BasePlans.Offers.Create()` |
| `monetization offers get` | 495 | `Monetization.Subscriptions.BasePlans.Offers.Get()` |
| `monetization offers list` | 509 | `Monetization.Subscriptions.BasePlans.Offers.List()` |
| `monetization offers delete` | 522 | `Monetization.Subscriptions.BasePlans.Offers.Delete()` |
| `monetization offers activate` | 534 | `Monetization.Subscriptions.BasePlans.Offers.Activate()` |
| `monetization offers deactivate` | 546 | `Monetization.Subscriptions.BasePlans.Offers.Deactivate()` |
| `monetization offers batch-get` | 558 | `Monetization.Subscriptions.BasePlans.Offers.BatchGet()` |
| `monetization offers batch-update` | 570 | `Monetization.Subscriptions.BasePlans.Offers.BatchUpdate()` |
| `monetization offers batch-update-states` | 582 | `Monetization.Subscriptions.BasePlans.Offers.BatchUpdateStates()` |
| `monetization capabilities` | 590 | Informational |

- `--product-id`, `--base-plan-id`, `--offer-id` required
- Offers define pricing phases (free trial, intro price, regular price)
- `create` accepts `--from-json` for full offer definition
- Batch operations for bulk management

## Implementation Order

1. **Products CRUD** (5) — simplest, standalone resources
2. **Subscriptions list/get** (2) — read-only
3. **Subscriptions create/update/patch** (3) — write ops
4. **Base plans activate/deactivate** (2) — state management
5. **Offers list/get/create** (3) — offer reads and creation
6. **Remaining subscription ops** (4) — delete/archive/batch
7. **Remaining base plan ops** (4) — delete/migrate/batch
8. **Remaining offer ops** (6) — delete/activate/deactivate/batch
9. **Capabilities** (1) — informational

## Notes

- None of these require edit transactions — direct API calls
- Products API (`Inappproducts`) is older; Subscriptions/BasePlans/Offers use the newer `Monetization` resource
- `--from-json` flag for create/update commands should accept file path or stdin
- Price migration is region-specific and requires careful validation
- Consider `--dry-run` for destructive operations (delete, archive)

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_purchases_monetization.go` | Implement `Run()` for all 30 commands |
| `internal/api/client.go` | Ensure Monetization service methods are accessible |

## Testing

- Mock Inappproducts CRUD operations
- Mock Monetization.Subscriptions CRUD + batch
- Mock BasePlans state transitions
- Mock Offers CRUD + batch
- Test pagination on list operations
- Test `--from-json` input parsing
- Test batch operations with mixed success/failure
