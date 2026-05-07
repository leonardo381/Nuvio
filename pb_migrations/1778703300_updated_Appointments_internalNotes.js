/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: add internal notes field to Booking Appointments.
  const collection = app.findCollectionByNameOrId("pbc_1661203900")
  const fields = collection.fields || []

  let hasInternalNotes = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldName = `${field?.name || ""}`.trim().toLowerCase()
    if (["internalnotes", "internal_notes"].includes(fieldName)) {
      hasInternalNotes = true
      break
    }
  }

  if (!hasInternalNotes) {
    collection.fields.addAt(10, new Field({
      "autogeneratePattern": "",
      "hidden": false,
      "id": "text1778703301",
      "max": 0,
      "min": 0,
      "name": "internalNotes",
      "pattern": "",
      "presentable": false,
      "primaryKey": false,
      "required": false,
      "system": false,
      "type": "text"
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: add internal notes field to Booking Appointments.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203900")
  const fields = collection.fields || []

  let hasCreatedInternalNotesField = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldId = `${field?.id || ""}`.trim()
    if (fieldId === "text1778703301") {
      hasCreatedInternalNotesField = true
      break
    }
  }

  if (hasCreatedInternalNotesField) {
    collection.fields.removeById("text1778703301")
  }

  return app.save(collection)
})
