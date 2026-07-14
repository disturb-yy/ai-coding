# DDD 订单系统示例

一个 Java 21 + Maven + Spring Boot 的最小可运行示例，重点演示五模块分层 DDD，而不是完整电商功能。

## 模块与依赖

```text
order-system-domain  订单聚合、值对象、领域仓储接口
order-system-api     对外 Command、View、应用服务接口
order-system-app     用例编排，调用聚合行为和领域仓储接口
order-system-base    基础设施适配器；本例使用内存仓储
order-system-main    Spring Boot 启动、Controller、配置与依赖组装

main -> api, app, base
app  -> api, domain
base -> domain
api  -> (无业务模块依赖)
domain -> (无项目模块依赖)
```

`OrderCommandServiceImpl` 从不直接修改订单状态：确认用例调用 `order.confirm()`，由聚合根保证只有待确认订单能被确认。

根目录和每个模块的 `AGENTS.md` 同时说明模块边界；它们由 ArchUnit 规则和 Maven 构建共同兜底，而非仅靠文本约定。
需要新增或维护项目规约时，先阅读 [项目规约编写指南](PROJECT_GOVERNANCE_GUIDE.md)。

## 目录层级与职责

```text
ddd-order-system-example/       聚合构建、跨模块规则、项目总览
├── order-system-domain/        领域模型与领域端口
├── order-system-api/           对外稳定契约
├── order-system-app/           用例编排
├── order-system-base/          基础设施适配器
└── order-system-main/          启动、Web 适配与依赖组装
```

每个模块统一遵循以下三级结构：

| 目录 | 职责 |
| --- | --- |
| 模块根的 `pom.xml` / `AGENTS.md` | 声明依赖边界和模块规则；不放业务实现。 |
| `src/main/java` | 生产代码，包按模块内部职责进一步分组。 |
| `src/test/java` | 与生产包镜像对应，存放单元测试和架构测试。 |
| `src/main/resources` | 运行时资源；应用配置归 `main`，基础设施私有资源归 `base`。 |

模块内部的推荐包名如下；当前是最小示例，类可直接位于模块根包，规模增加后再按这些职责拆分：

| 模块 | 推荐子包 | 说明 |
| --- | --- | --- |
| `domain` | `model`、`repository`、`event` | 聚合和值对象、仓储接口、领域事件。 |
| `api` | `command`、`query`、`view`、`service` | 输入、读取契约、输出和服务接口。 |
| `app` | `service`、`assembler` | 用例实现与 API/领域对象转换。 |
| `base` | `repository`、`client`、`messaging` | 存储、外部调用与消息适配。 |
| `main` | `bootstrap`、`config`、`web` | 启动、Bean 组装与 HTTP 适配。 |

目录分组不能成为跨层依赖的通道：例如 `base.repository` 仍不得被 `app` 依赖，
`main.web` 仍只能依赖 `api` 契约。

## 启动与验证

需要 JDK 21 和 Maven 3.9+：

```bash
mvn clean verify
mvn -pl order-system-main spring-boot:run
```

创建并确认订单：

```bash
curl -X POST http://localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -d '{"customerId":"customer-1","totalAmount":12.50}'

curl -X POST http://localhost:8080/orders/{orderId}/confirm
```

`mvn clean verify` 会执行：

- 聚合和应用用例的单元测试；
- Maven Enforcer 的 Java 21 与依赖收敛检查；
- ArchUnit 的模块依赖规则，防止 domain/app/api 越层依赖。

## 扩展方向

- 在 `base` 中新增 JPA/MyBatis 仓储实现，并保持 `app -> base` 不成立；
- 在 `domain` 中补充领域事件，在 `app` 中编排事件发布；
- 在 `api` 中演进查询契约，保持 Controller 只依赖 API。
