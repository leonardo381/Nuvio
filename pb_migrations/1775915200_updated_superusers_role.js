/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("_superusers")

  // add field
  collection.fields.addAt(collection.fields.length, new Field({
    "hidden": false,
    "id": "select3921814201",
    "maxSelect": 1,
    "name": "role",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "select",
    "values": [
      "admin",
      "client"
    ]
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("_superusers")

  // remove field
  collection.fields.removeById("select3921814201")

  return app.save(collection)
})
