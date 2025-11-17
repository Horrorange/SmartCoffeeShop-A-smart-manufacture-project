## 先决条件
- 安装并启动 `Docker Desktop`（Windows）
- 在项目根目录：`c:\Users\Orange\Documents\SmartCoffeeShop\SmartCoffeeShop-A-smart-manufacture-project`
- 使用 PowerShell 运行所有命令

## 构建并启动所有容器
- 在项目根目录执行：
  - `docker compose -f script\docker_sim\docker_compose.yml up -d --build`
- 若代码有较大更新，建议强制重建：
  - `docker compose -f script\docker_sim\docker_compose.yml build --no-cache`
  - `docker compose -f script\docker_sim\docker_compose.yml up -d`
- 容器名称与端口映射：
  - `grinder1` → `5021:502`，`grinder2` → `5022:502`
  - `coffee_machine` → `8888:8888`
  - `ice_maker` → `102:102`
  - `mqtt-broker` → `1883:1883`
  - `delivery_robots`（MQTT客户端，无端口暴露；环境变量：`MQTT_HOST=mqtt-broker`, `MQTT_PORT=1883`）

## 运行状态与日志查看
- 查看容器列表：`docker ps`
- 查看单个服务日志：
  - `docker logs -f coffee_machine`
  - `docker logs -f grinder1`
  - `docker logs -f ice_maker`
  - `docker logs -f mqtt-broker`
  - `docker logs -f delivery_robots`
- 若端口占用或异常：先执行 `docker compose -f script\docker_sim\docker_compose.yml down` 再 `up -d --build`

## 端口与连通性自检（宿主机）
- 咖啡机：`Test-NetConnection -ComputerName localhost -Port 8888`
- 磨豆机（选择实例）：`Test-NetConnection -ComputerName localhost -Port 5021`
- 制冰机：`Test-NetConnection -ComputerName localhost -Port 102`
- MQTT Broker：`Test-NetConnection -ComputerName localhost -Port 1883`

## 配置并运行 Go 统一指令网关（宿主机）
- 设置环境变量（根据选择的实例端口）：
  - `setx COFFEE_HOST localhost`；`setx COFFEE_PORT 8888`
  - `setx GRINDER_HOST localhost`；`setx GRINDER_PORT 5021`
  - `setx ICE_HOST localhost`；`setx ICE_RACK 0`；`setx ICE_SLOT 2`
  - `setx MQTT_HOST localhost`；`setx MQTT_PORT 1883`
- 启动网关（IDE/PowerShell均可在 `smart_gateway` 目录）：
  - `go run cmd\main.go`
- 运行示例程序（另一个终端）：
  - `go run cmd\demo\main.go`

## 验证交互
- 咖啡机：网关会在缺料时自动执行 `REFILL:ALL` 并重试 `MAKE`
- 磨豆机：自动补豆（当 `STATUS==2` 或 `BEAN_LEVEL<10`）后再磨粉
- 制冰机：按目标库存生产或按量出冰，轮询状态至完成
- 送餐机器人：向 `test/delivery_robot/command` 发布订单（简化发布模型）

## 常见问题处理
- 端口冲突：执行 `docker compose -f script\docker_sim\docker_compose.yml down` 清理后再 `up -d --build`
- 网络名：Compose 使用 `coffee-net`，保持默认即可
- 重启单个服务：`docker compose -f script\docker_sim\docker_compose.yml restart coffee_machine`
- 完全清理：`docker compose -f script\docker_sim\docker_compose.yml down -v`

## 可选：送餐机器人闭环确认（如需）
- 若需要“发布→收到确认”闭环，可在网关增加订阅 `test/delivery_robot/status` 并在 `deliver` 动作中短超时等待确认后返回。确认后我可以按此扩展实现。