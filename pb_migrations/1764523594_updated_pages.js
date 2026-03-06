/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_3945946014")

  // update collection data
  unmarshal({
    "name": "Pages"
  }, collection)

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_3945946014")

  // update collection data
  unmarshal({
    "name": "pages"
  }, collection)

  return app.save(collection)
})
