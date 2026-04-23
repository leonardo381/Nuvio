/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Collection-backed dashboard Reviews module storage.
  const collection = new Collection({
    "createRule": null,
    "deleteRule": null,
    "fields": [
      {
        "autogeneratePattern": "[a-z0-9]{15}",
        "hidden": false,
        "id": "text3208210256",
        "max": 15,
        "min": 15,
        "name": "id",
        "pattern": "^[a-z0-9]+$",
        "presentable": false,
        "primaryKey": true,
        "required": true,
        "system": true,
        "type": "text"
      },
      {
        "cascadeDelete": false,
        "collectionId": "pbc_2619338178",
        "hidden": false,
        "id": "relation1661203301",
        "maxSelect": 1,
        "minSelect": 1,
        "name": "website",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "relation"
      },
      {
        "hidden": false,
        "id": "select1661203302",
        "maxSelect": 1,
        "name": "source",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "select",
        "values": [
          "google"
        ]
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203303",
        "max": 0,
        "min": 0,
        "name": "externalId",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": true,
        "system": false,
        "type": "text"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203304",
        "max": 0,
        "min": 0,
        "name": "authorName",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": true,
        "system": false,
        "type": "text"
      },
      {
        "hidden": false,
        "id": "number1661203305",
        "max": 5,
        "min": 0,
        "name": "rating",
        "onlyInt": false,
        "presentable": false,
        "required": true,
        "system": false,
        "type": "number"
      },
      {
        "convertURLs": false,
        "hidden": false,
        "id": "editor1661203306",
        "maxSize": 0,
        "name": "text",
        "presentable": false,
        "required": false,
        "system": false,
        "type": "editor"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203307",
        "max": 0,
        "min": 0,
        "name": "publishedAt",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": false,
        "system": false,
        "type": "text"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203308",
        "max": 0,
        "min": 0,
        "name": "relativeTimeText",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": false,
        "system": false,
        "type": "text"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203309",
        "max": 0,
        "min": 0,
        "name": "reviewerPhotoUrl",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": false,
        "system": false,
        "type": "text"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203310",
        "max": 0,
        "min": 0,
        "name": "reviewUrl",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": false,
        "system": false,
        "type": "text"
      },
      {
        "hidden": false,
        "id": "autodate1661203311",
        "name": "syncedAt",
        "onCreate": true,
        "onUpdate": true,
        "presentable": false,
        "system": false,
        "type": "autodate"
      },
      {
        "hidden": false,
        "id": "autodate2990389176",
        "name": "created",
        "onCreate": true,
        "onUpdate": false,
        "presentable": false,
        "system": false,
        "type": "autodate"
      }
    ],
    "id": "pbc_1661203300",
    "indexes": [],
    "listRule": null,
    "name": "Reviews",
    "system": false,
    "type": "base",
    "updateRule": null,
    "viewRule": null
  });

  return app.save(collection);
  // NUVIO CUSTOM END: Collection-backed dashboard Reviews module storage.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203300");

  return app.delete(collection);
})
