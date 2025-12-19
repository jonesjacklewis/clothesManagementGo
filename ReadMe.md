# Clothes Management


This will serve as a simple clothes management API built with Go.

## Features

- User authentication
  - AWS Cognito?
    - This would allow for secure user management and authentication without having to build it from scratch.
    - Integration with AWS services for scalability and reliability
- Should allow management of clothes items
  - CRUD operations for clothes items (Create, Read, Update, Delete)
  - Each item should have attributes like type, size, color, brand, price, etc.
- Some data should be "global" (e.g., types of clothes, brands) and some should be user-specific (e.g., user's clothes inventory).
- Database integration
  - Need to choose between SQL (e.g., PostgreSQL, MySQL) or NoSQL (e.g., MongoDB, DynamoDB) based on the data structure and access patterns.
  - Pros of DynamoDB:
    - Fully managed, serverless database
    - Seamless integration with other AWS services
    - Scalable and high performance
  - Cons of DynamoDB:
    - Limited querying capabilities compared to SQL databases
    - Potentially higher costs for complex queries
  - Likely going with DynamoDB for this project due to its scalability and ease of use with AWS services.
- As this is a hobby project, user specific data will be cleared (manually) and periodically to save on costs.

## Rough User Flow

1. User signs up / logs in using AWS Cognito.
2. User can add clothes items to their inventory.
   1. If the clothing tye or brand does not exist in the global list, it is added.
3. User can view, update, or delete their clothes items.
4. Users can get certain statistics about their clothes (e.g., total number of items, average price, etc.).