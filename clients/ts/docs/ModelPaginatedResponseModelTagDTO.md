
# ModelPaginatedResponseModelTagDTO


## Properties

Name | Type
------------ | -------------
`items` | [Array&lt;ModelTagDTO&gt;](ModelTagDTO.md)
`total` | number

## Example

```typescript
import type { ModelPaginatedResponseModelTagDTO } from 'shiori-api'

// TODO: Update the object below with actual values
const example = {
  "items": null,
  "total": null,
} satisfies ModelPaginatedResponseModelTagDTO

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as ModelPaginatedResponseModelTagDTO
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


