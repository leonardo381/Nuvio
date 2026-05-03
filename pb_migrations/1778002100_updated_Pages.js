/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_3945946014")

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
  addFieldIfMissing("seo_social_image", {
    "hidden": false,
    "id": "file2714109041",
    "maxSelect": 1,
    "maxSize": 0,
    "mimeTypes": [],
    "name": "seo_social_image",
    "presentable": false,
    "protected": false,
    "required": false,
    "system": false,
    "thumbs": [],
    "type": "file"
  })

  // add field
  addFieldIfMissing("seo_canonical_url", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714109042",
    "max": 0,
    "min": 0,
    "name": "seo_canonical_url",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  // add field
  addFieldIfMissing("seo_noindex", {
    "hidden": false,
    "id": "bool2714109043",
    "name": "seo_noindex",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "bool"
  })

  // add field
  addFieldIfMissing("seo_exclude_from_sitemap", {
    "hidden": false,
    "id": "bool2714109044",
    "name": "seo_exclude_from_sitemap",
    "presentable": false,
    "required": false,
    "system": false,
    "type": "bool"
  })

  // add field
  addFieldIfMissing("seo_focus_keyword", {
    "autogeneratePattern": "",
    "hidden": false,
    "id": "text2714109045",
    "max": 0,
    "min": 0,
    "name": "seo_focus_keyword",
    "pattern": "",
    "presentable": false,
    "primaryKey": false,
    "required": false,
    "system": false,
    "type": "text"
  })

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_3945946014")

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
  removeFieldIfExists("file2714109041")

  // remove field
  removeFieldIfExists("text2714109042")

  // remove field
  removeFieldIfExists("bool2714109043")

  // remove field
  removeFieldIfExists("bool2714109044")

  // remove field
  removeFieldIfExists("text2714109045")

  return app.save(collection)
})
