
# ApiV1BookmarkDataResponse


## Properties

Name | Type
------------ | -------------
`archiveURL` | string
`content` | string
`ebookURL` | string
`hasArchive` | boolean
`hasContent` | boolean
`hasEbook` | boolean
`hasThumbnail` | boolean
`html` | string
`imageURL` | string

## Example

```typescript
import type { ApiV1BookmarkDataResponse } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "archiveURL": null,
  "content": null,
  "ebookURL": null,
  "hasArchive": null,
  "hasContent": null,
  "hasEbook": null,
  "hasThumbnail": null,
  "html": null,
  "imageURL": null,
} satisfies ApiV1BookmarkDataResponse

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ApiV1BookmarkDataResponse
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


