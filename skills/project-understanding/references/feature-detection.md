# Feature Detection Rules

A Feature is a business capability.

DO NOT use technical names.

Bad:

UserController
OrderService

Good:

User Login
Create Order
Refund Order

---

Route + Flow + Module

=> Feature

Example

POST /login

AuthController
AuthService
UserRepository

=> User Login

---

POST /orders

OrderService
PaymentService
InventoryService

=> Create Order

---

Feature Name Priority

1 Business Meaning
2 Route Name
3 Flow Name
4 Module Name