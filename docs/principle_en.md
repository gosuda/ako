# **Go Project Structure Guide**

## **1\. Introduction**

This document proposes a standard package structure for Go-based projects and aims to clearly define the roles and responsibilities of each component. This structure is designed based on modern software development principles with the following core goals:

* **Separation of Concerns:** Manages code complexity by ensuring each component focuses on a specific responsibility. For example, database interaction logic is separated from business rule decision logic, making individual parts easier to understand and modify.
* **Clear Dependency Management:** Controls the flow of dependencies between layers and packages clearly and predictably. This prevents circular dependencies and minimizes the impact of changes across the system, enhancing code stability.
* **Testability:** Supports isolating and independently testing each component through interface-based design and dependency injection. This facilitates writing unit and integration tests, improving code quality and reliability.
* **Maintainability & Scalability:** A well-defined structure keeps the codebase understandable and modifiable even as it grows. It also provides a flexible foundation for adding new features or changing existing ones, supporting the project's long-term growth.

This guide enables project participants to collaborate with higher productivity, reduce potential design errors, and build more robust and evolvable applications by organizing and managing code consistently.

## **2\. Top-Level Directory Structure**

It is generally recommended that the project root directory follows this structure:

myproject/  
├── cmd/         \# Application Entry Points & Wiring  
├── internal/    \# Business Logic Composition (Controller, Service, etc.)  
│   ├── controller/ \# Request Handling & Response Mapping  
│   └── service/    \# Core Business Logic & Orchestration  
├── lib/         \# Core Abstractions: Interfaces & Data Structures  
├── pkg/         \# Implementations: Interface Implementations & Infrastructure Logic  
├── proto/       \# Protocol Definition Sources (IDL Sources)  
├── go.mod       \# Go module definition file  
└── go.sum       \# Go module checksum file

## **3\. Detailed Directory Descriptions**

The roles, responsibilities, and rationale for each top-level directory are as follows:

### **3.1. cmd/ \- Execution and Wiring**

* **Role:** Serves as the **execution entry point (main package)** for the application. It holds the core responsibility of gathering components scattered across various layers and **assembling (Wiring)** them into a runnable application.
* **Structure:** Typically, subdirectories are created for each executable file (e.g., web server, microservice API, batch worker, CLI tool), such as cmd/api/, cmd/worker/, cmd/cli/. This allows for clear separation when a project has multiple executables.
* **Key Responsibilities:**
    * **Implementation Initialization:** Instantiates and applies necessary configurations to concrete implementations defined in the pkg/ directory (e.g., zerolog logger, pgx PostgreSQL connection pool, bcrypt password hasher).
    * **Service and Controller Initialization:** Initializes business logic and request handling components defined in internal/service/ and internal/controller/.
    * **Dependency Injection (DI):** One of the most crucial roles. Acts as the application's "glue," connecting interfaces and implementations.
        * Injects the corresponding implementation instance created from pkg/ (e.g., PostgresUserRepository) into the lib/ interfaces (e.g., UserRepository) that service components depend on.
        * Injects the service instance and necessary lib/ interfaces (e.g., InputValidator) that controller components depend on. This process allows each component to depend only on interfaces without needing to know the specific implementation details.
    * **Environment Setup and Initialization:** Performs preparatory tasks needed for application execution, such as loading environment variables, parsing command-line flags, and reading configuration files.
    * **Application Execution:** Once all preparations and wiring are complete, it finally starts the application (e.g., calling http.ListenAndServe, starting a gRPC server, running a worker loop).
* **Principle:** **Must never contain business logic, complex algorithms, or data transformation logic.** cmd/ should focus solely on the clear purposes of **Configuration, Wiring, and Execution**.

### **3.2. internal/ \- Business Logic Composition**

* **Role:** The parent directory for packages that compose and represent the application's core **business logic**. This is where the code performing the actual "work"—handling user requests, applying business rules, managing data—resides. In this structure, it's typically divided into controller and service subdirectories for separation of responsibilities.
* **Key Feature:** Packages within internal rely **only on stable abstractions defined in lib/ (interfaces, data structures) or other layers within the same internal (e.g., controller using service)** to compose the logic flow. Direct dependencies on specific implementation technologies or external systems (pkg/ or cmd/ dependencies) are strictly avoided.
* **Note:** Go's internal directory has special meaning. Packages under this directory can only be imported by code in the direct parent directory and its subdirectories within the same module.

