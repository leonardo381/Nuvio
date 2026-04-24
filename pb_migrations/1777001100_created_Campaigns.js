/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Newsletter V1 campaigns storage collection.
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
        "id": "relation1661203501",
        "maxSelect": 1,
        "minSelect": 1,
        "name": "website",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "relation"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203502",
        "max": 0,
        "min": 0,
        "name": "subject",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": true,
        "system": false,
        "type": "text"
      },
      {
        "convertURLs": false,
        "hidden": false,
        "id": "editor1661203503",
        "maxSize": 0,
        "name": "body",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "editor"
      },
      {
        "hidden": false,
        "id": "select1661203504",
        "maxSelect": 1,
        "name": "status",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "select",
        "values": [
          "draft",
          "sent"
        ]
      },
      {
        "hidden": false,
        "id": "select1661203505",
        "maxSelect": 1,
        "name": "recipientsType",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "select",
        "values": [
          "all",
          "manual"
        ]
      },
      {
        "hidden": false,
        "id": "json1661203506",
        "maxSize": 0,
        "name": "recipientsIds",
        "presentable": false,
        "required": false,
        "system": false,
        "type": "json"
      },
      {
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
      },
      {
        "hidden": false,
        "id": "date1661203508",
        "max": "",
        "min": "",
        "name": "sentAt",
        "presentable": false,
        "required": false,
        "system": false,
        "type": "date"
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
    "id": "pbc_1661203500",
    "indexes": [],
    "listRule": null,
    "name": "Campaigns",
    "system": false,
    "type": "base",
    "updateRule": null,
    "viewRule": null
  });

  return app.save(collection);
  // NUVIO CUSTOM END: Newsletter V1 campaigns storage collection.
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1661203500");

  return app.delete(collection);
})
