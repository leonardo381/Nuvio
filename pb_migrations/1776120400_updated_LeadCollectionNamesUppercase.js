/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const contacts = app.findCollectionByNameOrId("pbc_1661203100")

  // update collection data
  unmarshal({
    "name": "Contacts"
  }, contacts)
  app.save(contacts)

  const whatsapp = app.findCollectionByNameOrId("pbc_1661203200")

  // update collection data
  unmarshal({
    "name": "Whatsapp"
  }, whatsapp)

  return app.save(whatsapp)
}, (app) => {
  const contacts = app.findCollectionByNameOrId("pbc_1661203100")

  // update collection data
  unmarshal({
    "name": "contacts"
  }, contacts)
  app.save(contacts)

  const whatsapp = app.findCollectionByNameOrId("pbc_1661203200")

  // update collection data
  unmarshal({
    "name": "whatsapp_interactions"
  }, whatsapp)

  return app.save(whatsapp)
})
