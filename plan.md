仔细解读这份 Solar Fire 输出文件，梳理其列结构：

```
P1 (H)  Asp  P2 (H)  EXL  Type      Date         Time          Zone   Age     Pos1          Pos2
¿ (11)  –    ¸ (6)   (B)  Tr-Na   Feb 1 2026   00:00:00 am   AWST   28.122  27°Ý27' Þ    26°ˆ29' Þ
¶ (1)   á    ¾ (10)  (V)  Tr-Tr   Feb 1 2026   05:51:47 am   AWST   28.122  28°ƒ37' Þ    28°‹37' Þ
¿ (11)  SÞ        (D)  Tr      Feb 4 2026   10:33:06 am   AWST   28.131  27°Ý27' Þ
```

关键字段解读：
- **EXL**：`(B)` = Beginning（进入容许度）、`(V)` = Exact（精确交角）、`(X)` = Exact crossing（精确合）、`(E)` = Exact、`(L)` = Leaving（离开容许度）、`(S)` = Sign ingress、`(H)` = House ingress、`(D)` = Station Direct、`(R)` = Station Retrograde
- **Type**：`Tr-Na`、`Tr-Tr`、`Tr-Sp`、`Tr-Sa`、`Sp-Na`、`Sp-Sp`、`Sa-Na` 等，表示盘对关系
- **Pos1/Pos2**：事件时刻两个天体各自的黄道经度，逆行用 `Œ` 表示，顺行用 `Þ`
- **Age**：推运年龄（儒略年）

---

现在写完整规范化文档：

---

# 基建篇：Swisseph MCP 需求文档

## 1. 概述

**模块名称：** Swisseph MCP（占星计算 MCP 服务）

**模块定位：** 占星系统底层基础设施，提供标准化、高精度的星盘计算能力。所有上层业务均依赖本模块。

**技术依据：** 基于瑞士星历表（Swiss Ephemeris）封装，通过 MCP（Model Context Protocol）协议对外提供服务。精度参考 JPL DE430 星历表，时间分辨率约为 1 秒。

---

## 2. 基础工具能力

### 2.1 地点名称 → 经纬度坐标

```
输入：
  - location_name: string       # 支持中英文，如 "北京"、"London"

输出：
  - latitude: float             # 纬度，北纬为正（-90 ~ +90）
  - longitude: float            # 经度，东经为正（-180 ~ +180）
  - timezone: string            # 所属时区，如 "Asia/Shanghai"
  - display_name: string        # 标准化地名
```

---

### 2.2 公历时间 → 儒略日

```
输入：
  - datetime: string            # ISO 8601 格式，如 "1990-06-15T08:30:00+08:00"
  - calendar: enum              # GREGORIAN | JULIAN，默认 GREGORIAN

输出：
  - jd_ut: float                # 世界时儒略日（UT）
  - jd_tt: float                # 地球时儒略日（TT）
```

---

### 2.3 儒略日 → 公历时间

```
输入：
  - jd: float
  - timezone: string            # 目标时区，默认 UTC

输出：
  - datetime: string            # ISO 8601 格式
```

---

## 3. 星盘计算能力

### 3.1 静态计算

#### 3.1.1 单盘计算

**适用场景：** 出生盘、天象盘

```
输入：
  - latitude: float
  - longitude: float
  - jd_ut: float
  - planets: list[PlanetID]
  - orb_config: OrbConfig
  - house_system: HouseSystem

输出：
  ChartInfo
  ├── planets: list[PlanetPosition]
  │     ├── planet_id: PlanetID
  │     ├── longitude: float          # 黄道经度（0–360°）
  │     ├── latitude: float           # 黄道纬度
  │     ├── speed: float              # 经度速度（°/天，负值=逆行）
  │     ├── is_retrograde: bool
  │     ├── sign: string              # 所在星座
  │     ├── sign_degree: float        # 星座内度数（0–30°）
  │     └── house: int                # 所在宫位（1–12）
  │
  ├── houses: list[float]             # 12 宫头黄道经度，index 0 = 第一宫
  │
  ├── angles: AnglesInfo
  │     ├── asc: float
  │     ├── mc: float
  │     ├── dsc: float
  │     └── ic: float
  │
  └── aspects: list[AspectInfo]
        ├── planet_a: PlanetID
        ├── planet_b: PlanetID
        ├── aspect_type: string
        ├── aspect_angle: float
        ├── actual_angle: float
        ├── orb: float
        └── is_applying: bool
```

