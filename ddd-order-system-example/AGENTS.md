# DDD 模块约束

本项目使用 Java 21、Maven 多模块和分层 DDD。改动前先确定业务规则的归属，不以包名或目录名替代边界判断。

## 允许的依赖方向

```text
main -> api, app, base
app  -> api, domain
base -> domain
api  -> 无其他业务模块
domain -> 无其他项目模块
```

禁止 `domain -> app/base/main`、`app -> base`。基础设施若需实现应用层 Port，必须由 main 负责组装。

## 放置规则

- `domain`：聚合、实体、值对象、领域事件、领域服务和 Repository 接口；领域状态只能由聚合行为改变。
- `api`：稳定的 Command、Query、View 和服务接口；不得泄漏实体或依赖实现模块。
- `app`：用例编排、事务边界、聚合加载和保存；不得直接调用 Mapper 或修改聚合内部状态。
- `base`：Repository、数据库、缓存、消息和 HTTP/RPC Adapter 的实现；不是通用工具箱。
- `main`：Spring Boot 启动、Controller、配置和依赖组装；Controller 只依赖 API 契约。

`mvn clean verify` 是改动后的最低验证门槛；ArchUnit 测试是模块依赖规则的强制兜底。
