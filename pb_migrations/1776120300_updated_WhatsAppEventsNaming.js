/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203200")

  // update collection data
  unmarshal({
    "name": "whatsapp_interactions"
  }, collection)

  // add/update field
  collection.fields.addAt(2, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text1661203202",
    "max": 0,
    "min": 0,
    "name": "source",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  // add/update field
  collection.fields.addAt(3, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text1661203203",
    "max": 0,
    "min": 0,
    "name": "page",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203200")

  // remove field
  collection.fields.removeById("text1661203202")

  // remove field
  collection.fields.removeById("text1661203203")

  // update collection data
  unmarshal({
    "name": "whatsapp_clicks"
  }, collection)

  return app.save(collection)
})
