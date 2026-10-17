
# ApiV1CreateBookmarkPayload


## Properties

Name | Type
------------ | -------------
`excerpt` | string
`options` | [ApiV1CreateBookmarkOptions](ApiV1CreateBookmarkOptions.md)
`_public` | number
`title` | string
`url` | string

## Example

```typescript
import type { ApiV1CreateBookmarkPayload } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "excerpt": null,
  "options": null,
  "_public": null,
  "title": null,
  "url": null,
} satisfies ApiV1CreateBookmarkPayload

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ApiV1CreateBookmarkPayload
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


