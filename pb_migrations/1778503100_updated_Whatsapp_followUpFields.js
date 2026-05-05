/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: add follow-up fields to Whatsapp leads.
  const collection = app.findCollectionByNameOrId("pbc_1661203200")
  const fields = collection.fields || []

  let hasNotes = false
  let hasLastContactedAt = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldName = `${field?.name || ""}`.trim().toLowerCase()
    if (["notes", "note", "internalnotes", "internal_notes"].includes(fieldName)) {
      hasNotes = true
    }
    if (["lastcontactedat", "last_contacted_at", "lastcontacted", "last_contacted"].includes(fieldName)) {
      hasLastContactedAt = true
    }
  }

  if (!hasNotes) {
    collection.fields.addAt(5, new Field({
      "autogeneratePattern": "",
      "hidden": false,
      "id": "text1778503101",
      "max": 0,
      "min": 0,
      "name": "notes",
      "pattern": "",
      "presentable": false,
      "primaryKey": false,
      "required": false,
      "system": false,
      "type": "text"
    }))
  }

  if (!hasLastContactedAt) {
    collection.fields.addAt(6, new Field({
      "hidden": false,
      "id": "date1778503102",
      "max": "",
      "min": "",
      "name": "lastContactedAt",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "date"
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: add follow-up fields to Whatsapp leads.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203200")
  const fields = collection.fields || []

  let hasCreatedNotesField = false
  let hasCreatedLastContactedField = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldId = `${field?.id || ""}`.trim()
    if (fieldId === "text1778503101") {
      hasCreatedNotesField = true
    }
    if (fieldId === "date1778503102") {
      hasCreatedLastContactedField = true
    }
  }

  if (hasCreatedNotesField) {
    collection.fields.removeById("text1778503101")
  }

  if (hasCreatedLastContactedField) {
    collection.fields.removeById("date1778503102")
  }

  return app.save(collection)
})
