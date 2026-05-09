/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Booking Phase 2A persistent appointments inbox/archive state.
  const collection = app.findCollectionByNameOrId("pbc_1661203900")
  const fields = collection.fields || []

  let hasArchivedAt = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldName = `${field?.name || ""}`.trim().toLowerCase()

    if (["archivedat", "archived_at"].includes(fieldName)) {
      hasArchivedAt = true
      break
    }
  }

  if (!hasArchivedAt) {
    collection.fields.addAt(17, new Field({
      "hidden": false,
      "id": "date1779203801",
      "max": "",
      "min": "",
      "name": "archivedAt",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "date"
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: Booking Phase 2A persistent appointments inbox/archive state.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203900")
  const fields = collection.fields || []

  let hasArchivedAt = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldId = `${field?.id || ""}`.trim()

    if (fieldId === "date1779203801") {
      hasArchivedAt = true
      break
    }
  }

  if (hasArchivedAt) {
    collection.fields.removeById("date1779203801")
  }

  return app.save(collection)
})
