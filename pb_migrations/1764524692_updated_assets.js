/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_1321337024")

  // update collection data
  unmarshal({
    "name": "Assets"
  }, collection)

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1321337024")

  // update collection data
  unmarshal({
    "name": "assets"
  }, collection)

  return app.save(collection)
})