---

#### 3.1.2 双盘计算

**适用场景：** 行运盘（Transit）、比较盘（Synastry）

```
输入：
  - inner_latitude: float
  - inner_longitude: float
  - inner_jd_ut: float
  - inner_planets: list[PlanetID]

  - outer_latitude: float
  - outer_longitude: float
  - outer_jd_ut: float
  - outer_planets: list[PlanetID]

  - special_points: SpecialPointsConfig         # 可选
        ├── inner_points: list[SpecialPointID]
        └── outer_points: list[SpecialPointID]

  - orb_config: OrbConfig
  - house_system: HouseSystem

输出：
  - inner_chart: ChartInfo
  - outer_chart: ChartInfo
  - cross_aspects: list[CrossAspectInfo]
        ├── inner_body: PlanetID | SpecialPointID
        ├── outer_body: PlanetID | SpecialPointID
        ├── aspect_type: string
        ├── aspect_angle: float
        ├── actual_angle: float
        ├── orb: float
        └── is_applying: bool
```

---

### 3.2 动态计算

#### 核心设计：分治法 + 二分搜索

动态计算基于两类相位追踪问题：

**RQ1（动态-静态）：** 运动点 vs 固定参照点（本命行星、固定宫头、星座边界）。在每个顺逆行周期内行星黄经单调，使用**二分法**精确定位事件时刻。

**RQ2（动态-动态）：** 运动点 vs 运动点。将两颗行星的顺逆行周期取**交集**，在交集内两者相对运动单调，需以细步长**扫描**定位（不能用二分）。

**盘类型对应关系：**

| 事件类型 | 运动方 | 参照方 | 追踪类型 |
|---|---|---|---|
| Tr-Na | 行运盘（Transits） | 本命盘（Radix） | RQ1 |
| Tr-Tr | 行运盘（Transits） | 行运盘（Transits） | RQ2 |
| Tr-Sp | 行运盘（Transits） | 次限推运盘（Progressions） | RQ2 |
| Tr-Sa | 行运盘（Transits） | 太阳弧方向盘（Solar Arc） | RQ2 |
| Sp-Na | 次限推运盘（Progressions） | 本命盘（Radix） | RQ1 |
| Sp-Sp | 次限推运盘（Progressions） | 次限推运盘（Progressions） | RQ2 |
| Sa-Na | 太阳弧方向盘（Solar Arc） | 本命盘（Radix） | RQ1 |

**交角归一化：**
```
Δθ = wrap((θ₁ - θ₂) × sgn(d), -180°, 180°)
顺行 sgn(d) = +1，逆行 sgn(d) = -1
```
此方式可精确区分入相/离相，并天然支持逆行三击。

---

#### 3.2.1 推运计算

**适用场景：** 在指定时间范围内，搜索所有占星事件的精确时刻。

**接口定义：**

```
输入：
  # 本命盘（Radix，固定不变）
  - natal_latitude: float
  - natal_longitude: float
  - natal_jd_ut: float
  - natal_planets: list[PlanetID]

  # 推运地点
  - transit_latitude: float
  - transit_longitude: float

  # 推算时间范围
  - start_jd_ut: float
  - end_jd_ut: float

  # 各类推运盘配置（可选，按需开启）
  - progressions_config: ProgressionsConfig     # 次限推运盘配置
        ├── enabled: bool
        └── planets: list[PlanetID]

  - solar_arc_config: SolarArcConfig            # 太阳弧方向盘配置
        ├── enabled: bool
        └── planets: list[PlanetID]

  - transit_planets: list[PlanetID]             # 行运天体列表

  # 特殊点配置（可选）
  - special_points: SpecialPointsConfig
        ├── natal_points: list[SpecialPointID]
        ├── transit_points: list[SpecialPointID]
        ├── progressions_points: list[SpecialPointID]
        └── solar_arc_points: list[SpecialPointID]

  # 事件类型配置（可选，默认全部开启）
  - event_config: EventConfig
        ├── include_tr_na: bool       # 行运 × 本命相位
        ├── include_tr_tr: bool       # 行运 × 行运相位
        ├── include_tr_sp: bool       # 行运 × 次限推运相位
        ├── include_tr_sa: bool       # 行运 × 太阳弧相位
        ├── include_sp_na: bool       # 次限推运 × 本命相位
        ├── include_sp_sp: bool       # 次限推运 × 次限推运相位
        ├── include_sa_na: bool       # 太阳弧 × 本命相位
        ├── include_sign_ingress: bool
        ├── include_house_ingress: bool
        ├── include_station: bool
        └── include_void_of_course: bool

  # 各盘独立容许度配置
  - orb_config_transit: OrbConfig       # 行运盘容许度
  - orb_config_progressions: OrbConfig  # 次限推运盘容许度
  - orb_config_solar_arc: OrbConfig     # 太阳弧盘容许度

  - house_system: HouseSystem
```