#### **3.2.1. internal/controller/**

* **Role:** Handles incoming requests (e.g., HTTP, gRPC), invokes the appropriate service, and transforms the results into a format the external system understands for the response. Acts as the **entry point for request/response handling and flow control**.
* **Structure:** Usually organized by feature/domain (e.g., user, product) or API resource (e.g., /users, /orders) into sub-packages to group related handlers.
* **Key Responsibilities:**
    * **Request Handling & Data Extraction:** Receives various input forms like HTTP requests or gRPC messages, extracting and parsing necessary data.
    * **Input Validation:** Performs basic data format validation and delegates complex business rule validation.
    * **DTO Transformation:** Converts external request data into data structures the internal service layer can understand.
    * **Service Invocation:** Calls the appropriate service methods to request business logic execution.
    * **Result Transformation & Response:** Converts results received from the service into the expected client format and responds with the appropriate status code.
    * **Error Handling:** Catches errors and generates appropriate error responses.
* **Dependencies:** Depends on internal/service/ and lib/ packages.

#### **3.2.2. internal/service/**

* **Role:** Implements and executes the application's **core business logic and use cases**. Receives requests from the controller and **orchestrates** the actual processing steps, such as data retrieval, business rule application, data modification, and external system integration.
* **Structure:** Organized by feature/domain into sub-packages to group related service logic.
* **Key Responsibilities:**
    * **Business Rule Application & Validation:** Applies complex business rules based on domain knowledge and validates data integrity.
    * **Data Persistence Management:** Manages data in storage (databases, caches) via interfaces defined in lib/repository.
    * **Process Orchestration:** Coordinates various calls (repository, domain services, adapter) in the correct sequence to complete a specific use case.
    * **External System Integration:** Interacts with external systems (API calls, message queue operations) via lib/adapter interfaces.
    * **Transaction Management:** Manages transactions to ensure atomicity of data modification operations when necessary.
    * **Domain Event Publishing:** Publishes significant events resulting from business logic execution.
* **Dependencies:** Depends only on the lib/ package.

### **3.3. lib/ \- Core Abstractions**

* **Role:** The most **fundamental and stable abstraction layer** of the project. Contains **interfaces**, shared **data structures** (VOs, Entities, DTOs), and core **domain model** definitions that are independent of specific technologies or implementations.
* **Key Feature:** Contains only pure **definitions**, with **no concrete implementation code**. It serves as the foundation upon which all other internal layers (internal, pkg, cmd) depend and should have the lowest frequency of change.
* **Data Structure Definition:** DTOs, VOs, Entities, etc., are defined within appropriate sub-packages of lib (e.g., lib/repository/user, lib/domain/order) alongside related interfaces.
* **Dependencies:** **Does not depend on any other internal packages.** External library dependencies should also be minimized.
* **Structure:** Organized by role into subdirectories:
    * adapter/: Defines interfaces and related data structures abstracting communication with external systems.
    * gen/: Stores Go code auto-generated from IDL files.
    * repository/: Defines interfaces and related data structures for accessing the data persistence layer.
    * domain/: Defines core business rules, domain models, domain service interfaces, and related types.
* **Sub-structure:** Each subdirectory is further divided by feature/domain.

### **3.4. pkg/ \- Implementations**

* **Role:** Houses the **concrete implementations** of the interfaces defined in the lib/ directory. Contains actual code using external libraries, technology-specific logic, and infrastructure-related code.
* **Structure:** Organized based on the **specific implementation method, technology, library, or external system** (e.g., pkg/db/postgres, pkg/cache/redis, pkg/logger/zerolog).
* **Key Feature:** Contains the code that actually implements the interfaces from lib/.
* **Dependencies:** Depends on the lib/ package where the interfaces it implements are defined. **Must never depend on the internal/ package.**

### **3.5. proto/ \- Protocol Definitions**

* **Role:** Manages **original Interface Definition Language (IDL) files** like Protocol Buffers (.proto), OpenAPI/Swagger (.yaml, .json), etc., and related specifications.
* **Note:** **Auto-generated Go code** from these files should be placed in the lib/gen/ directory.

## **4\. Core Principles Summary**

To effectively utilize this package structure, understanding and adhering to these core principles is essential:

