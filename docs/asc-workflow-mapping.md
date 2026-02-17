# ASC Workflow Mapping (Submit, Builds, Analytics, Localization)

This guide maps common App Store Connect (ASC) operational workflows to `gpd` command sequences, including known Google Play model differences.

## 1) Submit / Release Workflow Mapping

### ASC-style goal: create a release and submit for review

Google Play equivalent is track-based release management:

```bash
# Draft release to internal track
gpd publish release --package <pkg> --track internal --status draft

# Roll out to a percentage
gpd publish rollout --package <pkg> --track production --percentage 10

# Check status
gpd publish status --package <pkg> --track production

# Halt rollout
gpd publish halt --package <pkg> --track production

# Rollback
gpd publish rollback --package <pkg> --track production

# Promote track
gpd publish promote --package <pkg> --from-track internal --to-track production
```

### Decision paths

- Increase staged rollout when metrics are healthy.
- Halt when regressions appear.
- Roll back if user impact is severe.
- Promote when lower environment validation is complete.

## 2) Apps & Builds Model Differences

ASC and Google Play differ in artifact and distribution models:

- ASC has a broader build registry view with TestFlight group semantics.
- Google Play is track-oriented for distribution and tester access.

### Recommended Play-native alternatives

- Use tracks as environment gates (internal -> alpha -> beta -> production).
- Use tester commands at track scope:
  - `gpd publish testers list`
  - `gpd publish testers add`
  - `gpd publish testers remove`

## 3) Analytics & Reporting Scope Differences

`gpd` covers Play Reporting and Android vitals-focused data, which does not fully mirror ASC analytics/sales breadth.

### Supported focus areas

- Android vitals anomalies, crashes, and ANR insights.
- Query-based analytics where exposed by Play reporting endpoints.

### Practical guidance

- Use `gpd analytics query` and vitals commands for API-available data.
- For unsupported ASC-analog metrics, use Play Console exports or complementary BI pipelines.

## 4) Metadata & Localization Task Mapping

### ASC task -> gpd sequence examples

1. Update listing text/localization

```bash
gpd publish listing get --package <pkg> --language en-US
gpd publish listing update --package <pkg> --language en-US --title "..." --short-description "..."
```

2. Update app details

```bash
gpd publish details get --package <pkg>
gpd publish details update --package <pkg> --contact-email ops@example.com
```

3. Upload localization images

```bash
gpd publish images upload --package <pkg> --language en-US --image-type phone-screenshots --file ./shot1.png
```

### Known boundaries

- Localization workflow follows Play listing/image models rather than ASC's exact structure.
- Pricing/availability and certain metadata flows differ conceptually and operationally.
