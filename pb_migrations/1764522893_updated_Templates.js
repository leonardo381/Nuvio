/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_184785686")

  // update collection data
  unmarshal({
    "name": "Components"
  }, collection)

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_184785686")

  // update collection data
  unmarshal({
    "name": "Templates"
  }, collection)

  return app.save(collection)
})
