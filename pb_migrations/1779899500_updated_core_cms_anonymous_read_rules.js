/// <reference path="../pb_data/types.d.ts" />
function findFirstExistingCollection(app, identifiers) {
  for (const identifier of identifiers) {
    if (!identifier) {
      continue
    }

    const collection = app.findCollectionByNameOrId(identifier)
    if (collection) {
      return collection
    }
  }

  return null
}

function requireCollection(app, label, identifiers) {
  const collection = findFirstExistingCollection(app, identifiers)
  if (!collection) {
    throw new Error("missing required core CMS collection: " + label + " (tried: " + identifiers.join(", ") + ")")
  }
  return collection
}

migrate((app) => {
  // NUVIO CUSTOM START: Security A3.3 lock anonymous raw reads for core CMS collections.
  const websitesCollection = requireCollection(app, "Websites", ["pbc_2619338178", "Websites", "websites"])
  const pagesCollection = requireCollection(app, "Pages", ["pbc_3945946014", "Pages", "pages"])
  const blocksCollection = requireCollection(app, "Blocks", ["pbc_4194232374", "Blocks", "blocks"])
  const componentsCollection = requireCollection(app, "Components", ["pbc_184785686", "Components", "components"])

  unmarshal({
    "listRule": "@request.auth.id != \"\"",
    "viewRule": "@request.auth.id != \"\""
  }, websitesCollection)
  app.save(websitesCollection)

  unmarshal({
    "listRule": "@request.auth.id != \"\"",
    "viewRule": "@request.auth.id != \"\""
  }, pagesCollection)
  app.save(pagesCollection)

  unmarshal({
    "listRule": "@request.auth.id != \"\"",
    "viewRule": "@request.auth.id != \"\""
  }, blocksCollection)
  app.save(blocksCollection)

  unmarshal({
    "listRule": "@request.auth.id != \"\"",
    "viewRule": "@request.auth.id != \"\""
  }, componentsCollection)
  app.save(componentsCollection)
  // NUVIO CUSTOM END: Security A3.3 lock anonymous raw reads for core CMS collections.
}, (app) => {
  const websitesCollection = requireCollection(app, "Websites", ["pbc_2619338178", "Websites", "websites"])
  const pagesCollection = requireCollection(app, "Pages", ["pbc_3945946014", "Pages", "pages"])
  const blocksCollection = requireCollection(app, "Blocks", ["pbc_4194232374", "Blocks", "blocks"])
  const componentsCollection = requireCollection(app, "Components", ["pbc_184785686", "Components", "components"])

  unmarshal({
    "listRule": "",
    "viewRule": ""
  }, websitesCollection)
  app.save(websitesCollection)

  unmarshal({
    "listRule": "",
    "viewRule": ""
  }, pagesCollection)
  app.save(pagesCollection)

  unmarshal({
    "listRule": "",
    "viewRule": ""
  }, blocksCollection)
  app.save(blocksCollection)

  unmarshal({
    "listRule": "",
    "viewRule": ""
  }, componentsCollection)
  app.save(componentsCollection)
})
