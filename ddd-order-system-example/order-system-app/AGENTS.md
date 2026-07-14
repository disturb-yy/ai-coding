# app 模块

只编排用例：加载聚合、调用领域行为、保存聚合和返回 API 契约。依赖 api 与 domain，不得依赖 base 或 main。

`service/` 放用例实现，`assembler/` 放 API 契约与领域对象之间的转换。不得放 Controller、Mapper、
SQL 或 Adapter；应用服务不得通过 setter 直接改变聚合状态。
