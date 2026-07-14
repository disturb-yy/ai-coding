# domain 模块

仅放业务规则、聚合、实体、值对象、领域事件和 Repository 接口。不得引入 Spring、持久化框架或其他项目模块。

聚合必须以行为方法维护不变量，例如 `order.confirm()`，不得为应用层暴露状态 setter。
