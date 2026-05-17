/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Public content translations Phase 1A Blocks translations field.
  const collection = app.findCollectionByNameOrId("pbc_4194232374")
  const fields = collection.fields || []

  let hasTranslations = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldName = `${field?.name || ""}`.trim().toLowerCase()

    if (fieldName === "translations") {
      hasTranslations = true
      break
    }
  }

  if (!hasTranslations) {
    collection.fields.addAt(collection.fields.length, new Field({
      "hidden": false,
      "id": "json1779404001",
      "maxSize": 0,
      "name": "translations",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "json"
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: Public content translations Phase 1A Blocks translations field.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_4194232374")
  const fields = collection.fields || []

  let hasTranslationsField = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldId = `${field?.id || ""}`.trim()

    if (fieldId === "json1779404001") {
      hasTranslationsField = true
      break
    }
  }

  if (hasTranslationsField) {
    collection.fields.removeById("json1779404001")
  }

  return app.save(collection)
})
