# api 模块

仅放对外稳定契约：Command、Query、View、DTO 和服务接口。不得依赖 domain、app、base 或 main。

规模增长后分别使用 `command/`、`query/`、`view/` 和 `service/`；一个 API 类型只能描述输入、输出或能力，
不得夹带领域实体、Spring 注解或实现逻辑。
