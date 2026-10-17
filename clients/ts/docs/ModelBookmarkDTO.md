
# ModelBookmarkDTO


## Properties

Name | Type
------------ | -------------
`author` | string
`createArchive` | boolean
`createEbook` | boolean
`createdAt` | string
`excerpt` | string
`hasArchive` | boolean
`hasContent` | boolean
`hasEbook` | boolean
`hasThumbnail` | boolean
`html` | string
`id` | number
`imageURL` | string
`modifiedAt` | string
`_public` | number
`tags` | [Array&lt;ModelTagDTO&gt;](ModelTagDTO.md)
`title` | string
`url` | string

## Example

```typescript
import type { ModelBookmarkDTO } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "author": null,
  "createArchive": null,
  "createEbook": null,
  "createdAt": null,
  "excerpt": null,
  "hasArchive": null,
  "hasContent": null,
  "hasEbook": null,
  "hasThumbnail": null,
  "html": null,
  "id": null,
  "imageURL": null,
  "modifiedAt": null,
  "_public": null,
  "tags": null,
  "title": null,
  "url": null,
} satisfies ModelBookmarkDTO

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ModelBookmarkDTO
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


