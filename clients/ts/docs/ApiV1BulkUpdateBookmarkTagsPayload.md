
# ApiV1BulkUpdateBookmarkTagsPayload


## Properties

Name | Type
------------ | -------------
`bookmarkIDs` | Array&lt;number&gt;
`tagIDs` | Array&lt;number&gt;

## Example

```typescript
import type { ApiV1BulkUpdateBookmarkTagsPayload } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "bookmarkIDs": null,
  "tagIDs": null,
} satisfies ApiV1BulkUpdateBookmarkTagsPayload

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ApiV1BulkUpdateBookmarkTagsPayload
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