```
输出：
  - events: list[TransitEvent]          # 所有事件，按 jd 升序排列

  # ── 所有事件公共字段 ──────────────────────────────────────────────
  TransitEvent:
    - event_type: enum                  # 见下方类型定义
    - chart_type: enum                  # 触发方所属盘类型
                                        # TRANSIT | PROGRESSIONS | SOLAR_ARC
    - planet: PlanetID                  # 触发事件的天体
    - jd: float                         # 事件发生时刻（儒略日，精确到秒）
    - age: float                        # 推运年龄（儒略年，如 28.122）
    - planet_longitude: float           # 事件时刻天体黄道经度（0–360°）
    - planet_sign: string               # 事件时刻天体所在星座
    - planet_house: int                 # 事件时刻天体所在宫位（本命固定宫头，1–12）
    - is_retrograde: bool               # 事件时刻天体是否逆行

  # ── event_type = ASPECT_ENTER ────────────────────────────────────
  # 天体进入与参照点的相位容许度
    - target_chart_type: enum           # 参照方所属盘 NATAL | TRANSIT | PROGRESSIONS | SOLAR_ARC
    - target: PlanetID | SpecialPointID
    - target_longitude: float           # 事件时刻参照天体黄道经度
    - target_sign: string
    - target_house: int
    - target_is_retrograde: bool
    - aspect_type: string
    - aspect_angle: float
    - orb_at_enter: float

  # ── event_type = ASPECT_EXACT ────────────────────────────────────
  # 天体与参照点达到精确交角
  # 三击时产生 3 个独立 ASPECT_EXACT，exact_count 标记第几击
    - target_chart_type: enum
    - target: PlanetID | SpecialPointID
    - target_longitude: float
    - target_sign: string
    - target_house: int
    - target_is_retrograde: bool
    - aspect_type: string
    - aspect_angle: float
    - exact_count: int                  # 第几次精确交角（1/2/3）

  # ── event_type = ASPECT_LEAVE ────────────────────────────────────
  # 天体离开相位容许度
    - target_chart_type: enum
    - target: PlanetID | SpecialPointID
    - target_longitude: float
    - target_sign: string
    - target_house: int
    - target_is_retrograde: bool
    - aspect_type: string
    - aspect_angle: float
    - orb_at_leave: float

  # ── event_type = SIGN_INGRESS ────────────────────────────────────
  # 天体进入新星座（对应 Solar Fire 中 EXL=(S)）
    - from_sign: string
    - to_sign: string

  # ── event_type = HOUSE_INGRESS ───────────────────────────────────
  # 天体进入新宫位（对应 Solar Fire 中 EXL=(H)）
  # 宫头使用本命盘固定宫头
    - from_house: int
    - to_house: int

  # ── event_type = STATION ─────────────────────────────────────────
  # 天体顺逆行切换（对应 Solar Fire 中 SÞ(D) / SŒ(R)）
    - station_type: enum                # RETROGRADE（顺转逆）| DIRECT（逆转顺）

  # ── event_type = VOID_OF_COURSE ──────────────────────────────────
  # 月亮空亡：月亮最后一个相位离开至换座之间的时段
  # 派生事件，由月亮相位事件与换座事件二次计算得出
    - void_start_jd: float             # 空亡开始（最后一个相位离开时刻）
    - void_end_jd: float               # 空亡结束（月亮进入下一星座时刻）
    - last_aspect_type: string         # 最后一个相位类型
    - last_aspect_target: PlanetID     # 最后相位的参照天体
    - next_sign: string                # 空亡结束后进入的星座
```

