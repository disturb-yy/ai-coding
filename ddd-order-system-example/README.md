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
