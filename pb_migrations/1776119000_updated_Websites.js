/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_2619338178")

  // NUVIO CUSTOM: add website-level settings JSON storage field.
  collection.fields.addAt(8, new Field({
    "hidden": false,
    "id": "json2964215044",
    "maxSize": 0,
    "name": "settings",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "json"
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_2619338178")

  // revert website-level settings JSON storage field.
  collection.fields.removeById("json2964215044")

  return app.save(collection)
})
