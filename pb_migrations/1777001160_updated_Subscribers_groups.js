/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: add multi-group relation for newsletter subscribers.
  const collection = app.findCollectionByNameOrId("pbc_1661203400")

  collection.fields.addAt(collection.fields.length, new Field({
    "cascadeDelete": false,
    "collectionId": "pbc_1661203600",
    "hidden": false,
    "id": "relation1661203410",
    "maxSelect": 0,
    "minSelect": 0,
    "name": "groups",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "relation"
  }))

  return app.save(collection)
  // NUVIO CUSTOM END: add multi-group relation for newsletter subscribers.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203400")

  collection.fields.removeById("relation1661203410")

  return app.save(collection)
})
