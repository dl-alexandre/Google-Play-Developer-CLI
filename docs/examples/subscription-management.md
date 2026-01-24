# Subscription Management Guide

This guide covers the complete subscription management workflow using the Google Play Developer CLI, including subscriptions, base plans, offers, and regional pricing.

## Table of Contents

1. [Overview](#overview)
2. [Subscription Commands](#subscription-commands)
3. [Base Plan Management](#base-plan-management)
4. [Offer Management](#offer-management)
5. [Regional Pricing](#regional-pricing)
6. [Common Workflows](#common-workflows)

---

## Overview

Google Play's modern subscription architecture uses a three-tier structure:

- **Subscriptions**: The top-level product that groups related subscription plans
- **Base Plans**: The actual subscription plans with pricing and billing periods (e.g., monthly, yearly)
- **Offers**: Promotional offers that modify base plan pricing (e.g., free trials, discounted pricing)

This architecture allows you to:
- Create multiple pricing tiers for the same subscription product
- Run targeted promotional campaigns
- Migrate prices without disrupting existing subscribers
- Manage regional pricing efficiently

### Key Concepts

- **Subscription Product**: A container for base plans (e.g., "premium_subscription")
- **Base Plan**: A specific pricing tier (e.g., "monthly_plan", "yearly_plan")
- **Offer**: A promotional modification to a base plan (e.g., "introductory_offer", "holiday_discount")
- **Price Migration**: Updating prices for new subscribers while preserving existing subscriber pricing
- **Regional Pricing**: Different prices for different countries/regions

---

## Subscription Commands

### Creating Subscriptions

Create a new subscription product using a JSON file:

```bash
gpd monetization subscriptions create \
  --package com.example.app \
  --product-id premium_subscription \
  --file subscription.json
```

**Example `subscription.json`:**

```json
{
  "packageName": "com.example.app",
  "productId": "premium_subscription",
  "basePlans": [
    {
      "basePlanId": "monthly_plan",
      "state": "ACTIVE",
      "regionalConfigs": [
        {
          "price": {
            "priceMicros": "999000",
            "currencyCode": "USD"
          },
          "regionCode": "US"
        }
      ],
      "subscriptionPeriod": {
        "duration": "P1M",
        "unit": "MONTH"
      }
    }
  ],
  "listings": [
    {
      "languageCode": "en-US",
      "title": "Premium Subscription",
      "description": "Unlock all premium features"
    }
  ]
}
```

### Listing Subscriptions

List all subscriptions:

```bash
# List active subscriptions
gpd monetization subscriptions list --package com.example.app

# Include archived subscriptions
gpd monetization subscriptions list \
  --package com.example.app \
  --show-archived

# Fetch all pages
gpd monetization subscriptions list \
  --package com.example.app \
  --all

# Pagination
gpd monetization subscriptions list \
  --package com.example.app \
  --page-size 50 \
  --page-token <token>
```

### Getting Subscription Details

Retrieve detailed information about a specific subscription:

```bash
gpd monetization subscriptions get premium_subscription \
  --package com.example.app
```

### Updating Subscriptions

Update a subscription using a full JSON file:

```bash
gpd monetization subscriptions update premium_subscription \
  --package com.example.app \
  --file subscription-updated.json
```

**Example `subscription-updated.json`:**

```json
{
  "packageName": "com.example.app",
  "productId": "premium_subscription",
  "basePlans": [
    {
      "basePlanId": "monthly_plan",
      "state": "ACTIVE",
      "regionalConfigs": [
        {
          "price": {
            "priceMicros": "999000",
            "currencyCode": "USD"
          },
          "regionCode": "US"
        },
        {
          "price": {
            "priceMicros": "799000",
            "currencyCode": "GBP"
          },
          "regionCode": "GB"
        }
      ],
      "subscriptionPeriod": {
        "duration": "P1M",
        "unit": "MONTH"
      }
    },
    {
      "basePlanId": "yearly_plan",
      "state": "ACTIVE",
      "regionalConfigs": [
        {
          "price": {
            "priceMicros": "9999000",
            "currencyCode": "USD"
          },
          "regionCode": "US"
        }
      ],
      "subscriptionPeriod": {
        "duration": "P1Y",
        "unit": "YEAR"
      }
    }
  ],
  "listings": [
    {
      "languageCode": "en-US",
      "title": "Premium Subscription",
      "description": "Unlock all premium features with monthly or yearly plans"
    }
  ]
}
```

### Patching Subscriptions

Update specific fields using a patch operation:

```bash
gpd monetization subscriptions patch premium_subscription \
  --package com.example.app \
  --file subscription-patch.json \
  --update-mask "basePlans,listings"
```

**Example `subscription-patch.json`:**

```json
{
  "basePlans": [
    {
      "basePlanId": "monthly_plan",
      "state": "ACTIVE"
    }
  ]
}
```

### Archiving Subscriptions

Archive a subscription (soft delete - preserves existing subscribers):

```bash
gpd monetization subscriptions archive premium_subscription \
  --package com.example.app
```

### Deleting Subscriptions

Permanently delete a subscription (requires confirmation):

```bash
gpd monetization subscriptions delete premium_subscription \
  --package com.example.app \
  --confirm
```

### Batch Operations

#### Batch Get Subscriptions

Retrieve multiple subscriptions at once:

```bash
gpd monetization subscriptions batchGet \
  --package com.example.app \
  --ids premium_subscription,basic_subscription,enterprise_subscription
```

#### Batch Update Subscriptions

Update multiple subscriptions in a single operation:

```bash
gpd monetization subscriptions batchUpdate \
  --package com.example.app \
  --file batch-update.json
```

**Example `batch-update.json`:**

```json
{
  "requests": [
    {
      "subscription": {
        "packageName": "com.example.app",
        "productId": "premium_subscription",
        "basePlans": [
          {
            "basePlanId": "monthly_plan",
            "state": "ACTIVE"
          }
        ]
      },
      "updateMask": "basePlans"
    },
    {
      "subscription": {
        "packageName": "com.example.app",
        "productId": "basic_subscription",
        "basePlans": [
          {
            "basePlanId": "monthly_plan",
            "state": "ACTIVE"
          }
        ]
      },
      "updateMask": "basePlans"
    }
  ]
}
```

---

## Base Plan Management

Base plans define the core subscription pricing and billing periods. Each subscription can have multiple base plans (e.g., monthly, yearly).

### Activating Base Plans

Activate a base plan to make it available for purchase:

```bash
gpd monetization baseplans activate premium_subscription monthly_plan \
  --package com.example.app
```

### Deactivating Base Plans

Deactivate a base plan to stop new subscriptions (existing subscribers remain active):

```bash
gpd monetization baseplans deactivate premium_subscription monthly_plan \
  --package com.example.app
```

### Deleting Base Plans

Permanently delete a base plan (requires confirmation):

```bash
gpd monetization baseplans delete premium_subscription monthly_plan \
  --package com.example.app \
  --confirm
```

### Price Migrations

Migrate prices for a base plan. This creates new price cohorts while preserving existing subscriber pricing:

```bash
gpd monetization baseplans migrate-prices premium_subscription monthly_plan \
  --package com.example.app \
  --region-code US \
  --price-micros 1099000
```

**Note**: Price migrations affect only new subscribers. Existing subscribers keep their original pricing until they cancel and resubscribe.

### Batch Price Migrations

Migrate prices for multiple regions at once:

```bash
gpd monetization baseplans batch-migrate-prices premium_subscription \
  --package com.example.app \
  --file migrate-prices.json
```

**Example `migrate-prices.json`:**

```json
{
  "requests": [
    {
      "basePlanId": "monthly_plan",
      "regionalPriceMigrations": [
        {
          "regionCode": "US",
          "priceMicros": "1099000"
        },
        {
          "regionCode": "GB",
          "priceMicros": "899000"
        },
        {
          "regionCode": "JP",
          "priceMicros": "1200000"
        }
      ],
      "regionsVersion": {
        "version": "2022/02"
      }
    },
    {
      "basePlanId": "yearly_plan",
      "regionalPriceMigrations": [
        {
          "regionCode": "US",
          "priceMicros": "10999000"
        }
      ],
      "regionsVersion": {
        "version": "2022/02"
      }
    }
  ]
}
```

### Batch State Updates

Update the state of multiple base plans:

```bash
gpd monetization baseplans batch-update-states premium_subscription \
  --package com.example.app \
  --file baseplan-states.json
```

**Example `baseplan-states.json`:**

```json
{
  "requests": [
    {
      "basePlanId": "monthly_plan",
      "state": "ACTIVE"
    },
    {
      "basePlanId": "yearly_plan",
      "state": "ACTIVE"
    },
    {
      "basePlanId": "legacy_plan",
      "state": "INACTIVE"
    }
  ]
}
```

---

## Offer Management

Offers are promotional modifications to base plans, such as free trials, introductory pricing, or limited-time discounts.

### Creating Offers

Create a promotional offer for a base plan:

```bash
gpd monetization offers create premium_subscription monthly_plan \
  --package com.example.app \
  --offer-id introductory_offer \
  --file offer.json
```

**Example `offer.json`:**

```json
{
  "packageName": "com.example.app",
  "productId": "premium_subscription",
  "basePlanId": "monthly_plan",
  "offerId": "introductory_offer",
  "state": "DRAFT",
  "phases": [
    {
      "duration": "P7D",
      "recurrenceCount": 1,
      "pricing": {
        "price": {
          "priceMicros": "0",
          "currencyCode": "USD"
        },
        "mode": "FREE_TRIAL"
      },
      "regionalConfigs": [
        {
          "priceMicros": "0",
          "regionCode": "US"
        }
      ]
    },
    {
      "duration": "P1M",
      "recurrenceCount": 2,
      "pricing": {
        "price": {
          "priceMicros": "499000",
          "currencyCode": "USD"
        },
        "mode": "PAID"
      },
      "regionalConfigs": [
        {
          "priceMicros": "499000",
          "regionCode": "US"
        }
      ]
    }
  ],
  "targeting": {
    "acquisitionTargetingRule": {
      "acquisitionType": "NEW_SUBSCRIBER_ONLY"
    }
  }
}
```

**Example: Discount Offer**

```json
{
  "packageName": "com.example.app",
  "productId": "premium_subscription",
  "basePlanId": "monthly_plan",
  "offerId": "holiday_discount",
  "state": "DRAFT",
  "phases": [
    {
      "duration": "P1M",
      "recurrenceCount": 3,
      "pricing": {
        "price": {
          "priceMicros": "699000",
          "currencyCode": "USD"
        },
        "mode": "PAID"
      },
      "regionalConfigs": [
        {
          "priceMicros": "699000",
          "regionCode": "US"
        }
      ]
    }
  ],
  "targeting": {
    "acquisitionTargetingRule": {
      "acquisitionType": "NEW_SUBSCRIBER_ONLY"
    }
  }
}
```

### Getting Offers

Retrieve offer details:

```bash
gpd monetization offers get premium_subscription monthly_plan introductory_offer \
  --package com.example.app
```

### Listing Offers

List all offers for a base plan:

```bash
# List all offers
gpd monetization offers list premium_subscription monthly_plan \
  --package com.example.app

# Fetch all pages
gpd monetization offers list premium_subscription monthly_plan \
  --package com.example.app \
  --all
```

### Activating Offers

Activate an offer to make it available:

```bash
gpd monetization offers activate premium_subscription monthly_plan introductory_offer \
  --package com.example.app
```

### Deactivating Offers

Deactivate an offer to stop new enrollments:

```bash
gpd monetization offers deactivate premium_subscription monthly_plan introductory_offer \
  --package com.example.app
```

### Deleting Offers

Permanently delete an offer:

```bash
gpd monetization offers delete premium_subscription monthly_plan introductory_offer \
  --package com.example.app \
  --confirm
```

### Batch Operations

#### Batch Get Offers

Retrieve multiple offers at once:

```bash
gpd monetization offers batchGet premium_subscription monthly_plan \
  --package com.example.app \
  --offer-ids introductory_offer,holiday_discount,student_discount
```

#### Batch Update Offers

Update multiple offers:

```bash
gpd monetization offers batchUpdate premium_subscription monthly_plan \
  --package com.example.app \
  --file batch-offers.json
```

**Example `batch-offers.json`:**

```json
{
  "requests": [
    {
      "offer": {
        "packageName": "com.example.app",
        "productId": "premium_subscription",
        "basePlanId": "monthly_plan",
        "offerId": "introductory_offer",
        "state": "ACTIVE",
        "phases": [
          {
            "duration": "P7D",
            "recurrenceCount": 1,
            "pricing": {
              "price": {
                "priceMicros": "0",
                "currencyCode": "USD"
              },
              "mode": "FREE_TRIAL"
            }
          }
        ]
      },
      "updateMask": "phases,state"
    }
  ]
}
```

#### Batch Update Offer States

Update the state of multiple offers:

```bash
gpd monetization offers batchUpdateStates premium_subscription monthly_plan \
  --package com.example.app \
  --file offer-states.json
```

**Example `offer-states.json`:**

```json
{
  "requests": [
    {
      "offerId": "introductory_offer",
      "state": "ACTIVE"
    },
    {
      "offerId": "holiday_discount",
      "state": "INACTIVE"
    },
    {
      "offerId": "student_discount",
      "state": "ACTIVE"
    }
  ]
}
```

---

## Regional Pricing

Google Play automatically converts prices across regions, but you can also use the CLI to preview conversions and set specific regional prices.

### Converting Region Prices

Convert a price from one currency to multiple regions:

```bash
# Convert $9.99 USD to all regions
gpd monetization convert-region-prices \
  --package com.example.app \
  --price-micros 999000 \
  --currency USD

# Convert to specific regions only
gpd monetization convert-region-prices \
  --package com.example.app \
  --price-micros 999000 \
  --currency USD \
  --to-regions US,GB,JP,DE,FR
```

**Example Output:**

```json
{
  "data": {
    "convertedRegionPrices": {
      "US": {
        "priceMicros": "999000",
        "currencyCode": "USD"
      },
      "GB": {
        "priceMicros": "799000",
        "currencyCode": "GBP"
      },
      "JP": {
        "priceMicros": "1200000",
        "currencyCode": "JPY"
      }
    }
  }
}
```

### Best Practices for Pricing

1. **Use Price Conversion**: Always use `convert-region-prices` to preview pricing before setting regional prices
2. **Round Prices**: Round converted prices to psychologically appealing amounts (e.g., $9.99, £7.99)
3. **Regional Configs**: Set explicit regional prices in base plans for key markets
4. **Price Migration**: Use price migrations to update prices for new subscribers without affecting existing ones
5. **Test Offers**: Create draft offers and test them before activating

**Example: Setting Regional Prices in Base Plan**

```json
{
  "basePlanId": "monthly_plan",
  "state": "ACTIVE",
  "regionalConfigs": [
    {
      "price": {
        "priceMicros": "999000",
        "currencyCode": "USD"
      },
      "regionCode": "US"
    },
    {
      "price": {
        "priceMicros": "799000",
        "currencyCode": "GBP"
      },
      "regionCode": "GB"
    },
    {
      "price": {
        "priceMicros": "1200000",
        "currencyCode": "JPY"
      },
      "regionCode": "JP"
    }
  ],
  "otherRegionsConfig": {
    "price": {
      "priceMicros": "999000",
      "currencyCode": "USD"
    }
  }
}
```

---

## Common Workflows

### Setting Up a New Subscription Product

Complete workflow for creating a new subscription with base plans and offers:

**Step 1: Create the subscription with base plans**

```bash
gpd monetization subscriptions create \
  --package com.example.app \
  --product-id premium_subscription \
  --file subscription.json
```

**Step 2: Verify base plans are active**

```bash
gpd monetization subscriptions get premium_subscription \
  --package com.example.app
```

**Step 3: Activate base plans (if needed)**

```bash
gpd monetization baseplans activate premium_subscription monthly_plan \
  --package com.example.app

gpd monetization baseplans activate premium_subscription yearly_plan \
  --package com.example.app
```

**Step 4: Create promotional offers**

```bash
# Create free trial offer
gpd monetization offers create premium_subscription monthly_plan \
  --package com.example.app \
  --offer-id free_trial \
  --file free-trial-offer.json

# Activate the offer
gpd monetization offers activate premium_subscription monthly_plan free_trial \
  --package com.example.app
```

**Step 5: Verify everything is set up**

```bash
# Check subscription
gpd monetization subscriptions get premium_subscription \
  --package com.example.app

# Check offers
gpd monetization offers list premium_subscription monthly_plan \
  --package com.example.app
```

### Running a Promotional Campaign

Workflow for launching a time-limited promotional offer:

**Step 1: Create the promotional offer (DRAFT state)**

```bash
gpd monetization offers create premium_subscription monthly_plan \
  --package com.example.app \
  --offer-id holiday_2024 \
  --file holiday-offer.json
```

**Step 2: Review the offer**

```bash
gpd monetization offers get premium_subscription monthly_plan holiday_2024 \
  --package com.example.app
```

**Step 3: Activate the offer when ready**

```bash
gpd monetization offers activate premium_subscription monthly_plan holiday_2024 \
  --package com.example.app
```

**Step 4: Monitor and deactivate when campaign ends**

```bash
gpd monetization offers deactivate premium_subscription monthly_plan holiday_2024 \
  --package com.example.app
```

**Step 5: Clean up (optional)**

```bash
gpd monetization offers delete premium_subscription monthly_plan holiday_2024 \
  --package com.example.app \
  --confirm
```

### Updating Prices Globally

Workflow for updating subscription prices across all regions:

**Step 1: Preview price conversions**

```bash
gpd monetization convert-region-prices \
  --package com.example.app \
  --price-micros 1099000 \
  --currency USD \
  --to-regions US,GB,JP,DE,FR,CA,AU
```

**Step 2: Migrate prices for each base plan**

```bash
# Migrate monthly plan prices
gpd monetization baseplans migrate-prices premium_subscription monthly_plan \
  --package com.example.app \
  --region-code US \
  --price-micros 1099000

# Or use batch migration for multiple regions
gpd monetization baseplans batch-migrate-prices premium_subscription \
  --package com.example.app \
  --file migrate-monthly-prices.json
```

**Step 3: Verify price updates**

```bash
gpd monetization subscriptions get premium_subscription \
  --package com.example.app
```

**Note**: Price migrations only affect new subscribers. Existing subscribers retain their original pricing.

### Managing Multiple Subscriptions

Use batch operations to manage multiple subscriptions efficiently:

**Step 1: Get all subscriptions**

```bash
gpd monetization subscriptions list \
  --package com.example.app \
  --all
```

**Step 2: Batch get specific subscriptions**

```bash
gpd monetization subscriptions batchGet \
  --package com.example.app \
  --ids premium_subscription,basic_subscription,enterprise_subscription
```

**Step 3: Batch update subscriptions**

```bash
gpd monetization subscriptions batchUpdate \
  --package com.example.app \
  --file batch-update-all.json
```

### Archiving Old Subscriptions

When deprecating a subscription product:

**Step 1: Deactivate all base plans**

```bash
gpd monetization baseplans deactivate old_subscription monthly_plan \
  --package com.example.app

gpd monetization baseplans deactivate old_subscription yearly_plan \
  --package com.example.app
```

**Step 2: Deactivate all offers**

```bash
gpd monetization offers list old_subscription monthly_plan \
  --package com.example.app \
  --all | jq -r '.data[].offerId' | while read offer; do
  gpd monetization offers deactivate old_subscription monthly_plan "$offer" \
    --package com.example.app
done
```

**Step 3: Archive the subscription**

```bash
gpd monetization subscriptions archive old_subscription \
  --package com.example.app
```

**Note**: Archived subscriptions preserve existing subscribers but prevent new subscriptions.

---

## Additional Resources

- [Google Play Developer API Documentation](https://developers.google.com/android-publisher/api-ref/rest/v3/monetization.subscriptions)
- [Subscription Best Practices](https://developer.android.com/google/play/billing/subscriptions)
- CLI Help: `gpd monetization --help`
- Capabilities: `gpd monetization capabilities --package com.example.app`

---

## Troubleshooting

### Common Issues

1. **"Base plan not found"**: Ensure the base plan exists and is spelled correctly
2. **"Offer state is DRAFT"**: Offers must be activated before they're available to users
3. **"Price migration failed"**: Verify the region code and price micros are valid
4. **"Batch operation partial failure"**: Check individual items in the batch response for errors

### Getting Help

```bash
# Check capabilities
gpd monetization capabilities --package com.example.app

# Verbose output for debugging
gpd monetization subscriptions list \
  --package com.example.app \
  --verbose

# Check authentication
gpd auth status
```
