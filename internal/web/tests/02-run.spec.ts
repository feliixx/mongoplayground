import { expect } from '@playwright/test';
import { test, setEditorContent, expectEditorContent } from './playwright'

test('run default page with button', async ({ page }) => {
  await expectEditorContent('result', '')

  await page.getByRole('button', { name: 'run' }).click()
  await expectEditorContent('result', `[
  {
    "_id": ObjectId("5a934e000102030405000000"),
    "key": 1
  },
  {
    "_id": ObjectId("5a934e000102030405000001"),
    "key": 2
  }
]`)
})

test('run default page with shortcut', async ({ page }) => {
  await expectEditorContent('result', '')

  await page.getByText('Template').press('Control+Enter')
  await expectEditorContent('result', `[
  {
    "_id": ObjectId("5a934e000102030405000000"),
    "key": 1
  },
  {
    "_id": ObjectId("5a934e000102030405000001"),
    "key": 2
  }
]`)
})

test('run incorrect config', async ({ page }) => {

  await setEditorContent('config', `[
  {
    a: invalid
  }
]`)

  await page.locator('#config').getByText('3').hover()
  await expect(page.locator('#config > .ace_tooltip')).toBeVisible()
  await expect(page.locator('#config > .ace_tooltip')).toHaveText(`Unknown type: 'invalid'`)
  await page.getByRole('button', { name: 'run' }).click()

  await expect(page.locator('#resultPanel')).toHaveClass('text_red')
  await expectEditorContent('result', `Invalid configuration:
Line 3: Unknown type: 'invalid'`)
})

test('run no result', async ({ page }) => {

  await setEditorContent('config', `[{k:1}]`)
  await setEditorContent('query', `db.collection.find({k:2})`)
  await page.getByRole('button', { name: 'run' }).click()

  await expect(page.locator('#resultPanel')).not.toHaveClass('text_red')
  await expectEditorContent('result', 'no document found')
})

test('aggregation query without stages', async ({ page }) => {
  await setEditorContent('query', `db.collection.aggregate([{}])`)
  await expect(page.getByText('Stage:')).toBeHidden()
  await expect(page.locator('#custom-aggregation_stages')).toBeHidden()
})

test('aggregation query with stages', async ({ page }) => {

  await setEditorContent('query', `db.collection.aggregate([
    {
      "$project": {
        "_id": 0
      }
    },
    {
      "$match": {
        "key": 1
      }
    }
  ])`)

  await page.getByRole('button', { name: '$match' }).click()
  await expect(page.locator('#custom-aggregation_stages ul').getByText('$project')).toBeVisible()
  await expect(page.locator('#custom-aggregation_stages ul').getByText('$match')).toBeVisible()

  await page.getByRole('button', { name: 'run' }).click()
  await expectEditorContent('result', `[
  {
    "key": 1
  }
]`)

  await page.getByRole('button', { name: '$match' }).click()
  await page.locator('#custom-aggregation_stages ul').getByText('$project').click()

  await page.getByRole('button', { name: 'run' }).click()
  await expectEditorContent('result', `[
  {
    "key": 1
  },
  {
    "key": 2
  }
]`)
})

test('test single db template', async ({ page }) => {

  await setEditorContent('config', '')
  await setEditorContent('query', '')
  await page.getByRole('button', { name: 'single collection' }).click()
  await page.locator('#custom-template ul').getByText('single collection').click()

  await expect(page.locator('#custom-mode').getByRole('button', { name: 'bson' })).toBeVisible();
  await expectEditorContent('config', `[
  {
    "key": 1
  },
  {
    "key": 2
  }
]`)

  await expectEditorContent('query', `db.collection.find()`)
  await expect(page.locator('#custom-aggregation_stage')).toBeHidden();

  await expectEditorContent('result', '')
  await page.getByRole('button', { name: 'run' }).click()
  await expectEditorContent('result', `[
  {
    "_id": ObjectId("5a934e000102030405000000"),
    "key": 1
  },
  {
    "_id": ObjectId("5a934e000102030405000001"),
    "key": 2
  }
]`)
})

