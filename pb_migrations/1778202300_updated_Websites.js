/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_2619338178")

  const hasFieldByName = (name) => {
    for (let index = 0; index < collection.fields.length; index++) {
      if (collection.fields[index]?.name === name) {
        return true
      }
    }

    return false
  }

  const addFieldIfMissing = (name, config) => {
    if (hasFieldByName(name)) {
      return
    }

    collection.fields.addAt(collection.fields.length, new Field(config))
  }

  // add field
  addFieldIfMissing("business_name", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209101",
    "max": 0,
    "min": 0,
    "name": "business_name",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_type", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209102",
    "max": 0,
    "min": 0,
    "name": "business_type",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_primary_category", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209103",
    "max": 0,
    "min": 0,
    "name": "business_primary_category",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_phone", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209104",
    "max": 0,
    "min": 0,
    "name": "business_phone",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_email", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209105",
    "max": 0,
    "min": 0,
    "name": "business_email",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_address", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209106",
    "max": 0,
    "min": 0,
    "name": "business_address",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_city", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209107",
    "max": 0,
    "min": 0,
    "name": "business_city",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_postal_code", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209108",
    "max": 0,
    "min": 0,
    "name": "business_postal_code",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_country", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209109",
    "max": 0,
    "min": 0,
    "name": "business_country",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_service_area", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209110",
    "max": 0,
    "min": 0,
    "name": "business_service_area",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_opening_hours", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209111",
    "max": 0,
    "min": 0,
    "name": "business_opening_hours",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_google_place_id", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209112",
    "max": 0,
    "min": 0,
    "name": "business_google_place_id",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_social_profiles", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209113",
    "max": 0,
    "min": 0,
    "name": "business_social_profiles",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("business_price_range", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714209114",
    "max": 0,
    "min": 0,
    "name": "business_price_range",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_2619338178")

  const hasFieldById = (id) => {
    for (let index = 0; index < collection.fields.length; index++) {
      if (collection.fields[index]?.id === id) {
        return true
      }
    }

    return false
  }

  const removeFieldIfExists = (id) => {
    if (hasFieldById(id)) {
      collection.fields.removeById(id)
    }
  }

  // remove field
  removeFieldIfExists("text2714209101")

  // remove field
  removeFieldIfExists("text2714209102")

  // remove field
  removeFieldIfExists("text2714209103")

  // remove field
  removeFieldIfExists("text2714209104")

  // remove field
  removeFieldIfExists("text2714209105")

  // remove field
  removeFieldIfExists("text2714209106")

  // remove field
  removeFieldIfExists("text2714209107")

  // remove field
  removeFieldIfExists("text2714209108")

  // remove field
  removeFieldIfExists("text2714209109")

  // remove field
  removeFieldIfExists("text2714209110")

  // remove field
  removeFieldIfExists("text2714209111")

  // remove field
  removeFieldIfExists("text2714209112")

  // remove field
  removeFieldIfExists("text2714209113")

  // remove field
  removeFieldIfExists("text2714209114")

  return app.save(collection)
})
