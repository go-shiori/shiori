
# ApiV1UpdateCachePayload


## Properties

Name | Type
------------ | -------------
`createArchive` | boolean
`createEbook` | boolean
`ids` | Array&lt;number&gt;
`keepMetadata` | boolean
`skipExist` | boolean

## Example

```typescript
import type { ApiV1UpdateCachePayload } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "createArchive": null,
  "createEbook": null,
  "ids": null,
  "keepMetadata": null,
  "skipExist": null,
} satisfies ApiV1UpdateCachePayload

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ApiV1UpdateCachePayload
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


