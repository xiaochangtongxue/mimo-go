# mimo-test


> web项目结构，便于理解

my-go-web-app/\
├── cmd/                 # 编译入口\
│   └── server/          # Web 服务的启动入口\
│       └── main.go      # 项目入口文件\
├── internal/            # 私有应用代码（不可被外部项目引用）\
│   ├── handler/         # 控制层：处理 HTTP 请求和响应\
│   ├── service/         # 业务逻辑层：处理具体的业务流程\
│   ├── repository/      # 数据持久层：操作数据库（ORM/SQL）\
│   ├── model/           # 模型层：数据库实体定义、结构体定义\
│   └── config/          # 配置加载逻辑\
├── pkg/                 # 公共代码：可被外部项目引用的工具类（可选）\
├── api/                 # API 定义文件（如 Swagger, Proto 文件）\
├── configs/             # 静态配置文件（如 config.yaml, .env）\
├── scripts/             # 脚本（部署、编译、数据库迁移脚本）\
├── web/                 # 静态资源、前端代码或模板（如有）\
├── go.mod               # 模块依赖管理文件（核心）\
├── go.sum               # 依赖版本校验文件\
└── README.md            # 项目说明文件\
小米mimo  api测试
小米mimo  api测试
