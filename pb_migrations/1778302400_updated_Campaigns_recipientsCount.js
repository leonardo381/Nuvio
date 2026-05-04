/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: allow non-negative recipientsCount values for newsletter campaigns.
  const collection = app.findCollectionByNameOrId("pbc_1661203500")

  // update field
  collection.fields.addAt(7, new Field({
    "hidden": false,
    "id": "number1661203507",
    "max": null,
    "min": 0,
    "name": "recipientsCount",
    "onlyInt": true,
    "presentable": false,
    "required": false,
    "system": false,
    "type": "number"
  }))

  return app.save(collection)
  // NUVIO CUSTOM END: allow non-negative recipientsCount values for newsletter campaigns.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203500")

  // restore previous validation
  collection.fields.addAt(7, new Field({
    "hidden": false,
    "id": "number1661203507",
    "max": 0,
    "min": 0,
    "name": "recipientsCount",
    "onlyInt": true,
    "presentable": false,
    "required": false,
    "system": false,
    "type": "number"
  }))

  return app.save(collection)
})
