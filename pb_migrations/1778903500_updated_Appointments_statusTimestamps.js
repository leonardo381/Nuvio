/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Booking Phase 6A appointment status timestamps.
  const collection = app.findCollectionByNameOrId("pbc_1661203900")
  const fields = collection.fields || []

  let hasConfirmedAt = false
  let hasCancelledAt = false
  let hasRescheduledAt = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldName = `${field?.name || ""}`.trim().toLowerCase()

    if (["confirmedat", "confirmed_at"].includes(fieldName)) {
      hasConfirmedAt = true
    }

    if (["cancelledat", "cancelled_at"].includes(fieldName)) {
      hasCancelledAt = true
    }

    if (["rescheduledat", "rescheduled_at"].includes(fieldName)) {
      hasRescheduledAt = true
    }
  }

  if (!hasConfirmedAt) {
    collection.fields.addAt(11, new Field({
      "hidden": false,
      "id": "date1778903501",
      "max": "",
      "min": "",
      "name": "confirmedAt",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "date"
    }))
  }

  if (!hasCancelledAt) {
    collection.fields.addAt(12, new Field({
      "hidden": false,
      "id": "date1778903502",
      "max": "",
      "min": "",
      "name": "cancelledAt",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "date"
    }))
  }

  if (!hasRescheduledAt) {
    collection.fields.addAt(13, new Field({
      "hidden": false,
      "id": "date1778903503",
      "max": "",
      "min": "",
      "name": "rescheduledAt",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "date"
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: Booking Phase 6A appointment status timestamps.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203900")
  const fields = collection.fields || []

  let hasConfirmedAt = false
  let hasCancelledAt = false
  let hasRescheduledAt = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldId = `${field?.id || ""}`.trim()

    if (fieldId === "date1778903501") {
      hasConfirmedAt = true
    }

    if (fieldId === "date1778903502") {
      hasCancelledAt = true
    }

    if (fieldId === "date1778903503") {
      hasRescheduledAt = true
    }
  }

  if (hasRescheduledAt) {
    collection.fields.removeById("date1778903503")
  }

  if (hasCancelledAt) {
    collection.fields.removeById("date1778903502")
  }

  if (hasConfirmedAt) {
    collection.fields.removeById("date1778903501")
  }

  return app.save(collection)
})