test('test multiple db template', async ({ page }) => {

  await page.getByRole('button', { name: 'single collection' }).click()
  await page.locator('#custom-template ul').getByText('multiple collection').click()

  await expect(page.locator('#custom-mode').getByRole('button', { name: 'bson' })).toBeVisible();
  await expectEditorContent('config', `db={
  "orders": [
    {
      "_id": 1,
      "item": "almonds",
      "price": 12,
      "quantity": 2
    },
    {
      "_id": 2,
      "item": "pecans",
      "price": 20,
      "quantity": 1
    },
    {
      "_id": 3
    }
  ],
  "inventory": [
    {
      "_id": 1,
      "sku": "almonds",
      "description": "product 1",
      "instock": 120
    },
    {
      "_id": 2,
      "sku": "bread",
      "description": "product 2",
      "instock": 80
    },
    {
      "_id": 3,
      "sku": "cashews",
      "description": "product 3",
      "instock": 60
    },
    {
      "_id": 4,
      "sku": "pecans",
      "description": "product 4",
      "instock": 70
    },
    {
      "_id": 5,
      "sku": null,
      "description": "Incomplete"
    }
  ]
}`)

  await expectEditorContent('query', `db.orders.aggregate([
  {
    "$lookup": {
      "from": "inventory",
      "localField": "item",
      "foreignField": "sku",
      "as": "inventory_docs"
    }
  }
])`)
  await expect(page.getByRole('button', { name: '$lookup' })).toBeVisible();

  await expectEditorContent('result', '')
  await page.getByRole('button', { name: 'run' }).click()
  await expectEditorContent('result', `[
  {
    "_id": 1,
    "inventory_docs": [
      {
        "_id": 1,
        "description": "product 1",
        "instock": 120,
        "sku": "almonds"
      }
    ],
    "item": "almonds",
    "price": 12,
    "quantity": 2
  },
  {
    "_id": 2,
    "inventory_docs": [
      {
        "_id": 4,
        "description": "product 4",
        "instock": 70,
        "sku": "pecans"
      }
    ],
    "item": "pecans",
    "price": 20,
    "quantity": 1
  },
  {
    "_id": 3,
    "inventory_docs": [
      {
        "_id": 5,
        "description": "Incomplete",
        "sku": null
      }
    ]
  }
]`)
})

test('test mgodatagen template', async ({ page }) => {

  await page.getByRole('button', { name: 'single collection' }).click()
  await page.locator('#custom-template ul').getByText('mgodatagen').click()

  await expect(page.locator('#custom-mode').getByRole('button', { name: 'mgodatagen' })).toBeVisible();
  await expectEditorContent('config', `[
  {
    "collection": "collection",
    "count": 3,
    "content": {
      "enum": {
        "type": "enum",
        "values": [
          "abc",
          "xyz",
          "123"
        ]
      },
      "int": {
        "type": "int",
        "min": 0,
        "max": 10
      },
      "coordinates": {
        "type": "enum",
        "values": [
          {
            "coordinates": [
              53.23,
              67.12
            ],
            "type": "Point"
          },
          {
            "coordinates": [
              54.23,
              67.12
            ],
            "type": "Point"
          },
          {
            "coordinates": [
              51.23,
              64.12
            ],
            "type": "Point"
          }
        ]
      }
    },
    "indexes": [
      {
        "name": "single_idx",
        "key": {
          "enum": 1
        }
      },
      {
        "name": "compund_idx",
        "key": {
          "enum": -1,
          "int": -1
        }
      },
      {
        "name": "2dsphere_idx",
        "key": {
          "coordinates": "2dsphere"
        }
      }
    ]
  }
]`)

  await expectEditorContent('query', `db.collection.find()`)
  await expect(page.locator('#custom-aggregation_stage')).toBeHidden();

  await expectEditorContent('result', '')
  await page.getByRole('button', { name: 'run' }).click()
  await expectEditorContent('result', `[
  {
    "_id": ObjectId("5a934e000102030405000000"),
    "key": 10
  },
  {
    "_id": ObjectId("5a934e000102030405000001"),
    "key": 2
  },
  {
    "_id": ObjectId("5a934e000102030405000002"),
    "key": 7
  },
  {
    "_id": ObjectId("5a934e000102030405000003"),
    "key": 6
  },
  {
    "_id": ObjectId("5a934e000102030405000004"),
    "key": 9
  },
  {
    "_id": ObjectId("5a934e000102030405000005"),
    "key": 10
  },
  {
    "_id": ObjectId("5a934e000102030405000006"),
    "key": 9
  },
  {
    "_id": ObjectId("5a934e000102030405000007"),
    "key": 10
  },
  {
    "_id": ObjectId("5a934e000102030405000008"),
    "key": 2
  },
  {
    "_id": ObjectId("5a934e000102030405000009"),
    "key": 1
  }
]`)
})

