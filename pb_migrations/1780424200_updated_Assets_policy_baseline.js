/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("pbc_1321337024")

  const findFieldIndexByName = (name) => {
    for (let index = 0; index < collection.fields.length; index++) {
      if (collection.fields[index]?.name === name) {
        return index
      }
    }

    return -1
  }

  const hasFieldById = (id) => {
    for (let index = 0; index < collection.fields.length; index++) {
      if (collection.fields[index]?.id === id) {
        return true
      }
    }

    return false
  }

  if (!hasFieldById("relation1780424201")) {
    collection.fields.addAt(collection.fields.length, new Field({
      "cascadeDelete": false,
      "collectionId": "pbc_2619338178",
      "hidden": false,
      "id": "relation1780424201",
      "maxSelect": 1,
      "minSelect": 0,
      "name": "website",
      "presentable": false,
      "required": false,
      "system": false,
      "type": "relation"
    }))
  }

  const fileFieldIndex = findFieldIndexByName("file")
  if (fileFieldIndex >= 0) {
    collection.fields.addAt(fileFieldIndex, new Field({
      "hidden": false,
      "id": "file2359244304",
      "maxSelect": 1,
      "maxSize": 8388608,
      "mimeTypes": [
        "image/jpeg",
        "image/png",
        "image/webp",
        "image/gif"
      ],
      "name": "file",
      "presentable": false,
      "protected": false,
      "required": false,
      "system": false,
      "thumbs": [],
      "type": "file"
    }))
  }

  return app.save(collection)
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_1321337024")

  const findFieldIndexByName = (name) => {
    for (let index = 0; index < collection.fields.length; index++) {
      if (collection.fields[index]?.name === name) {
        return index
      }
    }

    return -1
  }

  const hasFieldById = (id) => {
    for (let index = 0; index < collection.fields.length; index++) {
      if (collection.fields[index]?.id === id) {
        return true
      }
    }

    return false
  }

  if (hasFieldById("relation1780424201")) {
    collection.fields.removeById("relation1780424201")
  }

  const fileFieldIndex = findFieldIndexByName("file")
  if (fileFieldIndex >= 0) {
    collection.fields.addAt(fileFieldIndex, new Field({
      "hidden": false,
      "id": "file2359244304",
      "maxSelect": 1,
      "maxSize": 0,
      "mimeTypes": [],
      "name": "file",
      "presentable": false,
      "protected": false,
      "required": false,
      "system": false,
      "thumbs": [],
      "type": "file"
    }))
  }

  return app.save(collection)
})
