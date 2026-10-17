
# ApiV1LoginResponseMessage


## Properties

Name | Type
------------ | -------------
`expiration` | number
`token` | string

## Example

```typescript
import type { ApiV1LoginResponseMessage } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "expiration": null,
  "token": null,
} satisfies ApiV1LoginResponseMessage

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ApiV1LoginResponseMessage
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


