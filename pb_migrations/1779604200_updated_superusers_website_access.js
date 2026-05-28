/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("_superusers")

  // add field
  collection.fields.addAt(collection.fields.length, new Field({
    "cascadeDelete": false,
    "collectionId": "pbc_2619338178",
    "hidden": false,
    "id": "relation1779604200",
    "maxSelect": 999,
    "minSelect": 0,
    "name": "websiteAccess",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "relation"
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("_superusers")

  // remove field
  collection.fields.removeById("relation1779604200")

  return app.save(collection)
})

