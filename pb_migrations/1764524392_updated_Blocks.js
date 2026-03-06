/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_4194232374")

  // add field
  collection.fields.addAt(1, new Field({
    "cascadeDelete": false,
    "collectionId": "pbc_3945946014",
    "hidden": false,
    "id": "relation336246304",
    "maxSelect": 1,
    "minSelect": 0,
    "name": "page",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "relation"
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_4194232374")

  // remove field
  collection.fields.removeById("relation336246304")

  return app.save(collection)
})
