/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const websites = app.findCollectionByNameOrId("pbc_2619338178")
  unmarshal({
    "createRule": null,
    "updateRule": null,
    "deleteRule": null
  }, websites)
  app.save(websites)

  const pages = app.findCollectionByNameOrId("pbc_3945946014")
  unmarshal({
    "createRule": null,
    "updateRule": null,
    "deleteRule": null
  }, pages)
  app.save(pages)

  const blocks = app.findCollectionByNameOrId("pbc_4194232374")
  unmarshal({
    "createRule": null,
    "updateRule": null,
    "deleteRule": null
  }, blocks)
  app.save(blocks)

  const components = app.findCollectionByNameOrId("pbc_184785686")
  unmarshal({
    "createRule": null,
    "updateRule": null,
    "deleteRule": null
  }, components)
  app.save(components)

  const assets = app.findCollectionByNameOrId("pbc_1321337024")
  unmarshal({
    "createRule": null,
    "updateRule": null,
    "deleteRule": null
  }, assets)

  return app.save(assets)
}, (app) => {
  const websites = app.findCollectionByNameOrId("pbc_2619338178")
  unmarshal({
    "createRule": null,
    "updateRule": null,
    "deleteRule": null
  }, websites)
  app.save(websites)

  const pages = app.findCollectionByNameOrId("pbc_3945946014")
  unmarshal({
    "createRule": null,
    "updateRule": null,
    "deleteRule": null
  }, pages)
  app.save(pages)

  const blocks = app.findCollectionByNameOrId("pbc_4194232374")
  unmarshal({
    "createRule": null,
    "updateRule": null,
    "deleteRule": null
  }, blocks)
  app.save(blocks)

  const components = app.findCollectionByNameOrId("pbc_184785686")
  unmarshal({
    "createRule": null,
    "updateRule": null,
    "deleteRule": null
  }, components)
  app.save(components)

  const assets = app.findCollectionByNameOrId("pbc_1321337024")
  unmarshal({
    "createRule": null,
    "updateRule": null,
    "deleteRule": null
  }, assets)

  return app.save(assets)
})
