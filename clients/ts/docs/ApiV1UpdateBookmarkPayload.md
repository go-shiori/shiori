
# ApiV1UpdateBookmarkPayload


## Properties

Name | Type
------------ | -------------
`excerpt` | string
`_public` | number
`title` | string
`url` | string

## Example

```typescript
import type { ApiV1UpdateBookmarkPayload } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "excerpt": null,
  "_public": null,
  "title": null,
  "url": null,
} satisfies ApiV1UpdateBookmarkPayload

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ApiV1UpdateBookmarkPayload
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


