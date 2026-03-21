
# ApiV1UpdateAccountPayload


## Properties

Name | Type
------------ | -------------
`config` | [ModelUserConfig](ModelUserConfig.md)
`newPassword` | string
`oldPassword` | string
`owner` | boolean
`username` | string

## Example

```typescript
import type { ApiV1UpdateAccountPayload } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "config": null,
  "newPassword": null,
  "oldPassword": null,
  "owner": null,
  "username": null,
} satisfies ApiV1UpdateAccountPayload

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ApiV1UpdateAccountPayload
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


