# Purchases Implementation Plan

**Status**: Not started
**Priority**: High — purchase verification is critical for backend workflows
**Date**: 2026-03-01

## Background

All purchase management commands are stubs. These are essential for server-side purchase validation, subscription lifecycle management, and fraud detection.

All stubs in `internal/cli/kong_purchases_monetization.go`.

## Commands to Implement

### Products (2 commands)

| Command | Line | API Method |
|---------|------|------------|
| `purchases products acknowledge` | 39 | `Purchases.Products.Acknowledge(packageName, productId, token)` |
| `purchases products consume` | 50 | `Purchases.Products.Consume(packageName, productId, token)` |

- `--product-id` and `--token` (purchase token) required
- `acknowledge` confirms delivery to prevent auto-refund
- `consume` marks consumable product as consumed (can be repurchased)

### Subscriptions (5 commands)

| Command | Line | API Method |
|---------|------|------------|
| `purchases subscriptions acknowledge` | 75 | `Purchases.Subscriptionsv2.Get()` then acknowledge |
| `purchases subscriptions cancel` | 86 | `Purchases.Subscriptions.Cancel(packageName, subId, token)` |
| `purchases subscriptions defer` | 99 | `Purchases.Subscriptions.Defer(packageName, subId, token, deferral)` |
| `purchases subscriptions refund` | 110 | `Purchases.Subscriptions.Refund(packageName, subId, token)` |
| `purchases subscriptions revoke` | 122 | `Purchases.Subscriptions.Revoke(packageName, subId, token)` |

- `--subscription-id` and `--token` required for all
- `defer` also needs `--expected-expiry` and `--new-expiry` timestamps
- `cancel` stops renewal at period end; `revoke` immediately terminates access
- `refund` issues refund but doesn't revoke access (separate operation)

### Verification (2 commands)

| Command | Line | API Method |
|---------|------|------------|
| `purchases verify` | 139 | `Purchases.Products.Get()` or `Purchases.Subscriptionsv2.Get()` |
| `purchases voided list` | 163 | `Purchases.Voidedpurchases.List(packageName)` |

- `verify` auto-detects product vs subscription based on `--type` flag or tries both
- Returns purchase state, acknowledgement status, consumption status, order ID
- `voided list` returns refunded/charged-back purchases
- `voided list` accepts `--start-time`, `--end-time`, `--type` (product/subscription), `--max-results`, `--page-token`

### Capabilities (1 command)

| Command | Line | Notes |
|---------|------|-------|
| `purchases capabilities` | 171 | Informational — list available purchase operations |

## Implementation Order

1. `purchases verify` — most commonly needed, read-only
2. `purchases products acknowledge` — simple write operation
3. `purchases subscriptions cancel/revoke` — common lifecycle ops
4. `purchases voided list` — fraud detection
5. `purchases products consume`
6. `purchases subscriptions acknowledge/defer/refund`
7. `purchases capabilities`

## Notes

- Use Subscriptions v2 API (`Purchases.Subscriptionsv2`) where available — it returns richer data
- Purchase tokens are long opaque strings — consider accepting from stdin for scripting
- These commands do NOT require edit transactions (direct API calls)
- Rate limits are tighter on purchase APIs — ensure retry logic handles 429s

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_purchases_monetization.go` | Implement `Run()` for all 10 commands |
| `internal/api/client.go` | Ensure Purchases service methods are accessible |

## Testing

- Mock Purchases.Products.Get/Acknowledge/Consume
- Mock Purchases.Subscriptions.Cancel/Defer/Refund/Revoke
- Mock Purchases.Subscriptionsv2.Get for verify
- Mock Purchases.Voidedpurchases.List with pagination
- Test verify auto-detection (product vs subscription)
- Test defer with invalid date ranges
