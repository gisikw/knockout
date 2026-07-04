# Project → Realm/Campaign Mapping

Both the **ko QQL shim** (`cmd_shim.go`) and the **Questbook bulk-import**
dispatch need to know which Questbook *realm* (and optionally *campaign*) a ko
project belongs to. That binding lives in one static, hand-maintained file. This
document is its contract; the shape is shared by both sides.

- **Default location:** `~/.config/knockout/qql-mapping.yaml`
  (honors `$XDG_CONFIG_HOME`).
- **Override:** `KO_QQL_MAPPING=/path/to/file`.
- **Example:** `qql-mapping.example.yaml` in this repo.

## Format

```yaml
# Fallback realm slug for any ko project not listed below.
default_realm: knockout-legacy

projects:
  fort-nix:                      # ko registry tag (the #tag)
    realm: fort-nix              # Questbook realm slug (the anchor)
    campaign: fort-nix-maintenance   # optional Questbook campaign slug
  questbook:
    realm: questbook
    campaign: questbook-buildout
  gee:
    realm: gee                   # campaign optional; realm alone is fine
```

The keys under `projects:` are ko project **tags** — the same keys that appear
as `tag` in `ko export` output (see `EXPORT_SCHEMA.md`), so the two files line
up 1:1.

## Resolution rules

For a given project tag (see `QQLMapping.Resolve`):

1. If the tag is listed, use its `realm` and `campaign`.
2. If `realm` is empty (or the tag is unlisted), fall back to `default_realm`.
3. If `default_realm` is also empty, fall back to **the tag itself** as the
   realm slug.

So the realm is **never empty** — every quest always has a realm anchor, which
satisfies Questbook's quest invariant (`realm_id OR campaign_id OR parent_id`).

## Realm vs. campaign (why realm is the anchor)

- **Realms** need only a slug + name. The shim will **create a missing realm**
  automatically (it's cheap and safe).
- **Campaigns** require a mandatory `goal` on creation (a design decision — a
  campaign is a line of intent with an exit condition / north star). The shim
  therefore **never auto-creates a campaign**: an unmapped or not-yet-existing
  campaign is a warning, and the quest is anchored to the realm only. Create
  campaigns deliberately with `qb mutate` (with a real goal), then reference
  their slug here.

## A missing mapping file is fine

If the file does not exist, the shim loads an empty mapping and every project
falls back to `tag → realm slug`. This keeps the shim usable out of the box
during early cutover; add explicit entries as realms/campaigns get organized.
