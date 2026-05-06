/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  // NUVIO CUSTOM START: Booking MVP Phase 1 collections.
  const bookingServices = new Collection({
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
        "id": "relation1661203701",
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
        "id": "text1661203702",
        "max": 0,
        "min": 1,
        "name": "name",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": true,
        "system": false,
        "type": "text"
      },
      {
        "hidden": false,
        "id": "number1661203703",
        "max": 480,
        "min": 5,
        "name": "durationMinutes",
        "onlyInt": true,
        "presentable": false,
        "required": true,
        "system": false,
        "type": "number"
      },
      {
        "hidden": false,
        "id": "bool1661203704",
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
    "id": "pbc_1661203700",
    "indexes": [],
    "listRule": null,
    "name": "BookingServices",
    "system": false,
    "type": "base",
    "updateRule": null,
    "viewRule": null
  })

  const bookingAvailability = new Collection({
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
        "id": "relation1661203801",
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
        "id": "select1661203802",
        "maxSelect": 1,
        "name": "dayOfWeek",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "select",
        "values": [
          "mon",
          "tue",
          "wed",
          "thu",
          "fri",
          "sat",
          "sun"
        ]
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203803",
        "max": 5,
        "min": 5,
        "name": "startTime",
        "pattern": "^([01]\\d|2[0-3]):[0-5]\\d$",
        "presentable": false,
        "primaryKey": false,
        "required": true,
        "system": false,
        "type": "text"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203804",
        "max": 5,
        "min": 5,
        "name": "endTime",
        "pattern": "^([01]\\d|2[0-3]):[0-5]\\d$",
        "presentable": false,
        "primaryKey": false,
        "required": true,
        "system": false,
        "type": "text"
      },
      {
        "hidden": false,
        "id": "bool1661203805",
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
    "id": "pbc_1661203800",
    "indexes": [],
    "listRule": null,
    "name": "BookingAvailability",
    "system": false,
    "type": "base",
    "updateRule": null,
    "viewRule": null
  })

  const appointments = new Collection({
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
        "id": "relation1661203901",
        "maxSelect": 1,
        "minSelect": 1,
        "name": "website",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "relation"
      },
      {
        "cascadeDelete": false,
        "collectionId": "pbc_1661203700",
        "hidden": false,
        "id": "relation1661203902",
        "maxSelect": 1,
        "minSelect": 1,
        "name": "service",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "relation"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203903",
        "max": 0,
        "min": 1,
        "name": "name",
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
        "id": "text1661203904",
        "max": 0,
        "min": 3,
        "name": "email",
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
        "id": "text1661203905",
        "max": 0,
        "min": 0,
        "name": "phone",
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
        "id": "text1661203906",
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
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203907",
        "max": 5,
        "min": 5,
        "name": "time",
        "pattern": "^([01]\\d|2[0-3]):[0-5]\\d$",
        "presentable": false,
        "primaryKey": false,
        "required": true,
        "system": false,
        "type": "text"
      },
      {
        "hidden": false,
        "id": "select1661203908",
        "maxSelect": 1,
        "name": "status",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "select",
        "values": [
          "pending",
          "confirmed",
          "cancelled"
        ]
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text1661203909",
        "max": 0,
        "min": 0,
        "name": "notes",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": false,
        "system": false,
        "type": "text"
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
    "id": "pbc_1661203900",
    "indexes": [],
    "listRule": null,
    "name": "Appointments",
    "system": false,
    "type": "base",
    "updateRule": null,
    "viewRule": null
  })

  app.save(bookingServices)
  app.save(bookingAvailability)
  return app.save(appointments)
  // NUVIO CUSTOM END: Booking MVP Phase 1 collections.
}, (app) => {
  const appointments = app.findCollectionByNameOrId("pbc_1661203900")
  app.delete(appointments)

  const bookingAvailability = app.findCollectionByNameOrId("pbc_1661203800")
  app.delete(bookingAvailability)

  const bookingServices = app.findCollectionByNameOrId("pbc_1661203700")
  return app.delete(bookingServices)
})
