/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_2619338178")

  // add field
  collection.fields.addAt(5, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text1258567589",
    "max": 0,
    "min": 0,
    "name": "seoTitle",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  // add field
  collection.fields.addAt(6, new Field({
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2387610480",
    "max": 0,
    "min": 0,
    "name": "seoDescription",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  }))

  // add field
  collection.fields.addAt(7, new Field({
    "hidden": false,
    "id": "file2769243025",
    "maxSelect": 1,
    "maxSize": 0,
    "mimeTypes": [],
    "name": "seoImage",
    "presentable": false,
    "protected": false,
    "required": false,
    "system": false,
    "thumbs": [],
    "type": "file"
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_2619338178")

  // remove field
  collection.fields.removeById("text1258567589")

  // remove field
  collection.fields.removeById("text2387610480")

  // remove field
  collection.fields.removeById("file2769243025")

  return app.save(collection)
})
