# 电脑室教师机答题系统（label-3130）

## 🛠 技术栈
- Frontend: React + Vite + Tailwind CSS + DaisyUI + Zod
- Backend: Golang + Gin + GORM + JWT
- Database: MySQL 8.0

## 🚀 How to Run
1. 确保 Docker Desktop 已启动。
2. 进入项目目录：
   ```bash
   cd label-3130
   ```
3. 一键启动：
   ```bash
   docker compose up --build
   ```
4. 首次启动会自动完成数据库建表和演示数据 Seed。

## 🔗 Services
- Frontend: http://localhost:3130
- Backend API: http://localhost:8130
- Backend Health: http://localhost:8130/health
- Database: localhost:3330 (user: root / pass: root / db: quizlab)

## 🧪 Verification
1. 打开 `http://localhost:3130`，使用教师账号 `admin / 123456` 登录，进入教师机看板。
2. 在“题库管理”中新增、编辑、删除题目，或上传 JSON 题库文件，确认题库列表实时变化。
3. 注册一个学生账号（选择班级），进入学生答题中心，点击“开始新一轮答题”，确认选项顺序被随机打散。
4. 提交答题后，确认学生侧出现“历史成绩”和“错题本（易错题）”。
5. 返回教师账号，确认“最近成绩同步”和“班级错题热区（自动统计）”已更新。

## 🧪 测试账号
- Teacher: `admin / 123456`
- Student(Seed): `stu001 / 123456`

## 📦 题库上传格式（JSON）
可上传两种格式：
- 顶层对象：`{"questions": [...]}`
- 顶层数组：`[...]`

示例：
```json
{
  "questions": [
    {
      "title": "HTTP 状态码 200 表示？",
      "description": "Web 基础",
      "options": [
        { "content": "请求成功", "isCorrect": true },
        { "content": "资源未找到", "isCorrect": false },
        { "content": "服务器错误", "isCorrect": false },
        { "content": "未授权", "isCorrect": false }
      ]
    }
  ]
}
```

## 🧱 项目结构
```text
label-3130/
├── backend/
│   ├── cmd/server
│   └── internal/{auth,config,database,dto,handler,logger,middleware,models,seed,service}
├── frontend/
│   ├── src/{api,components,pages,utils}
│   ├── Dockerfile
│   └── nginx.conf
├── docker-compose.yml
├── .dockerignore
├── .gitignore
└── README.md
```
