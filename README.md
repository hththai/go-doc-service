https://dev.to/lucasdeataides/why-clean-architecture-struggles-in-golang-and-what-works-better-m4g

https://medium.com/@harshgharat663/understanding-handlers-services-repositories-middlewares-request-context-in-backend-2977539931d9

### Example clean structure

handler(controller/api) --> services -->(interface) -> repo (database)

Client → Router → Middleware (CORS, Auth, Logging)
       → Handler (Extract, Validate)
       → Service (Business Logic)
       → Repository (Database)
       → Response to Client


 Layer          | Responsibility            | Knows About                   |
| -------------- | ------------------------- | ----------------------------- |
| **Handler**    | Input/output format, HTTP | Request/Response              |
| **Service**    | Core business logic       | Internal operations           |
| **Repository** | Database interaction      | SQL, queries                  |
| **Middleware** | Cross-cutting concerns    | Logging, auth, error handling |
| **Context**    | Scoped request state      | Shared between layers         |
