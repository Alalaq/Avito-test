# User Segmentation API

This is a simple User Segmentation API that allows you to manage user segments and their history.

## Getting Started

1. Clone this repository.
2. Install Go (Golang) on your machine if you haven't already.
3. Install required dependencies 
4. Set up your MySQL database and update the database connection details in `main.go`.
5. Run the API 
6. The API will start and listen on `http://localhost:8080`.
---
## Endpoints

### Create User

- **URL:** `/users/create`
- **Method:** POST

### Update User Segments

- **URL:** `/users/update-segments`
- **Method:** POST
- **Request Body:**
```json
{
 "user_id": 1,
 "segments_to_add": ["NEW_SEGMENT"],
 "segments_to_remove": ["OLD_SEGMENT"],
 "expires_at": "2023-09-01T00:00:00Z"
}
```
- **Response:**
```json
{
  "message": [
    "\"NEW_SEGMENT\" added successfully",
    "\"OLD_SEGMENT\" removed successfully"
  ]
}
```
### Create Segment
- **URL:** `/segments/create`
- **Method:** POST
- **Request Body:**
```json
{
  "slug": "NEW_SEGMENT",
  "auto_add": true,
  "auto_pct": 10
}
```
- **Response:**
```json
{
  "message": "Segment created"
}
```
### Delete Segment
- **URL:** `/segments/delete`
- **Method:** DELETE
- **Request Body:**
```json
{
  "slug": "OLD_SEGMENT"
}
```
- **Response:**
```json
{
  "message": "Segment removed"
}
```
### Get User Segments
- **URL:** `/segments/user-segments`
- **Method:** GET
- **Query Parameters:** 
  - `year` (integer) - Year
  - `month` (integer) - Month
- **Response:** CSV report containing segment history
---
### Error Handling
In case of errors, appropriate error messages will be returned along with the corresponding HTTP status codes.
``` json
Bad Request:
{
  "message": "Invalid year format"
}

Internal Server Error:
{
  "message": "Internal Server Error"
}
```