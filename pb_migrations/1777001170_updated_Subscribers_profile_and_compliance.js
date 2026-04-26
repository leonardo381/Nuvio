/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: add optional subscriber name field for newsletter.
  const collection = app.findCollectionByNameOrId("pbc_1661203400")

  collection.fields.addAt(collection.fields.length, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text1661203411",
    "max": 0,
    "min": 0,
    "name": "name",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  return app.save(collection)
  // NUVIO CUSTOM END: add optional subscriber name field for newsletter.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203400")

  collection.fields.removeById("text1661203411")

  return app.save(collection)
})
