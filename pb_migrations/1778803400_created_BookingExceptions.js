/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Booking Phase 5F.1 availability exceptions foundation.
  const bookingExceptions = new Collection({
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
        "id": "relation1778803401",
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
        "id": "text1778803402",
        "max": 10,
        "min": 10,
        "name": "date",
        "pattern": "^\\d{4}-\\d{2}-\\d{2}$",
        "presentable": false,
        "primaryKey": false,
        "required": true,
        "system": false,
        "type": "text"
      },
      {
        "hidden": false,
        "id": "select1778803403",
        "maxSelect": 1,
        "name": "type",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "select",
        "values": [
          "closed",
          "customHours"
        ]
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1778803404",
        "max": 5,
        "min": 0,
        "name": "startTime",
        "pattern": "^([01]\\d|2[0-3]):[0-5]\\d$",
        "presentable": false,
        "primaryKey": false,
        "required": false,
        "system": false,
        "type": "text"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1778803405",
        "max": 5,
        "min": 0,
        "name": "endTime",
        "pattern": "^([01]\\d|2[0-3]):[0-5]\\d$",
        "presentable": false,
        "primaryKey": false,
        "required": false,
        "system": false,
        "type": "text"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1778803406",
        "max": 0,
        "min": 0,
        "name": "note",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": false,
        "system": false,
        "type": "text"
      },
      {
        "hidden": false,
        "id": "bool1778803407",
        "name": "active",
        "presentable": false,
        "required": false,
        "system": false,
        "type": "bool"
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
      },
      {
        "hidden": false,
        "id": "autodate3332085495",
        "name": "updated",
        "onCreate": true,
        "onUpdate": true,
        "presentable": false,
        "system": false,
        "type": "autodate"
      }
    ],
    "id": "pbc_1778803400",
    "indexes": [],
    "listRule": null,
    "name": "BookingExceptions",
    "system": false,
    "type": "base",
    "updateRule": null,
    "viewRule": null
  })

  return app.save(bookingExceptions)
  // NUVIO CUSTOM END: Booking Phase 5F.1 availability exceptions foundation.
}, (app) => {
  const bookingExceptions = app.findCollectionByNameOrId("pbc_1778803400")
  return app.delete(bookingExceptions)
})