---

**输出示例（对照 Solar Fire 格式）：**

Solar Fire 原始行：
```
¿ (11)  –  ¸ (6)  (B)  Tr-Na  Feb 1 2026  00:00:00 am  AWST  28.122  27°Ý27' Þ  26°ˆ29' Þ
¿ (11)  –  ¸ (6)  (L)  Tr-Na  Feb 13 2026  10:13:23 pm  AWST  28.157  27°Ý29' Þ  26°ˆ29' Þ
¿ (11)  SÞ      (D)  Tr   Feb 4 2026  10:33:06 am  AWST  28.131  27°Ý27' Þ
¶ (2)  ß  „ (2)  (S)  Tr-Tr  Feb 1 2026  08:08:52 am  AWST  28.123  00°„00' Þ  00°„00' Þ
¿ (12)  ß  Hs (12)  (H)  Tr-Na  Jun 17 2026  10:11:24 am  AWST  28.495  02°‚58' Þ  02°‚58' Þ
```

对应本系统输出：
```json
[
  {
    "event_type": "ASPECT_ENTER",
    "chart_type": "TRANSIT",
    "planet": "NEPTUNE",
    "jd": 2461042.5000,
    "age": 28.122,
    "planet_longitude": 357.45,
    "planet_sign": "双鱼座",
    "planet_house": 11,
    "is_retrograde": false,
    "target_chart_type": "NATAL",
    "target": "NATAL_SATURN",
    "target_longitude": 266.48,
    "target_sign": "射手座",
    "target_house": 6,
    "target_is_retrograde": false,
    "aspect_type": "对分相",
    "aspect_angle": 180.0,
    "orb_at_enter": 6.97
  },
  {
    "event_type": "ASPECT_LEAVE",
    "chart_type": "TRANSIT",
    "planet": "NEPTUNE",
    "jd": 2461054.9260,
    "age": 28.157,
    "planet_longitude": 357.48,
    "planet_sign": "双鱼座",
    "planet_house": 11,
    "is_retrograde": false,
    "target_chart_type": "NATAL",
    "target": "NATAL_SATURN",
    "target_longitude": 266.48,
    "target_sign": "射手座",
    "target_house": 6,
    "target_is_retrograde": false,
    "aspect_type": "对分相",
    "aspect_angle": 180.0,
    "orb_at_leave": 6.98
  },
  {
    "event_type": "STATION",
    "chart_type": "TRANSIT",
    "planet": "NEPTUNE",
    "jd": 2461046.9410,
    "age": 28.131,
    "planet_longitude": 357.45,
    "planet_sign": "双鱼座",
    "planet_house": 11,
    "is_retrograde": false,
    "station_type": "DIRECT"
  },
  {
    "event_type": "SIGN_INGRESS",
    "chart_type": "TRANSIT",
    "planet": "MOON",
    "jd": 2461042.8395,
    "age": 28.123,
    "planet_longitude": 60.0,
    "planet_sign": "双子座",
    "planet_house": 2,
    "is_retrograde": false,
    "from_sign": "金牛座",
    "to_sign": "双子座"
  },
  {
    "event_type": "HOUSE_INGRESS",
    "chart_type": "TRANSIT",
    "planet": "NEPTUNE",
    "jd": 2461189.9247,
    "age": 28.495,
    "planet_longitude": 357.97,
    "planet_sign": "双鱼座",
    "planet_house": 12,
    "is_retrograde": false,
    "from_house": 11,
    "to_house": 12
  }
]
```

---

## 4. 附录

### 附录 A：天体标识符（PlanetID）

