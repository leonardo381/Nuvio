/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203100")

  // update collection data
  unmarshal({
    "createRule": "@request.auth.collectionName = \"users\" && @request.auth.id = \"jsevkudq5m7tgfz\""
  }, collection)

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203100")

  // update collection data
  unmarshal({
    "createRule": null
  }, collection)

  return app.save(collection)
})
