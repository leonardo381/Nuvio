/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_184785686")

  // update collection data
  unmarshal({
    "name": "components"
  }, collection)

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_184785686")

  // update collection data
  unmarshal({
    "name": "templates"
  }, collection)

  return app.save(collection)
})
