# SolarSage

最全面的开源占星计算引擎。
**40 个 MCP 工具 · 40 个 REST 接口 · 38 个包 · 824+ 测试用例 · 11 种宫位系统 · 5 种恒星时差 · 50+ 恒星 · 27 宿 · 15+ 阿拉伯幸运点 · 7 种相位模式** — 亚弧秒精度。

可用作 **Go 库**、AI 助手的 **MCP 服务器**，或面向 Web/移动端的 **RESTful HTTP API**。

基于 [瑞士星历表](https://www.astro.com/swisseph/) 构建。
经过独立验证，**精度 100%**（247/247 过境事件与 Solar Fire 9 完全吻合）。

> [English Documentation →](README.md)

---

## 目录

- [为什么选择 SolarSage？](#为什么选择-solarsage)
- [功能列表](#功能列表)
- [快速开始](#快速开始)
  - [环境要求与构建](#环境要求与构建)
  - [作为 MCP 服务器运行](#作为-mcp-服务器运行)
  - [作为 REST API 服务器运行](#作为-rest-api-服务器运行)
  - [作为 Go 库使用](#作为-go-库使用)
- [MCP 工具参考](#mcp-工具参考)
- [REST API 参考](#rest-api-参考)
- [架构](#架构)
- [性能](#性能)
- [精度](#精度)
- [Docker](#docker)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

---

## 为什么选择 SolarSage？

| 功能 | SolarSage | flatlib (Python) | Kerykeion (Python) | Swiss Ephemeris (C) |
|---|---|---|---|---|
| 语言 | Go | Python | Python | C |
| 过境检测 | 7 种类型，1 秒精度 | 基础 | 无 | 手动 |
| 太阳/月亮回归 | 支持序列 | 单次 | 单次 | 手动 |
| 合盘 | 中点法 | 无 | 无 | 手动 |
| 合盘评分 | 分类详细 | 无 | 基础 | 无 |
| 食相检测 | 日食 + 月食 | 无 | 无 | 底层 |
| 小限法 | 年度 + 月度 | 无 | 无 | 无 |
| 阿拉伯幸运点 | 15+ 含昼夜区分 | 无 | 无 | 无 |
| 本质尊贵 | 完整 + 互容 | 基础 | 基础 | 无 |
| 相位模式 | 7 种 | 无 | 无 | 无 |
| 恒星 | 50+ 目录 | 无 | 无 | 底层 |
| 中点分析 | 90°宇宙生物学盘 + 激活 | 无 | 无 | 无 |
| 谐波盘 | 1-180 | 无 | 无 | 无 |
| 行星时 | 迦勒底体系 | 无 | 无 | 无 |
| 宫位系统 | 11 种 | 7 种 | 3 种 | 全部 |
| 恒星/吠陀 | 星宿 + 大运 | 无 | 无 | 手动 |
| 主星链 | 完整链条 | 无 | 无 | 无 |
| 主限推运 | 托勒密半弧法 | 无 | 无 | 无 |
| 符号推运 | 4 种方法 | 无 | 无 | 无 |
| 菲尔达利亚 | 昼夜序列 | 无 | 无 | 无 |
| 偕日升落 | 能见度算法 | 无 | 无 | 底层 |
| 八宫分点 | Bindu 表 | 无 | 无 | 无 |
| 瑜伽检测 | 10+ 种瑜伽 | 无 | 无 | 无 |
| 分部盘 | 16 种 Varga | 无 | 无 | 无 |
| 吉凶修正 | 基于相位评分 | 无 | 无 | 无 |
| 一键报告 | 所有技法合一 | 无 | 无 | 无 |
| 图表可视化 | 星盘坐标 | 无 | 无 | 无 |
| MCP 服务器 | 40 个工具 | 无 | 无 | 无 |
| REST API | 40 个接口 | 无 | 无 | 无 |
| 精度验证 | 247/247 (100%) | 否 | 否 | 不适用 |
| 线程安全 | 是（互斥锁） | 否 | 否 | 否 |

---

## 功能列表

### 星盘计算
- **本命盘** — 行星位置、宫位（11 种系统）、轴点、相位（9 种类型）
- **双盘叠加** — 合盘/过境叠加，含跨盘相位
- **合盘** — 中点法关系分析
- **戴维森盘** — 时空中点关系盘
- **谐波盘** — 分部盘（5 次五分相、7 次七分相、9 次九分相等）
- **恒星时盘** — 5 种恒星时差系统

### 预测技法
- **过境检测** — Tr-Na、Tr-Tr、Tr-Sp、Tr-Sa、Sp-Na、Sp-Sp、Sa-Na，精度达 1 秒
- **二次推运** — 一年一天法推运行星位置及事件
- **太阳弧推运** — 太阳弧指向位置及事件
- **主限推运** — 托勒密半弧法配合纳伯德键
- **符号推运** — 每年 1 度法、纳伯德、小限、自定义比率
- **太阳/月亮回归** — 精确回归盘，支持序列计算
- **年度小限** — 时主技法，含月度子限
- **菲尔达利亚** — 行星周期系统（昼夜序列）
- **入相位/入宫** — 行星换星座和换宫检测
- **停滞** — 逆行及顺行停滞检测

### 传统占星
- **本质尊贵** — 主宰、擢升、失势、陷落，含评分
- **互容** — 主宰和擢升互容
- **日夜派** — 昼夜行星对位分析
- **阿拉伯幸运点** — 15+ 幸运点（福运点、灵魂点、爱神点、胜利点等），含昼夜换算
- **旬区与界** — 迦勒底旬区和埃及/托勒密界分
- **行星时** — 迦勒底行星时，含日出/日落计算
- **对镜点** — 至点和分点镜像点及配对检测
- **主星链** — 主星链条、最终主星、互相主星
- **吉凶修正** — 基于相位的行星状态分析
- **偕日升落** — 瑞士星历表能见度算法

### 模式检测
- **相位模式** — 大三角、T 型三角、大十字、神秘三角、风筝、神秘矩形、星群
- **恒星** — 50+ 主要恒星目录，岁差修正合相检测
- **中点分析** — 完整中点树、90 度宇宙生物学盘、激活点

### 吠陀/恒星占星
- **恒星时盘** — 5 种恒星时差（拉伊里、拉曼、克里希纳穆提、法根-布拉德利、尤克特斯瓦尔）
- **星宿** — 全部 27 个月亮星宿，含四分位和维姆肖塔里主星
- **维姆肖塔里大运** — 月亮星宿完整大运序列
- **分部盘** — 16 种 Varga 盘（D1-D60、Navamsa、Dasamsa 等）
- **八宫分点** — Bindu 表和 Sarvashtakavarga 总分
- **瑜伽检测** — Mahapurusha、Raja、Dhana、Gajakesari 等

### 天文学
- **月相** — 新月/满月查找，相角，照明百分比
- **食相查找** — 日食和月食检测，含类型分类
- **月亮空亡** — 自动检测空亡时段及相位上下文

### 关系分析
- **合盘评分** — 兼容性分析，含分类细分（爱情、激情、沟通、承诺）
- **合盘** — 中点法，含相位
- **戴维森盘** — 时空中点关系盘

### 可视化
- **星盘轮坐标** — 行星 x/y 坐标、宫位线、相位线、星座分段，适用于 SVG/Canvas 渲染

### 支持天体

**行星：** 太阳、月亮、水星、金星、火星、木星、土星、天王星、海王星、冥王星、凯龙星、北交点（真实/平均）、南交点、莉莉丝（平均/真实）

**特殊点：** 上升点、中天、下降点、天底、顶点、东方点、福运点、灵魂点

**宫位系统：** 普拉西德斯、科赫、等宫、整星座、坎帕纳斯、雷吉奥蒙塔努斯、波菲里、莫里纳斯、地形、阿尔卡比提乌斯、子午线

**输出格式：** 所有图表类型均支持 JSON 和 CSV。Unicode 占星字符（♈♉♊♋♌♍♎♏♐♑♒♓，☉☽☿♀♂♃♄，☌☍△□✱）。

---

## 快速开始

### 环境要求与构建

**系统要求：**
- Go 1.25+
- GCC（用于 CGO / 瑞士星历表编译）
- 瑞士星历表 C 源码必须位于 `third_party/swisseph/`（详见 [贡献指南](CONTRIBUTING.md)）

```bash
git clone https://github.com/shaobaobaoer/solarsage-mcp.git
cd solarsage-mcp
make build        # → bin/solarsage-mcp  （MCP 服务器）
make build-api    # → bin/solarsage-api  （REST API 服务器）
```

`make build` 步骤通过 CGO 编译瑞士星历表 C 库并将其静态链接到 Go 二进制文件中。无需单独安装 `libswisseph`，所有内容均已内置于 `third_party/`。

### 作为 MCP 服务器运行

```bash
./bin/solarsage-mcp

# 指定自定义星历数据路径
SWISSEPH_EPHE_PATH=/path/to/ephe ./bin/solarsage-mcp
```

#### Claude Desktop 集成

在 `claude_desktop_config.json` 中添加：

```json
{
  "mcpServers": {
    "astrology": {
      "command": "/path/to/solarsage-mcp",
      "env": {
        "SWISSEPH_EPHE_PATH": "/path/to/ephe"
      }
    }
  }
}
```

#### Cursor / 其他 MCP 客户端

```json
{
  "mcpServers": {
    "solarsage": {
      "command": "/path/to/solarsage-mcp"
    }
  }
}
```

### 作为 REST API 服务器运行

```bash
./bin/solarsage-api --port 8080

# 启用 API 密钥认证
./bin/solarsage-api --port 8080 --api-key your-secret-key

# 示例：本命盘请求
curl -X POST http://localhost:8080/api/v1/chart/natal \
  -H "Content-Type: application/json" \
  -d '{"latitude": 51.5074, "longitude": -0.1278, "jd_ut": 2451545.0}'

# 示例：过境事件
curl -X POST http://localhost:8080/api/v1/transit \
  -H "Content-Type: application/json" \
  -d '{
    "natal_lat": 51.5074, "natal_lon": -0.1278, "natal_jd": 2451545.0,
    "start_jd": 2460676.5, "end_jd": 2460736.5
  }'
```

全部 40 个接口均在 `/api/v1/` 下。已启用 CORS。可通过 `X-API-Key` 请求头进行可选 API 密钥认证。

健康检查：`GET /api/v1/health`

---

## 作为 Go 库使用

### 高级 API（推荐）

`solarsage` 包提供了高级 API，具有合理的默认值。传入 ISO 8601 日期时间字符串，而非儒略日数：

```go
package main

import (
    "fmt"
    "github.com/shaobaobaoer/solarsage-mcp/pkg/solarsage"
)

func main() {
    solarsage.Init("/path/to/ephe")
    defer solarsage.Close()

    // 本命盘
    chart, _ := solarsage.NatalChart(51.5074, -0.1278, "1990-06-15T14:30:00Z")
    for _, p := range chart.Planets {
        fmt.Printf("%s 在 %s（第 %d 宫）\n", p.PlanetID, p.Sign, p.House)
    }

    // 2025 年太阳回归
    sr, _ := solarsage.SolarReturn(51.5074, -0.1278, "1990-06-15T14:30:00Z", 2025)
    fmt.Printf("太阳回归：年龄 %.1f\n", sr.Age)

    // 当前月相
    phase, _ := solarsage.MoonPhase("2025-03-18T12:00:00Z")
    fmt.Printf("月亮：%s（照明 %.0f%%）\n", phase.PhaseName, phase.Illumination*100)

    // 指定日期范围内的食相
    eclipses, _ := solarsage.Eclipses("2025-01-01", "2026-01-01")
    for _, e := range eclipses {
        fmt.Printf("食相：%s 在 %s\n", e.Type, e.MoonSign)
    }

    // 关系兼容性
    score, _ := solarsage.Compatibility(
        51.5074, -0.1278, "1990-06-15T14:30:00Z",
        40.7128, -74.006, "1992-03-22T08:00:00Z",
    )
    fmt.Printf("兼容性：%.0f%%\n", score.Compatibility)

    // 单行星位置
    pos, _ := solarsage.PlanetPosition("Venus", "2025-03-18T12:00:00Z")
    fmt.Printf("金星：%s %.2f°\n", pos.Sign, pos.SignDegree)

    // 吠陀恒星时盘（含星宿）
    vedic, _ := solarsage.SiderealChart(51.5074, -0.1278, "1990-06-15T14:30:00Z")
    for _, p := range vedic.Planets {
        fmt.Printf("%s：%s（星宿：%s，第 %d 四分位）\n",
            p.PlanetID, p.SiderealSign, p.Nakshatra, p.NakshatraPada)
    }

    // 维姆肖塔里大运周期
    periods, _ := solarsage.Dasha(51.5074, -0.1278, "1990-06-15T14:30:00Z")
    for _, d := range periods {
        fmt.Printf("年龄 %.0f–%.0f：%s 大运\n", d.StartAge, d.StartAge+d.Years, d.Lord)
    }

    // 星盘轮坐标（用于 SVG/Canvas 渲染）
    wheel, _ := solarsage.ChartWheel(51.5074, -0.1278, "1990-06-15T14:30:00Z")
    for _, p := range wheel.Planets {
        fmt.Printf("%s 坐标 (%.2f, %.2f)\n", p.PlanetID, p.Position.X, p.Position.Y)
    }

    // 综合报告（一次调用包含所有技法）
    report, _ := solarsage.FullReport(51.5074, -0.1278, "1990-06-15T14:30:00Z")
    fmt.Printf("四象：火象=%d 土象=%d 风象=%d 水象=%d\n",
        report.ElementBalance["Fire"], report.ElementBalance["Earth"],
        report.ElementBalance["Air"], report.ElementBalance["Water"])
}
```

### 底层 API

直接导入各个包，对弧度容许度、宫位系统和行星选择进行完全控制：

```go
import (
    "github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
    "github.com/shaobaobaoer/solarsage-mcp/pkg/models"
    "github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
    "github.com/shaobaobaoer/solarsage-mcp/pkg/transit"
)

func main() {
    sweph.Init("/path/to/ephe")
    defer sweph.Close()

    planets := []models.PlanetID{
        models.PlanetSun, models.PlanetMoon, models.PlanetVenus,
    }

    // 完全控制弧度容许度、宫位系统和行星选择
    info, _ := chart.CalcSingleChart(
        51.5074, -0.1278, 2451545.0,
        planets,
        models.OrbConfig{Conjunction: 10, Trine: 8, Square: 8},
        models.HouseKoch,
    )

    // 过境搜索，完整选项
    events, _ := transit.CalcTransitEvents(transit.TransitCalcInput{
        NatalLat:         51.5074,
        NatalLon:         -0.1278,
        NatalJD:          2451545.0,
        NatalPlanets:     planets,
        TransitLat:       51.5074,
        TransitLon:       -0.1278,
        StartJD:          2460676.5,
        EndJD:            2460706.5,
        TransitPlanets:   planets,
        EventConfig:      models.DefaultEventConfig(),
        OrbConfigTransit: models.DefaultOrbConfig(),
        HouseSystem:      models.HousePlacidus,
    })
    _ = info
    _ = events
}
```

---

## MCP 工具参考

### 工具类

| 工具 | 描述 |
|------|------|
| `geocode` | 将地名转换为经纬度和时区 |
| `datetime_to_jd` | 将 ISO 8601 日期时间字符串转换为儒略日（UT 和 TT） |
| `jd_to_datetime` | 将儒略日数转换为 ISO 8601 日期时间字符串 |

### 星盘计算

| 工具 | 描述 |
|------|------|
| `calc_planet_position` | 指定时间的单行星位置、星座、宫位和速度 |
| `calc_single_chart` | 完整本命/事件盘：位置、宫位（11 种系统）和相位 |
| `calc_double_chart` | 合盘/过境双盘，含两盘跨盘相位 |
| `calc_composite_chart` | 关系分析用合盘（中点法） |
| `calc_davison_chart` | 戴维森关系盘（时空中点） |
| `calc_harmonic_chart` | 任意谐波数的谐波（分部）盘 |
| `calc_sidereal_chart` | 含星宿、四分位和维姆肖塔里主星的恒星时盘 |
| `calc_divisional_chart` | 吠陀 Varga 盘（Navamsa D9、Dasamsa D10 等） |
| `calc_chart_wheel` | 用于 SVG/Canvas 渲染的星盘轮 x/y 坐标 |

### 预测技法

| 工具 | 描述 |
|------|------|
| `calc_transit` | 指定日期范围内的完整过境事件搜索（JSON 或 CSV 输出） |
| `calc_progressions` | 二次推运行星位置（一年一天法） |
| `calc_solar_arc` | 太阳弧指向行星位置 |
| `calc_primary_directions` | 主限推运（托勒密半弧法 + 纳伯德键） |
| `calc_symbolic_directions` | 符号推运（每年 1°、纳伯德、小限、自定义比率） |
| `calc_solar_return` | 指定年份的太阳回归盘 |
| `calc_lunar_return` | 下次月亮回归盘 |
| `calc_profection` | 年度/月度小限，含激活时主 |
| `calc_firdaria` | 菲尔达利亚行星周期时间轴（昼夜序列） |

### 传统占星

| 工具 | 描述 |
|------|------|
| `calc_dignity` | 所有行星的本质尊贵、互容和昼夜派 |
| `calc_bonification` | 基于相位和尊贵的吉凶修正评分 |
| `calc_lots` | 阿拉伯幸运点——福运点、灵魂点、爱神点、胜利点等 10+ 种 |
| `calc_bounds` | 迦勒底旬区和埃及/托勒密界（界分）主星 |
| `calc_planetary_hours` | 任意日期/地点的迦勒底行星时（含日出/日落） |
| `calc_antiscia` | 至点/分点对镜点和逆对镜点 |
| `calc_heliacal_events` | 偕日升落（瑞士星历表能见度算法） |

### 模式检测

| 工具 | 描述 |
|------|------|
| `calc_aspect_patterns` | 检测大三角、T型三角、大十字、神秘三角、风筝、神秘矩形、星群 |
| `calc_fixed_stars` | 50+ 恒星目录的合相（岁差修正） |
| `calc_midpoints` | 中点树，含 90 度宇宙生物学盘和激活点 |

### 天文学

| 工具 | 描述 |
|------|------|
| `calc_lunar_phase` | 当前月相、照明百分比和相角 |
| `calc_lunar_phases` | 查找指定日期范围内的新月、满月和四分相 |
| `calc_eclipses` | 日食和月食查找，含类型分类 |

### 分析

| 工具 | 描述 |
|------|------|
| `calc_synastry` | 关系兼容性评分，含分类细分 |
| `calc_dispositors` | 主星链条、最终主星和互相主星 |
| `calc_natal_report` | 综合本命分析（所有技法合一） |

### 吠陀/恒星占星

| 工具 | 描述 |
|------|------|
| `calc_vimshottari_dasha` | 基于月亮星宿的维姆肖塔里大运周期 |
| `calc_ashtakavarga` | 八宫分点 Bindu 表和 Sarvashtakavarga 总分 |
| `calc_yogas` | 吠陀瑜伽检测（Mahapurusha、Raja、Dhana、Gajakesari 等） |

---

## REST API 参考

所有接口均为 `POST /api/v1/<path>`，接受并返回 JSON。所有路由已启用 CORS。可通过 `X-API-Key` 请求头进行可选认证。

| 接口 | 描述 |
|------|------|
| `GET  /api/v1/health` | 健康检查 |
| `POST /api/v1/geocode` | 地名地理编码 |
| `POST /api/v1/datetime/to-jd` | ISO 8601 → 儒略日 |
| `POST /api/v1/datetime/from-jd` | 儒略日 → ISO 8601 |
| `POST /api/v1/planet/position` | 单行星位置 |
| `POST /api/v1/chart/natal` | 本命盘 |
| `POST /api/v1/chart/double` | 双盘/合盘 |
| `POST /api/v1/chart/composite` | 合盘 |
| `POST /api/v1/chart/davison` | 戴维森盘 |
| `POST /api/v1/chart/harmonic` | 谐波盘 |
| `POST /api/v1/chart/sidereal` | 恒星时/吠陀盘 |
| `POST /api/v1/chart/divisional` | Varga 分部盘 |
| `POST /api/v1/chart/wheel` | 星盘轮坐标 |
| `POST /api/v1/transit` | 过境事件 |
| `POST /api/v1/progressions` | 二次推运 |
| `POST /api/v1/solar-arc` | 太阳弧推运 |
| `POST /api/v1/primary-directions` | 主限推运 |
| `POST /api/v1/symbolic-directions` | 符号推运 |
| `POST /api/v1/solar-return` | 太阳回归盘 |
| `POST /api/v1/lunar-return` | 月亮回归盘 |
| `POST /api/v1/dignity` | 本质尊贵 |
| `POST /api/v1/bonification` | 吉凶修正 |
| `POST /api/v1/dispositors` | 主星链条 |
| `POST /api/v1/profection` | 年度小限 |
| `POST /api/v1/firdaria` | 菲尔达利亚周期 |
| `POST /api/v1/lots` | 阿拉伯幸运点 |
| `POST /api/v1/bounds` | 旬区与界 |
| `POST /api/v1/antiscia` | 对镜点 |
| `POST /api/v1/planetary-hours` | 行星时 |
| `POST /api/v1/heliacal` | 偕日升落 |
| `POST /api/v1/aspects/patterns` | 相位模式 |
| `POST /api/v1/fixed-stars` | 恒星合相 |
| `POST /api/v1/midpoints` | 中点分析 |
| `POST /api/v1/synastry` | 合盘评分 |
| `POST /api/v1/vedic/dasha` | 维姆肖塔里大运 |
| `POST /api/v1/vedic/ashtakavarga` | 八宫分点 |
| `POST /api/v1/vedic/yogas` | 瑜伽检测 |
| `POST /api/v1/lunar/phase` | 月相 |
| `POST /api/v1/lunar/phases` | 月相列表 |
| `POST /api/v1/lunar/eclipses` | 食相查找 |
| `POST /api/v1/report/natal` | 综合本命报告 |

---

## Go API 文档

所有导出类型和函数的完整 API 文档位于 [`doc/`](doc/) 目录，由 [gomarkdoc](https://github.com/princjef/gomarkdoc) 从 Go 源码注释自动生成。从 [`doc/README.md`](doc/README.md) 开始查看索引。

也可以用 Go 官方文档服务器在本地浏览：

```bash
go install golang.org/x/pkgsite/cmd/pkgsite@latest
pkgsite -http=:6060
# 打开 http://localhost:6060/github.com/shaobaobaoer/solarsage-mcp
```

---

## 架构

```
cmd/
  server/          MCP 服务器入口（JSON-RPC over stdio）
  api/             REST API 服务器入口（net/http）
pkg/
  solarsage/       高级便捷 API（推荐入口）
  mcp/             MCP 协议处理器（40 个工具）
  api/             REST API 处理器（40 个接口）
  chart/           星盘计算（行星位置、宫位、相位）
  transit/         过境事件检测引擎（100% 精度验证）
  progressions/    二次推运 & 太阳弧推运
  returns/         太阳 & 月亮回归盘
  composite/       合盘（中点法）& 戴维森盘
  synastry/        合盘兼容性评分
  primary/         主限推运（托勒密半弧法、纳伯德）
  symbolic/        符号推运（每年 1°、纳伯德、小限、自定义）
  dignity/         本质尊贵、互容、昼夜派
  dispositor/      主星链条 & 最终主星
  report/          综合星盘分析报告
  vedic/           恒星时盘、星宿、维姆肖塔里大运
  divisional/      吠陀 Varga 盘（D1-D60）
  ashtakavarga/    八宫分点 Bindu 表
  yoga/            吠陀瑜伽检测
  profection/      年度 & 月度小限
  firdaria/        菲尔达利亚行星周期系统
  lots/            阿拉伯幸运点计算器
  bounds/          迦勒底旬区 & 埃及界
  antiscia/        对镜点 & 逆对镜点
  fixedstars/      恒星目录 & 合相检测
  midpoint/        中点分析 & 宇宙生物学盘
  harmonic/        谐波（分部）盘
  planetary/       迦勒底行星时
  heliacal/        偕日升落
  lunar/           月相 & 食相检测
  render/          星盘轮可视化坐标
  models/          核心数据类型和常量
  julian/          儒略日转换（ISO 8601 ↔ JD）
  geo/             地理编码和时区查找
  export/          CSV/JSON 导出工具
  sweph/           瑞士星历表 C 绑定（CGO，线程安全互斥锁）
internal/
  aspect/          相位计算 & 7 种模式检测引擎
third_party/
  swisseph/        瑞士星历表 C 源码 + 头文件 + libswe.a + ephe/
```

### 关键设计决策

- **CGO + 静态库** — 瑞士星历表编译为 `libswe.a` 并静态链接。无运行时 `.so` 依赖，零安装摩擦。
- **`pkg/sweph` 中的全局互斥锁** — 瑞士星历表 C 库不是线程安全的。所有 C 调用通过单个 `sync.Mutex` 串行化。
- **`pkg/solarsage` 作为稳定 API** — 高级包封装了所有底层包，提供稳定、符合人体工程学的接口。底层包故意保持单一职责。
- **过境精度门控** — `pkg/transit/solarfire_test.go` 以 100% 匹配率验证 247 个事件（对照 Solar Fire 9 参考 CSV）。对 `transit.go` 的任何修改都必须保持这一精度。

---

## 性能

| 操作 | 耗时 | 吞吐量 |
|------|------|--------|
| 行星位置计算 | ~380 纳秒 | ~260 万次/秒 |
| 本命盘（10 行星） | ~80 微秒 | ~12,400 次/秒 |
| 双盘 + 跨盘相位 | ~347 微秒 | ~2,880 次/秒 |
| 30 天过境扫描（5 行星） | ~764 毫秒 | — |
| 1 年过境扫描（外行星） | ~2.1 秒 | — |

运行 `make bench` 在你的硬件上复现测试结果。

---

## 精度

经过独立验证，**事件完全匹配率 100%**（247/247 过境事件），涵盖 1 年时间段内全部 7 种图表类型组合（Tr-Na、Tr-Tr、Tr-Sp、Tr-Sa、Sp-Na、Sp-Sp、Sa-Na），以行业标准桌面占星软件 Solar Fire 9 为基准进行验证。

验证测试位于 `pkg/transit/solarfire_test.go`，作为 `make test` 的一部分运行。

---

## Docker

```bash
# 构建镜像（在容器内编译瑞士星历表 + Go 二进制文件）
docker build -t solarsage-mcp .

# 运行 MCP 服务器
docker run -i solarsage-mcp

# 运行 REST API 服务器
docker run -p 8080:8080 --entrypoint solarsage-api solarsage-mcp --port 8080
```

---

## 贡献指南

开发环境搭建（含瑞士星历表构建说明）和贡献规范，请参阅 [CONTRIBUTING.md](CONTRIBUTING.md)。

---

## 许可证

MIT — 详见 [LICENSE](LICENSE)。

瑞士星历表依据 AGPL-3.0 授权（或 Astrodienst 商业授权）。详见 `third_party/swisseph/LICENSE`。
