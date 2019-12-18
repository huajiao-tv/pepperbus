# dashboard

## dashboard-ui

dashboard-ui是pepper_bus的后台管理页面

### 功能特点
1. ui界面才用vue框架
2. 代码在vue-admin的基础上开发


### 服务搭建
1. npm install
2. npm run dev　（测试运行）
3. npm run build:prod　(正式环境打包，运行需要nginx配合)

## dashboard

### 测试环境配置
1. 测试环境数据库在花椒测试机

### 服务搭建
1. go build -o ./bin/dashboard  github.com/huajiao-tv/pepperbus/dashboard
2. ./bin/dashboard
