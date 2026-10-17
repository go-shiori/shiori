
# ModelAccountDTO


## Properties

Name | Type
------------ | -------------
`config` | [ModelUserConfig](ModelUserConfig.md)
`id` | number
`owner` | boolean
`password` | string
`username` | string

## Example

```typescript
import type { ModelAccountDTO } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "config": null,
  "id": null,
  "owner": null,
  "password": null,
  "username": null,
} satisfies ModelAccountDTO

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ModelAccountDTO
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


