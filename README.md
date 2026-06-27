# Term Render

一个在终端中实时渲染 3D 模型的工具，使用 tcell 和自定义软件渲染管线。

<video src="mac10.mp4" controls width="100%"></video>

## 用法

```bash
# 编译
make

# 运行
make run ARGS="models/ak47.glb"

# 或直接执行
./term-render models/ak47.glb
```

## 命令

| 命令 | 说明 |
|------|------|
| `make` | 编译项目（默认目标） |
| `make run ARGS="<glb路径>"` | 运行程序 |
| `make clear` | 删除编译产物 |

## 依赖

- Go 1.26+
- 支持 3D 模型文件（GLB 格式）
