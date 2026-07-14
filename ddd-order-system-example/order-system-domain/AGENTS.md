# domain 模块

仅放业务规则、聚合、实体、值对象、领域事件和 Repository 接口。不得引入 Spring、持久化框架或其他项目模块。

聚合必须以行为方法维护不变量，例如 `order.confirm()`，不得为应用层暴露状态 setter。

规模增长后，`model/` 放聚合、实体和值对象，`repository/` 放仓储接口，`event/` 放领域事件。
目录只按领域概念再分组，不能出现 `controller`、`mapper`、`service` 或基础设施实现。