* **Clear Layer Separation:** Maintain distinct layers: lib/ (Core Abstraction) → internal/service/ (Business Logic) → internal/controller/ (Request Handling) → cmd/ (Execution/Wiring). pkg/ (Implementation) implements lib/ and is separate.
* **Strict Interface/Implementation Separation:** lib/ contains only definitions; pkg/ contains only implementations of lib/. The internal layer uses lib/ interfaces to compose logic.
* **Unidirectional Dependencies:** Dependencies must always flow from outer layers towards inner, more stable layers (cmd/ → internal/, pkg/ → lib/, internal/ → lib/). **pkg/ must not depend on internal/, and service must not depend on controller.**
* **Implementation-Centric pkg/ Structure:** pkg/ is organized based on specific technologies or libraries ('how' it's implemented).
* **Dependency Injection (DI) Utilization:** Actual connections between layers are made loosely via interfaces, with concrete implementations wired together in cmd/.
* **Data Structure Location:** DTOs, VOs, etc., are co-located with the relevant interfaces or logic within lib's sub-packages to enhance cohesion.

## **5\. Comparison with Other Layered Architectures**

### **5.1. Architecture Overviews**

* **Hexagonal Architecture (Ports and Adapters):** Centers around a core business logic (Core), with interfaces (Ports) for interaction and implementations (Adapters) connecting to the outside world. Dependencies always point inward from Adapters to the Core.
* **Onion Architecture:** Places the Domain Model at the center, surrounded by layers like Domain Services, Application Services, and Infrastructure. Dependencies always point inward towards the Domain Model.

### **5.2. Commonalities**

All three architectures (including the proposed Go structure) share the core goals of **Separation of Concerns**, adherence to the **Dependency Inversion Principle**, improved **Testability**, and enhanced **Maintainability & Flexibility**. They focus on protecting the core business logic from external changes.

### **5.3. Differences and Mapping to Go Structure**

| Feature | Proposed Go Structure | Hexagonal Architecture | Onion Architecture |
| :---- | :---- | :---- | :---- |
| **Core Logic** | internal/service/, lib/domain/ | Core (Application/Domain) | Domain Model, Domain Services, App Services |
| **Abstraction/IF** | lib/ (All interfaces centralized) | Ports (Primary/Secondary) | Defined in Inner Layers |
| **Implementation** | pkg/ (Technology-specific implementations) | Adapters (Primary/Secondary) | Outermost Layer (Infrastructure) |
| **Request Handling** | internal/controller/ | Primary Adapters | UI / Infrastructure Layer |
| **DI Wiring** | cmd/ | Composition Root (No explicit dir) | Composition Root (No explicit dir) |
| **Main Emphasis** | Go features, Tech-centric pkg, Centralized lib | Clear boundary (Ports) | Domain Model centricity |

The proposed Go structure follows the core principles of Hexagonal and Onion architectures but adopts a pragmatic approach reflecting Go's characteristics and community conventions, notably the centralized lib for all interfaces and the technology-based organization of pkg.

## **6\. Conclusion: Towards Building Robust and Flexible Go Applications**

The package structure presented in this guide goes beyond mere directory organization; it embodies a design philosophy aimed at making Go applications **Robust, Flexible, and Sustainable**.

**Key Benefits:**

* **Improved Maintainability:** Clear role separation facilitates code understanding, modification, and limits the impact scope of changes.
* **Increased Scalability:** Adding new features or swapping technologies is easier with minimal impact on existing code.
* **High Testability:** Interface-based design and DI enable component isolation and straightforward unit testing.
* **Enhanced Team Collaboration:** A consistent structure aids code convention adherence and mutual understanding.
* **Low Coupling, High Cohesion:** Unidirectional dependencies reduce coupling, while co-locating related code increases cohesion.

**Considerations for Application:**

This structure might not be a perfect fit for every project. Its suitability depends on factors like project size, complexity, and team experience. Adjustments can be made, especially for smaller projects. The crucial aspect is understanding and consistently applying the **core principles (Layer Separation, Unidirectional Dependencies, Interface-Based Design, Dependency Injection)**.

Adhering to these principles allows effective management of complexity as the project grows and changes, ensuring a stable and efficient development process over the long term. Ultimately, this guide provides a solid foundation for leveraging Go's strengths to successfully build and evolve complex, high-quality software.