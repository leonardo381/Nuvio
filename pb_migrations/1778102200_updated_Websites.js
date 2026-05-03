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
  addFieldIfMissing("seo_title_template", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714109056",
    "max": 0,
    "min": 0,
    "name": "seo_title_template",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("seo_title_separator", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714109057",
    "max": 0,
    "min": 0,
    "name": "seo_title_separator",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("seo_canonical_domain", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714109058",
    "max": 0,
    "min": 0,
    "name": "seo_canonical_domain",
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
  removeFieldIfExists("text2714109056")

  // remove field
  removeFieldIfExists("text2714109057")

  // remove field
  removeFieldIfExists("text2714109058")

  return app.save(collection)
})
