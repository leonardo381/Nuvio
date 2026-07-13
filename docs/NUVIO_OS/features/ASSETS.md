# Assets Agent Card

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Feature Cards](README.md)

## 1. Purpose

Guide agents working on CMS assets, images, media fields, native PocketBase file fields, storage files, upload metadata, previews, and public runtime image implications.

## 2. Current operating status

Done but needs regression.

## 3. Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Establishes project boundaries. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Confirms routing and stop conditions for this feature. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Assets touch storage and public rendering. |
| 1 | Admin UI Contract | [../../NUVIO_ADMIN_UI_CONTRACT.md](../../NUVIO_ADMIN_UI_CONTRACT.md) | Defines scoped CMS asset behavior. |
| 1 | SchemaForm Contract | [../../NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md](../../NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md) | Defines file/input behavior in forms. |
| 2 | Operating Manual Assets | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Assets and Images.md` | Human feature guide for asset policy. |
| 2 | Snapshot and Restore | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Snapshot and Restore.md` | Explains storage-file restore requirements. |
| 2 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Selects checks. |
| 2 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Defines required final response structure. |

## 4. Likely code areas

- `examples/base/nuvio_cms_backoffice.go`
- `examples/base/nuvio_cms_backoffice_assets_test.go`
- `ui/src/components/cms/PageCms.svelte`
- `ui/src/components/base/nuvio/schema/InputFile.svelte`
- `tools/dev/cmsqasnapshot/*`
- `pb_data/storage` as runtime storage, not source code.
- Public runtime media helpers only after confirming the target repo.

## 5. Decisions to preserve

- Use scoped asset upload endpoints; why: website access and metadata are enforced there; agent implication: do not upload through raw PB UI calls.
- Keep current MIME policy and size limits unless explicitly requested; why: asset upload is a security boundary; agent implication: do not broaden allowed files casually.
- SVG remains blocked unless a security review changes the policy; why: SVG can carry script-like risk; agent implication: do not add SVG for convenience.
- Native file fields and physical storage files must move together; why: records without files create broken images; agent implication: snapshot/restore must validate storage presence.
- Public runtime image URLs are downstream consumers; why: asset changes can break live pages; agent implication: check image rendering after upload/storage changes.
- Do not change storage provider/path behavior casually; why: it affects backups, restore, and deployment volumes; agent implication: escalate before changing storage layout.

## 6. Allowed work now

- UI polish around asset picker/display.
- Metadata bug fixes on scoped upload.
- Tests for MIME, size, checksum, metadata, and website scoping.
- Snapshot/restore validation for required physical files.
- Documentation clarifications.

## 7. Do not change unless explicitly requested

- File storage provider behavior.
- Runtime storage paths or volume assumptions.
- MIME policy, SVG policy, or max upload size.
- Native file field schema.
- Public runtime image contract.
- Automatic destructive cleanup of storage files.

## 8. Common agent failure modes

- Updating record fields but not storage files.
- Treating file names in manifests as proof that files were copied.
- Weakening MIME validation to make one upload pass.
- Breaking public image URLs by changing field names or path construction.
- Running destructive storage cleanup during a polish task.

## 9. Validation checklist

- Run focused backend asset tests when backend upload/storage code changes.
- Run `cd ui; npm run build` when UI changed.
- Manually upload an allowed image and confirm preview renders.
- Manually confirm SVG and oversized files remain rejected if upload policy was touched.
- For snapshot work, confirm records and physical storage files are restored together and missing required files fail safely.

## 10. Reporting requirements

- Changed files.
- Whether upload, UI, snapshot, docs, or tests changed.
- MIME/size/SVG policy confirmation.
- Storage copy/restore behavior.
- Public runtime image impact.
- Validation results.
- Any existing asset/backfill limitations.
