# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [myservice/v1/myservice.proto](#myservice_v1_myservice-proto)
    - [ChatRequest](#myservice-v1-ChatRequest)
    - [ChatResponse](#myservice-v1-ChatResponse)
    - [GetUserRequest](#myservice-v1-GetUserRequest)
    - [GetUserResponse](#myservice-v1-GetUserResponse)
    - [ListUsersRequest](#myservice-v1-ListUsersRequest)
    - [ListUsersResponse](#myservice-v1-ListUsersResponse)
    - [UpdateUsersRequest](#myservice-v1-UpdateUsersRequest)
    - [UpdateUsersResponse](#myservice-v1-UpdateUsersResponse)
    - [User](#myservice-v1-User)
  
    - [MyService](#myservice-v1-MyService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="myservice_v1_myservice-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## myservice/v1/myservice.proto



<a name="myservice-v1-ChatRequest"></a>

### ChatRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| text | [string](#string) |  |  |






<a name="myservice-v1-ChatResponse"></a>

### ChatResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| text | [string](#string) |  |  |






<a name="myservice-v1-GetUserRequest"></a>

### GetUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |






<a name="myservice-v1-GetUserResponse"></a>

### GetUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#myservice-v1-User) |  |  |






<a name="myservice-v1-ListUsersRequest"></a>

### ListUsersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  |  |
| page_token | [string](#string) |  |  |






<a name="myservice-v1-ListUsersResponse"></a>

### ListUsersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [User](#myservice-v1-User) | repeated |  |
| next_page_token | [string](#string) |  |  |






<a name="myservice-v1-UpdateUsersRequest"></a>

### UpdateUsersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [User](#myservice-v1-User) | repeated |  |






<a name="myservice-v1-UpdateUsersResponse"></a>

### UpdateUsersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| updated_count | [int32](#int32) |  |  |






<a name="myservice-v1-User"></a>

### User



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user_id | [string](#string) |  |  |
| name | [string](#string) |  |  |





 

 

 


<a name="myservice-v1-MyService"></a>

### MyService


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetUser | [GetUserRequest](#myservice-v1-GetUserRequest) | [GetUserResponse](#myservice-v1-GetUserResponse) | Unary RPC |
| ListUsers | [ListUsersRequest](#myservice-v1-ListUsersRequest) | [ListUsersResponse](#myservice-v1-ListUsersResponse) stream | Server Streaming RPC |
| UpdateUsers | [UpdateUsersRequest](#myservice-v1-UpdateUsersRequest) stream | [UpdateUsersResponse](#myservice-v1-UpdateUsersResponse) | Client Streaming RPC |
| Chat | [ChatRequest](#myservice-v1-ChatRequest) stream | [ChatResponse](#myservice-v1-ChatResponse) stream | Bidirectional Streaming RPC |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

