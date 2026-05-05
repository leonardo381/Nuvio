/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: extend Contacts status to support archived inbox workflow.
  const collection = app.findCollectionByNameOrId("pbc_1661203100")

  // update field
  collection.fields.addAt(8, new Field({
    "hidden": false,
    "id": "select1661203107",
    "maxSelect": 1,
    "name": "status",
    "presentable": false,
    "required": true,
    "system": false,
    "type": "select",
    "values": [
      "new",
      "read",
      "archived"
    ]
  }))

  return app.save(collection)
  // NUVIO CUSTOM END: extend Contacts status to support archived inbox workflow.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203100")

  // restore previous status values
  collection.fields.addAt(8, new Field({
    "hidden": false,
    "id": "select1661203107",
    "maxSelect": 1,
    "name": "status",
    "presentable": false,
    "required": true,
    "system": false,
    "type": "select",
    "values": [
      "new",
      "read"
    ]
  }))

  return app.save(collection)
})
