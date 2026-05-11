/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Newsletter lifecycle token/timestamp fields.
  const collection = app.findCollectionByNameOrId("pbc_1661203400")
  const fields = collection.fields || []

  let hasConfirmationTokenHash = false
  let hasConfirmationTokenExpiresAt = false
  let hasUnsubscribeTokenHash = false
  let hasUnsubscribedAt = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldName = `${field?.name || ""}`.trim().toLowerCase()

    if (["confirmationtokenhash", "confirmation_token_hash"].includes(fieldName)) {
      hasConfirmationTokenHash = true
    }

    if (["confirmationtokenexpiresat", "confirmation_token_expires_at"].includes(fieldName)) {
      hasConfirmationTokenExpiresAt = true
    }

    if (["unsubscribetokenhash", "unsubscribe_token_hash"].includes(fieldName)) {
      hasUnsubscribeTokenHash = true
    }

    if (["unsubscribedat", "unsubscribed_at"].includes(fieldName)) {
      hasUnsubscribedAt = true
    }
  }

  if (!hasConfirmationTokenHash) {
    collection.fields.addAt(collection.fields.length, new Field({
      "autogeneratePattern": "",
      "hidden": false,
      "id": "text1779303901",
      "max": 0,
      "min": 0,
      "name": "confirmationTokenHash",
      "pattern": "",
      "presentable": false,
      "primaryKey": false,
      "required": false,
      "system": false,
      "type": "text"
    }))
  }

  if (!hasConfirmationTokenExpiresAt) {
    collection.fields.addAt(collection.fields.length, new Field({
      "hidden": false,
      "id": "date1779303902",
      "max": "",
      "min": "",
      "name": "confirmationTokenExpiresAt",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "date"
    }))
  }

  if (!hasUnsubscribeTokenHash) {
    collection.fields.addAt(collection.fields.length, new Field({
      "autogeneratePattern": "",
      "hidden": false,
      "id": "text1779303903",
      "max": 0,
      "min": 0,
      "name": "unsubscribeTokenHash",
      "pattern": "",
      "presentable": false,
      "primaryKey": false,
      "required": false,
      "system": false,
      "type": "text"
    }))
  }

  if (!hasUnsubscribedAt) {
    collection.fields.addAt(collection.fields.length, new Field({
      "hidden": false,
      "id": "date1779303904",
      "max": "",
      "min": "",
      "name": "unsubscribedAt",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "date"
    }))
  }

  return app.save(collection)
  // NUVIO CUSTOM END: Newsletter lifecycle token/timestamp fields.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203400")
  const fields = collection.fields || []

  let hasConfirmationTokenHash = false
  let hasConfirmationTokenExpiresAt = false
  let hasUnsubscribeTokenHash = false
  let hasUnsubscribedAt = false

  for (let i = 0; i < fields.length; i += 1) {
    const field = fields[i]
    const fieldId = `${field?.id || ""}`.trim()

    if (fieldId === "text1779303901") {
      hasConfirmationTokenHash = true
    }

    if (fieldId === "date1779303902") {
      hasConfirmationTokenExpiresAt = true
    }

    if (fieldId === "text1779303903") {
      hasUnsubscribeTokenHash = true
    }

    if (fieldId === "date1779303904") {
      hasUnsubscribedAt = true
    }
  }

  if (hasUnsubscribedAt) {
    collection.fields.removeById("date1779303904")
  }

  if (hasUnsubscribeTokenHash) {
    collection.fields.removeById("text1779303903")
  }

  if (hasConfirmationTokenExpiresAt) {
    collection.fields.removeById("date1779303902")
  }

  if (hasConfirmationTokenHash) {
    collection.fields.removeById("text1779303901")
  }

  return app.save(collection)
})
