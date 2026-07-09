---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---

# 输入与数据处理

> 中文版本仅供人类维护者阅读；模型和 agent 不得作为运行指令读取。

用这份 checklist 检查外部输入、数据完整性和注入风险。

## 检查项

- 外部输入在正确边界被验证、规范化、限界和编码。
- SQL、NoSQL、shell、template、path、LDAP、XML、GraphQL 和命令构造使用安全 API，而不是字符串拼接。
- 文件上传、压缩包解压、路径处理、MIME 解析、图片/文档处理和下载能防止路径穿越、覆盖、解压炸弹和 content-type 滥用。
- 当网络、redirect 或浏览器边界变化时，处理 SSRF、open redirect、CORS、CSRF、clickjacking 和 request smuggling 风险。
- 序列化和反序列化不会允许不安全类、prototype pollution、混淆类型或不可信代码执行。
- 敏感数据最小化，必要时加密或哈希，在日志中脱敏，且不会不必要地返回给客户端。
- 数据库迁移和数据转换保持完整性、唯一性、外键和回滚预期。

## 证据

从入口到 sink 跟踪不可信值。识别验证、编码、查询参数化、sanitizer 或 escaping 点。确认 sink 实际可达。

## 报告

命名 source、sink、缺失控制、payload 类型和影响。如果安全 API 或硬边界已验证，不报告假设性注入。