| 标识符 | 名称 | 备注 |
|---|---|---|
| SUN | 太阳 | |
| MOON | 月亮 | |
| MERCURY | 水星 | |
| VENUS | 金星 | |
| MARS | 火星 | |
| JUPITER | 木星 | |
| SATURN | 土星 | |
| URANUS | 天王星 | |
| NEPTUNE | 海王星 | |
| PLUTO | 冥王星 | |
| CHIRON | 凯龙星 | 小行星 |
| NORTH_NODE_TRUE | 北交点（真实值） | |
| NORTH_NODE_MEAN | 北交点（平均值） | |
| SOUTH_NODE | 南交点 | |
| LILITH_MEAN | 黑月莉莉丝（平均值） | |
| LILITH_TRUE | 黑月莉莉丝（真实值） | |

---

### 附录 B：特殊点标识符（SpecialPointID）

| 标识符 | 名称 | 经度来源 |
|---|---|---|
| ASC | 上升点 | 所属盘坐标 + 时间动态计算 |
| MC | 中天 | 所属盘坐标 + 时间动态计算 |
| DSC | 下降点 | ASC + 180° |
| IC | 天底 | MC + 180° |
| VERTEX | 命运点 | 所属盘坐标 + 时间动态计算 |
| ANTI_VERTEX | 反命运点 | VERTEX + 180° |
| EAST_POINT | 东点 | 所属盘坐标 + 时间动态计算 |
| LOT_FORTUNE | 幸运点 | 日间：ASC + 月亮 − 太阳；夜间：反转 |
| LOT_SPIRIT | 精神点 | 日间：ASC + 太阳 − 月亮；夜间：反转 |

> 动态推运中，本命特殊点（natal_points）为固定值，参与 RQ1 计算；行运/推运特殊点随时间变化，参与 RQ2 计算。

---

### 附录 C：宫位系统（HouseSystem）

| 标识符 | 名称 | 说明 |
|---|---|---|
| PLACIDUS | 普拉西德 | 最常用，时间等分 |
| KOCH | 科赫 | 基于出生地纬度 |
| EQUAL | 等宫制 | 从 ASC 起每 30° 一宫 |
| WHOLE_SIGN | 整宫制 | ASC 所在星座为第一宫 |
| CAMPANUS | 坎帕努斯 | 空间等分 |
| REGIOMONTANUS | 雷吉奥蒙塔努斯 | 天赤道等分 |
| PORPHYRY | 波菲利 | 三等分四轴间距 |

---

### 附录 D：相位容许度配置（OrbConfig）

| 字段 | 相位名称 | 标准角度 | 建议默认值 |
|---|---|---|---|
| conjunction | 合相 | 0° | 8° |
| opposition | 对分相 | 180° | 8° |
| trine | 三分相 | 120° | 7° |
| square | 刑相 | 90° | 7° |
| sextile | 六分相 | 60° | 5° |
| quincunx | 补十二分相 | 150° | 3° |
| semi_sextile | 十二分相 | 30° | 2° |
| semi_square | 八分相 | 45° | 2° |
| sesquiquadrate | 倍半刑 | 135° | 2° |

---

### 附录 E：Solar Fire EXL 字段对照

| Solar Fire EXL | 本系统 event_type | 说明 |
|---|---|---|
| (B) Beginning | ASPECT_ENTER | 进入容许度 |
| (V) / (E) Exact | ASPECT_EXACT | 精确交角 |
| (L) Leaving | ASPECT_LEAVE | 离开容许度 |
| (S) Sign ingress | SIGN_INGRESS | 换座 |
| (H) House ingress | HOUSE_INGRESS | 变宫 |
| SÞ (D) Station Direct | STATION（DIRECT） | 逆转顺站点 |
| SŒ (R) Station Retrograde | STATION（RETROGRADE） | 顺转逆站点 |

---

### 附录 F：Solar Fire Type 字段对照

| Solar Fire Type | chart_type | target_chart_type | 追踪类型 |
|---|---|---|---|
| Tr-Na | TRANSIT | NATAL | RQ1 |
| Tr-Tr | TRANSIT | TRANSIT | RQ2 |
| Tr-Sp | TRANSIT | PROGRESSIONS | RQ2 |
| Tr-Sa | TRANSIT | SOLAR_ARC | RQ2 |
| Sp-Na | PROGRESSIONS | NATAL | RQ1 |
| Sp-Sp | PROGRESSIONS | PROGRESSIONS | RQ2 |
| Sa-Na | SOLAR_ARC | NATAL | RQ1 |

