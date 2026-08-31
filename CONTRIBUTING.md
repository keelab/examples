# 贡献指南

感谢你参与 Keelith examples。示例项目的目标是用最小、可运行的代码说明一个运行时能力；
新增示例时，请保持单一主题、独立启动，并同步更新根目录 README。

## 开始之前

- 使用 `go.mod` 指定版本的 Go。
- 从仓库根目录运行目标示例，例如 `go run ./09-discovery-client`。
- 不要提交 IDE 配置、编译产物、凭据或本地环境文件。

## 提交前检查

```bash
gofmt -w <changed-go-files>
go test ./<changed-package>
git diff --check
```

如果仓库中存在与本次改动无关的依赖或生成代码问题，请在 Pull Request 中明确说明。

## 提交规范

提交信息采用 Conventional Commits，格式为：

```text
<type>: <简短说明>
```

常用类型包括 `feat`、`fix`、`docs`、`refactor`、`build`、`ci` 和 `chore`。一个提交应聚焦
一个逻辑变更，避免把无关格式化混入功能提交。

## Pull Request

Pull Request 应说明变更目的、涉及的示例目录、运行方式和验证结果。新增或修改示例时，
请同步补充 README，并保持示例中的端口、路径和输出描述与代码一致。
