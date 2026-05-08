/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Booking Phase 7A appointment service snapshots.
  const collection = app.findCollectionByNameOrId("pbc_1661203900")
  const fields = collection.fields || []

  let hasServiceNameSnapshot = false
  let hasServiceDurationSnapshot = false
  let hasServiceDescriptionSnapshot = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldName = `${field?.name || ""}`.trim().toLowerCase()

    if (["servicenamesnapshot", "service_name_snapshot"].includes(fieldName)) {
      hasServiceNameSnapshot = true
    }

    if (["servicedurationminutessnapshot", "service_duration_minutes_snapshot"].includes(fieldName)) {
      hasServiceDurationSnapshot = true
    }

    if (["servicedescriptionsnapshot", "service_description_snapshot"].includes(fieldName)) {
      hasServiceDescriptionSnapshot = true
    }
  }

  if (!hasServiceNameSnapshot) {
    collection.fields.addAt(14, new Field({
      "autogeneratePattern": "",
      "hidden": false,
      "id": "text1779103701",
      "max": 0,
      "min": 0,
      "name": "serviceNameSnapshot",
      "pattern": "",
      "presentable": false,
      "primaryKey": false,
      "required": false,
      "system": false,
      "type": "text"
    }))
  }

  if (!hasServiceDurationSnapshot) {
    collection.fields.addAt(15, new Field({
      "hidden": false,
      "id": "number1779103702",
      "max": null,
      "min": 1,
      "name": "serviceDurationMinutesSnapshot",
      "onlyInt": true,
      "presentable": false,
      "required": false,
      "system": false,
      "type": "number"
    }))
  }

  if (!hasServiceDescriptionSnapshot) {
    collection.fields.addAt(16, new Field({
      "autogeneratePattern": "",
      "hidden": false,
      "id": "text1779103703",
      "max": 0,
      "min": 0,
      "name": "serviceDescriptionSnapshot",
      "pattern": "",
      "presentable": false,
      "primaryKey": false,
      "required": false,
      "system": false,
      "type": "text"
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: Booking Phase 7A appointment service snapshots.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203900")
  const fields = collection.fields || []

  let hasServiceNameSnapshot = false
  let hasServiceDurationSnapshot = false
  let hasServiceDescriptionSnapshot = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldId = `${field?.id || ""}`.trim()

    if (fieldId === "text1779103701") {
      hasServiceNameSnapshot = true
    }

    if (fieldId === "number1779103702") {
      hasServiceDurationSnapshot = true
    }

    if (fieldId === "text1779103703") {
      hasServiceDescriptionSnapshot = true
    }
  }

  if (hasServiceDescriptionSnapshot) {
    collection.fields.removeById("text1779103703")
  }

  if (hasServiceDurationSnapshot) {
    collection.fields.removeById("number1779103702")
  }

  if (hasServiceNameSnapshot) {
    collection.fields.removeById("text1779103701")
  }

  return app.save(collection)
})

