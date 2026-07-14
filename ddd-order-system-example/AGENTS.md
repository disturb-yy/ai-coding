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

## 目录层级

| 层级 | 负责内容 | 不应包含 |
| --- | --- | --- |
| 仓库根 | 聚合 `pom.xml`、跨模块规则、总览文档 | 业务 Java 类、数据库实现、Controller |
| 模块根 | 模块 `pom.xml`、模块 `AGENTS.md`、`src` | 其他模块的实现类 |
| `src/main/java` | 当前模块的生产代码 | 测试夹具、环境配置 |
| `src/test/java` | 与生产包同路径的单元/架构测试 | 生产实现 |
| `src/main/resources` | 运行时配置或模块私有资源 | 领域规则；本例仅 `main` 使用应用配置 |

模块内包按职责继续收敛：`domain` 使用 `model`、`repository`、`event`；`api` 使用
`command`、`query`、`view`、`service`；`app` 使用 `service`、`assembler`；`base` 使用
`repository`、`client`、`messaging`；`main` 使用 `web`、`config`、`bootstrap`。最小示例中的类
可暂时直接位于模块根包；当同类职责超过一个清晰概念时，再移动到对应子包，不能借拆包改变模块依赖。

`mvn clean verify` 是改动后的最低验证门槛；ArchUnit 测试是模块依赖规则的强制兜底。
