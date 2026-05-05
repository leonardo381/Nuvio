/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: add optional WhatsApp status to support inbox/archive lead actions.
  const collection = app.findCollectionByNameOrId("pbc_1661203200")
  const fields = collection.fields || []
  let existingStatusField = null
  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    if (`${field?.name || ""}`.trim().toLowerCase() === "status") {
      existingStatusField = field
      break
    }
  }

  // add field only if missing
  if (!existingStatusField) {
    collection.fields.addAt(4, new Field({
      "hidden": false,
      "id": "select1778402601",
      "maxSelect": 1,
      "name": "status",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "select",
      "values": [
        "new",
        "read",
        "archived"
      ]
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: add optional WhatsApp status to support inbox/archive lead actions.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203200")
  const fields = collection.fields || []
  let createdStatusField = null
  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    if (`${field?.id || ""}`.trim() === "select1778402601") {
      createdStatusField = field
      break
    }
  }

  // remove only field created by this migration
  if (createdStatusField) {
    collection.fields.removeById("select1778402601")
  }

  return app.save(collection)
})
