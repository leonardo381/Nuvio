/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: add contact attribution fields for Leads context.
  const collection = app.findCollectionByNameOrId("pbc_1661203100")
  const fields = collection.fields || []

  let hasSource = false
  let hasPage = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldName = `${field?.name || ""}`.trim().toLowerCase()
    if (fieldName === "source") {
      hasSource = true
    }
    if (fieldName === "page") {
      hasPage = true
    }
  }

  if (!hasSource) {
    collection.fields.addAt(collection.fields.length, new Field({
      "autogeneratePattern": "",
      "hidden": false,
      "id": "text1780504301",
      "max": 120,
      "min": 0,
      "name": "source",
      "pattern": "",
      "presentable": false,
      "primaryKey": false,
      "required": false,
      "system": false,
      "type": "text"
    }))
  }

  if (!hasPage) {
    collection.fields.addAt(collection.fields.length, new Field({
      "autogeneratePattern": "",
      "hidden": false,
      "id": "text1780504302",
      "max": 200,
      "min": 0,
      "name": "page",
      "pattern": "",
      "presentable": false,
      "primaryKey": false,
      "required": false,
      "system": false,
      "type": "text"
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: add contact attribution fields for Leads context.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203100")
  const fields = collection.fields || []

  let hasCreatedSourceField = false
  let hasCreatedPageField = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldId = `${field?.id || ""}`.trim()
    if (fieldId === "text1780504301") {
      hasCreatedSourceField = true
    }
    if (fieldId === "text1780504302") {
      hasCreatedPageField = true
    }
  }

  if (hasCreatedSourceField) {
    collection.fields.removeById("text1780504301")
  }

  if (hasCreatedPageField) {
    collection.fields.removeById("text1780504302")
  }

  return app.save(collection)
})