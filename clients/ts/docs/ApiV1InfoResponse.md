
# ApiV1InfoResponse


## Properties

Name | Type
------------ | -------------
`database` | string
`os` | string
`version` | [ApiV1InfoResponseVersion](ApiV1InfoResponseVersion.md)

## Example

```typescript
import type { ApiV1InfoResponse } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "database": null,
  "os": null,
  "version": null,
} satisfies ApiV1InfoResponse

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ApiV1InfoResponse
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


