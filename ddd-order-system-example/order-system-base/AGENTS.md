# base 模块

实现领域仓储和外部 Adapter。实现细节不能泄漏到 domain 或 api；本例内存仓储可替换为 JPA/MyBatis 实现。

`repository/` 放仓储实现，`client/` 放外部 HTTP/RPC 适配器，`messaging/` 放消息 Adapter。
持久化实体、Mapper 和客户端 SDK 只留在本模块，不能被 `app` 直接引用。
