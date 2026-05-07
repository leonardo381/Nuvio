/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Booking Phase 6E.1 service description and display order.
  const collection = app.findCollectionByNameOrId("pbc_1661203700")
  const fields = collection.fields || []

  let hasDescription = false
  let hasDisplayOrder = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldName = `${field?.name || ""}`.trim().toLowerCase()

    if (fieldName === "description") {
      hasDescription = true
    }

    if (["displayorder", "display_order"].includes(fieldName)) {
      hasDisplayOrder = true
    }
  }

  if (!hasDescription) {
    collection.fields.addAt(4, new Field({
      "autogeneratePattern": "",
      "hidden": false,
      "id": "text1779003601",
      "max": 0,
      "min": 0,
      "name": "description",
      "pattern": "",
      "presentable": false,
      "primaryKey": false,
      "required": false,
      "system": false,
      "type": "text"
    }))
  }

  if (!hasDisplayOrder) {
    collection.fields.addAt(5, new Field({
      "hidden": false,
      "id": "number1779003602",
      "max": null,
      "min": 0,
      "name": "displayOrder",
      "onlyInt": true,
      "presentable": false,
      "required": false,
      "system": false,
      "type": "number"
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: Booking Phase 6E.1 service description and display order.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203700")
  const fields = collection.fields || []

  let hasDescriptionField = false
  let hasDisplayOrderField = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldId = `${field?.id || ""}`.trim()

    if (fieldId === "text1779003601") {
      hasDescriptionField = true
    }

    if (fieldId === "number1779003602") {
      hasDisplayOrderField = true
    }
  }

  if (hasDisplayOrderField) {
    collection.fields.removeById("number1779003602")
  }

  if (hasDescriptionField) {
    collection.fields.removeById("text1779003601")
  }

  return app.save(collection)
})