test('test update template', async ({ page }) => {

  await page.getByRole('button', { name: 'single collection' }).click()
  await page.locator('#custom-template ul').getByText('update').click()

  await expect(page.locator('#custom-mode').getByRole('button', { name: 'bson' })).toBeVisible();
  await expectEditorContent('config', `[
  {
    "key": 1
  },
  {
    "key": 2
  }
]`)

  await expectEditorContent('query', `db.collection.update({
  "key": 2
},
{
  "$set": {
    "updated": true
  }
},
{
  "multi": false,
  "upsert": false
})`)
  await expect(page.locator('#custom-aggregation_stage')).toBeHidden();

  await expectEditorContent('result', '')
  await page.getByRole('button', { name: 'run' }).click()
  await expectEditorContent('result', `[
  {
    "_id": ObjectId("5a934e000102030405000000"),
    "key": 1
  },
  {
    "_id": ObjectId("5a934e000102030405000001"),
    "key": 2,
    "updated": true
  }
]`)
})

test('test index template', async ({ page }) => {

  await page.getByRole('button', { name: 'single collection' }).click()
  await page.locator('#custom-template ul').getByText('index').click()

  await expect(page.locator('#custom-mode').getByRole('button', { name: 'mgodatagen' })).toBeVisible();
  await expectEditorContent('config', `[
  {
    "collection": "collection",
    "count": 5,
    "content": {
      "description": {
        "type": "enum",
        "values": [
          "Coffee and cakes",
          "Gourmet hamburgers",
          "Just coffee",
          "Discount clothing",
          "Indonesian goods"
        ]
      }
    },
    "indexes": [
      {
        "name": "description_text_idx",
        "key": {
          "description": "text"
        }
      }
    ]
  }
]`)

  await expectEditorContent('query', `db.collection.find({
  "$text": {
    "$search": "coffee"
  }
})`)
  await expect(page.locator('#custom-aggregation_stage')).toBeHidden();

  await expectEditorContent('result', '')
  await page.getByRole('button', { name: 'run' }).click()
  await expectEditorContent('result', `[
  {
    "_id": ObjectId("5a934e000102030405000002"),
    "description": "Just coffee"
  },
  {
    "_id": ObjectId("5a934e000102030405000000"),
    "description": "Coffee and cakes"
  }
]`)
})

test('test explain template', async ({ page }) => {

  await page.getByRole('button', { name: 'single collection' }).click()
  await page.locator('#custom-template ul').getByText('explain').click()

  await expect(page.locator('#custom-mode').getByRole('button', { name: 'bson' })).toBeVisible();
  await expectEditorContent('config', `[
  {
    "_id": 1,
    "item": "ABC",
    "price": 80,
    "sizes": [
      "S",
      "M",
      "L"
    ]
  },
  {
    "_id": 2,
    "item": "EFG",
    "price": 120,
    "sizes": []
  },
  {
    "_id": 3,
    "item": "IJK",
    "price": 160,
    "sizes": "M"
  },
  {
    "_id": 4,
    "item": "LMN",
    "price": 10
  },
  {
    "_id": 5,
    "item": "XYZ",
    "price": 5.75,
    "sizes": null
  }
]`)

  await expectEditorContent('query', `db.collection.aggregate([
  {
    "$unwind": {
      "path": "$sizes",
      "preserveNullAndEmptyArrays": true
    }
  },
  {
    "$group": {
      "_id": "$sizes",
      "averagePrice": {
        "$avg": "$price"
      }
    }
  },
  {
    "$sort": {
      "averagePrice": -1
    }
  }
]).explain("executionStats")`)
  await expect(page.getByRole('button', { name: "$sort" })).toBeVisible();

  await expectEditorContent('result', '')
  await page.getByRole('button', { name: 'run' }).click()
})