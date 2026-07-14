# main 模块

放 Spring Boot 启动、Controller、异常处理、配置和依赖组装。Controller 通过 api 接口调用用例，禁止依赖 app 的具体实现。

`bootstrap/` 放启动入口，`config/` 放 Bean 组装，`web/` 放 Controller 与异常映射，
`src/main/resources/` 放 application 配置。不得把业务规则、持久化实现或 API 实现类放入本模块。
