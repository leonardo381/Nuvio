/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_184785686")

  // update field
  collection.fields.addAt(3, new Field({
    "hidden": false,
    "id": "select1542800728",
    "maxSelect": 1,
    "name": "gallery",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "select",
    "values": [
      "flowbite",
      "tailwindUI"
    ]
  }))

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_184785686")

  // update field
  collection.fields.addAt(3, new Field({
    "hidden": false,
    "id": "select1542800728",
    "maxSelect": 1,
    "name": "field",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "select",
    "values": [
      "flowbite",
      "tailwindUI"
    ]
  }))

  return app.save(collection)
})
