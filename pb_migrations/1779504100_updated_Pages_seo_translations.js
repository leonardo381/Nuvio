/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Public content translations Phase 2B Pages seo_translations field.
  const collection = app.findCollectionByNameOrId("pbc_3945946014")
  const fields = collection.fields || []

  let hasSeoTranslations = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldName = `${field?.name || ""}`.trim().toLowerCase()

    if (fieldName === "seo_translations") {
      hasSeoTranslations = true
      break
    }
  }

  if (!hasSeoTranslations) {
    collection.fields.addAt(collection.fields.length, new Field({
      "hidden": false,
      "id": "json1779504101",
      "maxSize": 0,
      "name": "seo_translations",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "json"
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: Public content translations Phase 2B Pages seo_translations field.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_3945946014")
  const fields = collection.fields || []

  let hasSeoTranslationsField = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldId = `${field?.id || ""}`.trim()

    if (fieldId === "json1779504101") {
      hasSeoTranslationsField = true
      break
    }
  }

  if (hasSeoTranslationsField) {
    collection.fields.removeById("json1779504101")
  }

  return app.save(collection)
})
