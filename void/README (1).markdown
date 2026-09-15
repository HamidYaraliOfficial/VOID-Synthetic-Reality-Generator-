# VOID — Synthetic Reality Generator

**A distributed, deterministic engine for building, simulating, and stress-testing synthetic realities at massive scale.**

<p align="center">
  <a href="#english">English</a> ·
  <a href="#فارسی">فارسی</a> ·
  <a href="#中文">中文</a>
</p>

---

<a id="english"></a>
## 🇬🇧 English

### Overview

**VOID** lets you spin up a fully synthetic world — populated with millions of Users, Customers, Employees, Companies, Products, Accounts, Devices, Vehicles, Locations, Transactions and more — and watch it evolve on its own. Every simulated entity makes decisions based on rules, probabilities, goals, and history, not static scripts. VOID is built for engineers and researchers who need realistic, controllable, *synthetic* data and load: testing fraud detection, load-testing a staging API with human-shaped traffic, running "what if our biggest customer churned" business simulations, or generating large anonymized datasets for analytics.

The **simulation core is written in Go** — concurrent, deterministic, horizontally scalable, and fully headless-capable. The **Control Center dashboard is a Next.js/TypeScript app** with a Windows 11-inspired design system, full English/Persian/Chinese localization (with correct RTL/LTR handling), and five built-in themes.

> **Note on scope:** this repository ships a complete, working, end-to-end scaffold of every subsystem described below — world generation, entity/behavior modeling, the event engine, population/statistics generation, scenario injection, chaos engineering, load-testing adapters, social/business/financial/network/IoT simulators, snapshots & replay, a REST + WebSocket API, a CLI, and a themed/localized dashboard — all built on the standard library with zero required external dependencies, and verified to build, vet, test, and run end-to-end. Swapping in production-grade Postgres/Redis/NATS/S3 drivers behind the included storage interfaces, and hardening auth/RBAC for a multi-tenant deployment, is the natural next step for a production rollout at internet scale.

### Key Features

- **Simulation Core (Go):** goroutine worker pools, channels, atomic counters, batched processing, and memory-pooled ticking so millions of entities can be evaluated per tick with bounded resource use.
- **Generic Entity Model:** any Archetype (Customer, Employee, Company, Product, Account, Device, Vehicle, Server, Social Account, Driver, Location, or your own) with attributes, state, goals, memory, relationships, resources, skills and a risk profile.
- **Behavior Engine:** rule-based conditions, a state machine, and a utility-AI goal scorer — entities *decide*, they don't just replay a script.
- **Population Generator:** Normal, Uniform, Gaussian Mixture, Power Law, Zipf, Poisson and Custom distributions, with cross-attribute correlation, noise, missing-data and outlier controls for realistic synthetic datasets.
- **Event Engine:** a priority, tick-ordered bus with dependency gating, backpressure, and a dead-letter queue — every world change is a structured Event with trigger, conditions, payload, source/target and consequences.
- **Distributed Simulation Coordinator:** partitions entities across Workers, tracks heartbeats/health, and reassigns partitions from stale workers automatically.
- **Scenario Builder + Template Library:** 19 built-in scenarios (sudden user surge, service attack, bankruptcy, viral growth, DB crash, network outage, buying frenzy, fraud spike, resource shortage, customer behavior shift) plus vertical starter templates (e-commerce, banking, SaaS, social, IoT, logistics, cyber range, smart city, cloud infra).
- **Chaos Engine:** fault injection for network delay, packet loss, service/database/worker failure, resource starvation and dependency failure, with an impact report.
- **Load Testing Generator:** turns simulated behavior into real HTTP/TCP/WebSocket traffic against a target system, with ramp-up/down, concurrency, and rate limiting (gRPC/broker adapters plug into the same interface).
- **Vertical Simulators:** Social Network (follow graph, virality, community detection), Business (companies/market share), Finance (payments/transfers/fraud), Network Topology (dependency-aware failure cascades), and IoT/Smart City (sensors, traffic signals, energy).
- **State Management:** Snapshots with save/restore/clone/fork/diff/rollback and full versioning; a Replay Engine for exact, seed-based reproduction.
- **Telemetry:** a Prometheus-compatible `/metrics` endpoint plus a JSON snapshot for the dashboard — entity counts, tick rate, queue depth, dead-letters, heap, goroutines.
- **Scheduler & fully user-defined Operating Hours:** queue/retry long-running jobs, and configure — entirely from the UI, with no hardcoded defaults — which weekdays and clock-hours the system is considered "open", with a live countdown to the next open/close transition.
- **Control Center Dashboard:** worlds, entities, scenarios, simulations, workers, metrics, and event explorer, real-time via WebSocket, in **English, Persian (RTL) and Chinese**, with **Light, Dark, Windows-default, Red and Blue** themes.
- **Export:** JSON, NDJSON, CSV, and a Parquet-compatible columnar format for downstream Big Data pipelines.
- **CLI:** `serve`, `create-world`, `run` / `run-headless`, `generate-dataset`, `benchmark`.

### Project Structure

```
void/
├── backend/                  # Go simulation core + API
│   ├── cmd/void/              # CLI entrypoint
│   └── internal/
│       ├── entity/            # Generic entity model + archetypes
│       ├── behavior/           # Rule engine, state machine, utility AI
│       ├── population/         # Statistical distributions + generator
│       ├── event/               # Priority event bus
│       ├── world/                # World Builder
│       ├── simulation/            # Concurrent tick engine
│       ├── coordinator/            # Distributed partitioning/health
│       ├── scenario/                # Scenario Builder + template library
│       ├── chaos/                    # Fault injection engine
│       ├── loadtest/                  # HTTP/TCP/WS load adapters
│       ├── social/ business/           # Vertical simulators
│       ├── finance/ network/ iot/
│       ├── snapshot/ replay/ export/    # State management + I/O
│       ├── scheduler/                    # Jobs + Operating Hours
│       ├── storage/                       # Postgres/Redis/NATS/S3-ready DAL
│       ├── metrics/                        # Prometheus-compatible telemetry
│       └── api/                             # REST + WebSocket server
├── frontend/                 # Next.js + TypeScript Control Center
│   ├── app/                   # Pages (dashboard, worlds, entities, ...)
│   ├── components/             # Sidebar, theme/language switchers, charts
│   └── lib/                     # i18n catalogs + typed API client
├── deploy/helm/               # Kubernetes Helm chart
├── docker-compose.yml
└── Makefile
```

### Getting Started

**Prerequisites:** Go ≥ 1.22, Node.js ≥ 20, (optionally) Docker & Docker Compose.

**1. Run the backend API:**

```bash
cd backend
go build -o void ./cmd/void
./void serve
# API listening on :8080 — try curl localhost:8080/healthz
```

**2. Run the dashboard:**

```bash
cd frontend
npm install
npm run dev
# Dashboard on :3000, talking to the API at NEXT_PUBLIC_API_BASE (default :8080)
```

**3. Or run everything with Docker Compose:**

```bash
docker compose up --build
# API      → http://localhost:8080
# Dashboard → http://localhost:3000
```

**4. Or drive it entirely from the CLI (no dashboard needed):**

```bash
cd backend
go run ./cmd/void create-world --seed 42 --size medium
go run ./cmd/void run --agents 100000 --ticks 500 --workers 8
go run ./cmd/void generate-dataset --agents 50000 --out dataset.csv --format csv
go run ./cmd/void benchmark --agents 1000000 --ticks 50 --workers 16
```

### Configuration

All configuration is via environment variables (see `.env.example`):

| Variable | Purpose | Default |
|---|---|---|
| `VOID_HTTP_ADDR` | API bind address | `:8080` |
| `VOID_DATA_DIR` | Local data/export/snapshot directory | `./data` |
| `VOID_WORKER_COUNT` | Default simulation worker pool size | `8` |
| `VOID_JWT_SECRET` | Auth signing secret (set a real one in production) | — |
| `VOID_POSTGRES_DSN` / `VOID_REDIS_ADDR` / `VOID_NATS_URL` / `VOID_S3_ENDPOINT` | Optional durable backends behind the storage interfaces | blank = in-memory/file |
| `NEXT_PUBLIC_API_BASE` | API URL the dashboard calls | `http://localhost:8080` |

### Operating Hours (fully user-configurable)

Under **Settings → Operating Hours** in the dashboard (or `PUT /api/v1/scheduler/operating-hours`), you define — with no presets — exactly which weekdays and clock times the system should be treated as "open" for scheduled runs, in any IANA timezone. The panel shows a live open/closed indicator and a countdown to the next transition, computed entirely from what you entered.

### API

The full REST + WebSocket surface is documented at `GET /openapi.json` once the server is running. Highlights: `POST /api/v1/worlds`, `POST /api/v1/simulations`, `POST /api/v1/simulations/{id}/scenarios/{scenarioId}/inject`, `POST /api/v1/chaos/faults`, `GET /metrics` (Prometheus format), `GET /ws/simulations/{id}` (live tick stream).

### License

Provided as-is for evaluation and further development. Review and adapt security, authentication, and storage backends before any production or public deployment.

---

<a id="فارسی"></a>
<div dir="rtl" align="right">

## 🇮🇷 فارسی

### معرفی

**وید (VOID)** به شما امکان می‌دهد یک جهان کاملاً مصنوعی بسازید — با میلیون‌ها کاربر، مشتری، کارمند، شرکت، محصول، حساب، دستگاه، خودرو، مکان، تراکنش و موارد دیگر — و ببینید که چگونه خودش تکامل می‌یابد. هر موجودیت شبیه‌سازی‌شده بر اساس قوانین، احتمالات، اهداف و تاریخچه تصمیم می‌گیرد، نه یک اسکریپت ثابت. وید برای مهندسان و پژوهشگرانی ساخته شده که به داده و بار *مصنوعی*، واقع‌گرا و قابل‌کنترل نیاز دارند: آزمایش سیستم‌های تشخیص تقلب، تست بار یک API آزمایشی با ترافیکی شبیه انسان، اجرای شبیه‌سازی‌های کسب‌وکار مانند «اگر بزرگ‌ترین مشتری ما را از دست بدهیم چه می‌شود»، یا تولید مجموعه‌داده‌های ناشناس بزرگ برای تحلیل.

**هسته شبیه‌سازی با Go نوشته شده است** — همزمان (Concurrent)، قطعی (Deterministic)، مقیاس‌پذیر افقی و کاملاً قابل اجرا در حالت Headless. **داشبورد مرکز کنترل یک اپلیکیشن Next.js/TypeScript** با سیستم طراحی الهام‌گرفته از ویندوز ۱۱، بومی‌سازی کامل به سه زبان انگلیسی، فارسی و چینی (با رعایت صحیح راست‌چین/چپ‌چین) و پنج پوسته آماده است.

> **نکته درباره دامنه پروژه:** این مخزن یک اسکلت کامل، کاربردی و سرتاسری از تمام زیرسیستم‌های توضیح‌داده‌شده در ادامه را ارائه می‌دهد — تولید جهان، مدل موجودیت/رفتار، موتور رویداد، تولید جمعیت/آمار، تزریق سناریو، مهندسی آشوب (Chaos)، آداپتورهای تست بار، شبیه‌سازهای اجتماعی/کسب‌وکار/مالی/شبکه/IoT، اسنپ‌شات و بازپخش، یک API از نوع REST + WebSocket، یک CLI، و یک داشبورد بومی‌سازی‌شده و دارای پوسته — همگی صرفاً با کتابخانه استاندارد Go و بدون هیچ وابستگی بیرونی الزامی ساخته شده و از نظر Build، Vet، تست و اجرای سرتاسری تأیید شده‌اند. جایگزینی درایورهای واقعی Postgres/Redis/NATS/S3 پشت رابط‌های ذخیره‌سازی موجود، و سخت‌سازی احراز هویت/RBAC برای استقرار چندمستأجری، گام طبیعی بعدی برای استقرار در مقیاس اینترنتی است.

### ویژگی‌های کلیدی

- **هسته شبیه‌سازی (Go):** استخر ورکر مبتنی بر Goroutine، Channel، شمارنده‌های Atomic، پردازش دسته‌ای و Tick با Memory Pooling تا میلیون‌ها موجودیت با مصرف منابع کنترل‌شده در هر Tick ارزیابی شوند.
- **مدل موجودیت عمومی:** هر کهن‌الگو (Customer، Employee، Company، Product، Account، Device، Vehicle، Server، Social Account، Driver، Location یا کهن‌الگوی دلخواه شما) با ویژگی‌ها، وضعیت، اهداف، حافظه، روابط، منابع، مهارت‌ها و پروفایل ریسک.
- **موتور رفتار:** شرط‌های مبتنی بر قانون، ماشین حالت، و امتیازدهی هدف مبتنی بر Utility AI — موجودیت‌ها *تصمیم می‌گیرند*، نه اینکه فقط یک اسکریپت را بازپخش کنند.
- **تولیدکننده جمعیت:** توزیع‌های Normal، Uniform، Gaussian Mixture، Power Law، Zipf، Poisson و سفارشی، همراه با همبستگی بین ویژگی‌ها، نویز، داده گم‌شده و کنترل داده‌های پرت برای مجموعه‌داده‌های مصنوعی واقع‌گرا.
- **موتور رویداد:** یک باس اولویت‌دار و مرتب بر اساس Tick با کنترل وابستگی، Backpressure و صف پیام‌های مرده (Dead-Letter) — هر تغییر در جهان یک رویداد ساختاریافته با Trigger، شرط‌ها، Payload، منبع/هدف و پیامدهاست.
- **هماهنگ‌کننده شبیه‌سازی توزیع‌شده:** موجودیت‌ها را بین ورکرها تقسیم می‌کند، ضربان و سلامت را پیگیری می‌کند و پارتیشن ورکرهای از کار افتاده را به‌طور خودکار به دیگران واگذار می‌کند.
- **سازنده سناریو + کتابخانه قالب:** ۱۹ سناریوی آماده (افزایش ناگهانی کاربر، حمله به سرویس، ورشکستگی، رشد ویروسی، Crash دیتابیس، اختلال شبکه، هجوم خرید، افزایش تقلب، کمبود منابع، تغییر رفتار مشتری) به‌همراه قالب‌های شروع برای حوزه‌های مختلف (فروشگاه آنلاین، بانکداری، SaaS، شبکه اجتماعی، IoT، لجستیک، Cyber Range، شهر هوشمند، زیرساخت ابری).
- **موتور آشوب (Chaos):** تزریق خطا برای تأخیر شبکه، از دست رفتن بسته، خرابی سرویس/دیتابیس/ورکر، کمبود منابع و خرابی وابستگی، همراه با گزارش اثر.
- **تولیدکننده تست بار:** رفتار شبیه‌سازی‌شده را به ترافیک واقعی HTTP/TCP/WebSocket علیه یک سیستم هدف تبدیل می‌کند، با Ramp-Up/Down، همزمانی و محدودیت نرخ (آداپتورهای gRPC/بروکر پیام روی همان رابط قابل اتصال‌اند).
- **شبیه‌سازهای تخصصی:** شبکه اجتماعی (گراف فالوئینگ، ویروسی‌شدن، تشخیص جامعه)، کسب‌وکار (شرکت‌ها/سهم بازار)، مالی (پرداخت/انتقال/تقلب)، توپولوژی شبکه (خرابی زنجیره‌ای وابسته)، و IoT/شهر هوشمند (سنسورها، چراغ‌های راهنمایی، انرژی).
- **مدیریت وضعیت:** اسنپ‌شات با Save/Restore/Clone/Fork/Diff/Rollback و نسخه‌گذاری کامل؛ موتور بازپخش برای بازتولید دقیق و مبتنی بر Seed.
- **تله‌متری:** یک Endpoint سازگار با Prometheus در مسیر `/metrics` به‌همراه یک Snapshot از نوع JSON برای داشبورد — شامل تعداد موجودیت، نرخ Tick، عمق صف، پیام‌های مرده، Heap و Goroutine.
- **زمان‌بند و ساعات کاری کاملاً تعریف‌شده توسط کاربر:** صف‌بندی و تلاش مجدد Jobهای طولانی، و پیکربندی — کاملاً از داخل رابط کاربری و بدون هیچ مقدار پیش‌فرض ثابتی — اینکه در چه روزهای هفته و چه ساعاتی سیستم «باز» در نظر گرفته شود، همراه با شمارش معکوس زنده تا انتقال باز/بسته بعدی.
- **داشبورد مرکز کنترل:** جهان‌ها، موجودیت‌ها، سناریوها، شبیه‌سازی‌ها، ورکرها، متریک‌ها و کاوشگر رویداد، به‌صورت بلادرنگ از طریق WebSocket، به **سه زبان انگلیسی، فارسی (راست‌چین) و چینی**، با پوسته‌های **روشن، تیره، پیش‌فرض ویندوز، قرمز و آبی**.
- **خروجی‌گیری:** JSON، NDJSON، CSV و یک فرمت ستونی سازگار با Parquet برای پایپ‌لاین‌های Big Data.
- **CLI:** دستورات `serve`، `create-world`، `run` / `run-headless`، `generate-dataset`، `benchmark`.

### ساختار پروژه

```
void/
├── backend/                  # هسته شبیه‌سازی Go + API
│   ├── cmd/void/              # نقطه ورود CLI
│   └── internal/
│       ├── entity/            # مدل موجودیت عمومی + کهن‌الگوها
│       ├── behavior/           # موتور قانون، ماشین حالت، Utility AI
│       ├── population/         # توزیع‌های آماری + تولیدکننده
│       ├── event/               # باس رویداد اولویت‌دار
│       ├── world/                # سازنده جهان
│       ├── simulation/            # موتور Tick همزمان
│       ├── coordinator/            # پارتیشن‌بندی/سلامت توزیع‌شده
│       ├── scenario/                # سازنده سناریو + کتابخانه قالب
│       ├── chaos/                    # موتور تزریق خطا
│       ├── loadtest/                  # آداپتورهای بار HTTP/TCP/WS
│       ├── social/ business/           # شبیه‌سازهای تخصصی
│       ├── finance/ network/ iot/
│       ├── snapshot/ replay/ export/    # مدیریت وضعیت + I/O
│       ├── scheduler/                    # Jobها + ساعات کاری
│       ├── storage/                       # لایه دسترسی داده آماده Postgres/Redis/NATS/S3
│       ├── metrics/                        # تله‌متری سازگار با Prometheus
│       └── api/                             # سرور REST + WebSocket
├── frontend/                 # مرکز کنترل Next.js + TypeScript
│   ├── app/                   # صفحات (داشبورد، جهان‌ها، موجودیت‌ها، ...)
│   ├── components/             # نوار کناری، تعویض‌گر پوسته/زبان، نمودارها
│   └── lib/                     # کاتالوگ‌های i18n + کلاینت API تایپ‌شده
├── deploy/helm/               # چارت Helm برای Kubernetes
├── docker-compose.yml
└── Makefile
```

### شروع به کار

**پیش‌نیازها:** Go نسخه ۱٫۲۲ یا بالاتر، Node.js نسخه ۲۰ یا بالاتر، (اختیاری) Docker و Docker Compose.

**۱. اجرای API بک‌اند:**

```bash
cd backend
go build -o void ./cmd/void
./void serve
# API روی پورت ۸۰۸۰ در دسترس است — تست: curl localhost:8080/healthz
```

**۲. اجرای داشبورد:**

```bash
cd frontend
npm install
npm run dev
# داشبورد روی پورت ۳۰۰۰، متصل به API از طریق NEXT_PUBLIC_API_BASE (پیش‌فرض :8080)
```

**۳. یا اجرای همه‌چیز با Docker Compose:**

```bash
docker compose up --build
# API      → http://localhost:8080
# داشبورد → http://localhost:3000
```

**۴. یا اجرای کامل از طریق CLI (بدون نیاز به داشبورد):**

```bash
cd backend
go run ./cmd/void create-world --seed 42 --size medium
go run ./cmd/void run --agents 100000 --ticks 500 --workers 8
go run ./cmd/void generate-dataset --agents 50000 --out dataset.csv --format csv
go run ./cmd/void benchmark --agents 1000000 --ticks 50 --workers 16
```

### پیکربندی

تمام پیکربندی از طریق متغیرهای محیطی انجام می‌شود (فایل `.env.example` را ببینید):

| متغیر | کاربرد | پیش‌فرض |
|---|---|---|
| `VOID_HTTP_ADDR` | آدرس Bind سرور API | `:8080` |
| `VOID_DATA_DIR` | مسیر داده محلی/خروجی/اسنپ‌شات | `./data` |
| `VOID_WORKER_COUNT` | اندازه پیش‌فرض استخر ورکر شبیه‌سازی | `8` |
| `VOID_JWT_SECRET` | کلید امضای احراز هویت (در Production مقدار واقعی تنظیم کنید) | — |
| `VOID_POSTGRES_DSN` / `VOID_REDIS_ADDR` / `VOID_NATS_URL` / `VOID_S3_ENDPOINT` | زیرساخت‌های ماندگار اختیاری پشت رابط‌های ذخیره‌سازی | خالی = حافظه/فایل |
| `NEXT_PUBLIC_API_BASE` | آدرس API که داشبورد به آن متصل می‌شود | `http://localhost:8080` |

### ساعات کاری (کاملاً قابل تنظیم توسط کاربر)

در مسیر **تنظیمات ← ساعات کاری** در داشبورد (یا از طریق `PUT /api/v1/scheduler/operating-hours`)، شما — بدون هیچ مقدار از پیش تعیین‌شده‌ای — دقیقاً مشخص می‌کنید که سیستم در چه روزهای هفته و چه ساعاتی، در هر منطقه زمانی دلخواه (IANA)، برای اجرای زمان‌بندی‌شده «باز» در نظر گرفته شود. این پنل وضعیت باز/بسته را به‌صورت زنده نمایش می‌دهد و شمارش معکوسی تا انتقال بعدی ارائه می‌دهد که کاملاً بر اساس مقادیر وارد شده توسط شما محاسبه می‌شود.

### API

تمام سطح REST + WebSocket پس از اجرای سرور در مسیر `GET /openapi.json` مستند شده است. نمونه‌ها: `POST /api/v1/worlds`، `POST /api/v1/simulations`، `POST /api/v1/simulations/{id}/scenarios/{scenarioId}/inject`، `POST /api/v1/chaos/faults`، `GET /metrics` (فرمت Prometheus)، `GET /ws/simulations/{id}` (پخش زنده Tick).

### مجوز

این پروژه به‌همان‌صورت (as-is) برای ارزیابی و توسعه بیشتر ارائه شده است. پیش از هرگونه استقرار Production یا عمومی، امنیت، احراز هویت و Backendهای ذخیره‌سازی را بازبینی و متناسب‌سازی کنید.

</div>

---

<a id="中文"></a>
## 🇨🇳 中文

### 概述

**VOID** 让你能够搭建一个完全合成的世界——其中包含数百万用户、客户、员工、公司、产品、账户、设备、车辆、地点、交易等实体——并观察它自行演化。每个被模拟的实体都基于规则、概率、目标和历史做出决策，而不是执行固定脚本。VOID 专为需要真实、可控的*合成*数据与负载的工程师和研究人员而设计：测试欺诈检测系统、用类人流量对预发布环境的 API 进行负载测试、运行"如果我们最大的客户流失会怎样"之类的业务模拟，或生成用于分析的大规模匿名数据集。

**模拟核心使用 Go 编写**——并发、确定性、可水平扩展，并完全支持无头（Headless）运行。**控制中心仪表盘是一个 Next.js/TypeScript 应用**，采用受 Windows 11 启发的设计系统，完整支持英语、波斯语、中文三种语言的本地化（正确处理从右到左/从左到右的文字方向），并内置五种主题。

> **关于项目范围的说明：** 本仓库提供了下文所述每个子系统的完整、可运行、端到端脚手架——世界生成、实体/行为建模、事件引擎、人口/统计数据生成、场景注入、混沌工程、负载测试适配器、社交/商业/金融/网络/物联网模拟器、快照与重放、REST + WebSocket API、命令行工具，以及带主题和本地化的仪表盘——全部仅基于标准库构建，无需任何外部依赖，并已验证可编译、通过静态检查、通过测试并端到端运行。若要在互联网规模上进行生产部署，下一步自然是在已有的存储接口后面接入生产级的 Postgres/Redis/NATS/S3 驱动，并针对多租户部署强化身份验证与基于角色的访问控制（RBAC）。

### 核心功能

- **模拟核心（Go）：** 基于 Goroutine 的工作池、Channel、原子计数器、批处理以及内存池化的 Tick 循环，使数百万实体能够在每个 Tick 中以受控的资源占用被评估。
- **通用实体模型：** 任意原型（客户、员工、公司、产品、账户、设备、车辆、服务器、社交账户、司机、地点，或您自定义的原型），具备属性、状态、目标、记忆、关系、资源、技能与风险画像。
- **行为引擎：** 基于规则的条件判断、状态机，以及效用型 AI 目标评分——实体是在“决策”，而不仅仅是回放脚本。
- **人口生成器：** 支持正态分布、均匀分布、高斯混合分布、幂律分布、Zipf 分布、泊松分布及自定义分布，并支持跨属性相关性、噪声、缺失数据与异常值控制，以生成真实的合成数据集。
- **事件引擎：** 一个按优先级、按 Tick 排序的事件总线，支持依赖门控、背压控制与死信队列——世界中的每一次变化都是一个具有触发条件、条件判断、载荷、来源/目标及后续影响的结构化事件。
- **分布式模拟协调器：** 将实体分区到各个工作节点，跟踪心跳与健康状态，并自动将失效工作节点的分区重新分配给其他节点。
- **场景构建器 + 模板库：** 内置 19 个场景模板（用户激增、服务遭受攻击、公司破产、病毒式增长、数据库崩溃、网络中断、抢购潮、欺诈激增、资源短缺、客户行为转变），以及面向不同垂直领域的启动模板（电商、银行、SaaS、社交网络、物联网、物流、网络靶场、智慧城市、云基础设施）。
- **混沌引擎：** 支持网络延迟、丢包、服务/数据库/工作节点故障、资源枯竭及依赖失败等故障注入，并提供影响报告。
- **负载测试生成器：** 将模拟出的行为转化为对目标系统的真实 HTTP/TCP/WebSocket 流量，支持爬升/回落、并发控制与速率限制（gRPC/消息队列适配器可接入同一接口）。
- **垂直领域模拟器：** 社交网络（关注图谱、病毒式传播、社区发现）、商业（公司/市场份额）、金融（支付/转账/欺诈）、网络拓扑（具备依赖感知的级联故障）、以及物联网/智慧城市（传感器、交通信号灯、能源）。
- **状态管理：** 支持保存/恢复/克隆/分叉/差异对比/回滚的快照系统，并具备完整版本管理；重放引擎可基于种子实现精确复现。
- **遥测：** 一个兼容 Prometheus 的 `/metrics` 接口，以及供仪表盘使用的 JSON 快照——包含实体数量、Tick 速率、队列深度、死信数量、堆内存及协程数。
- **调度器与完全由用户定义的运营时间：** 对长时间运行的任务进行排队与重试，并且——完全通过界面配置、不含任何硬编码默认值——设置系统在每周的哪些天、哪些具体时段被视为“开放”，并实时倒计时显示距下一次开/关状态切换的时间。
- **控制中心仪表盘：** 世界、实体、场景、模拟、工作节点、指标与事件浏览器，通过 WebSocket 实现实时更新，支持**英语、波斯语（RTL）与中文**三种语言，并提供**浅色、深色、Windows 默认、红色与蓝色**五种主题。
- **数据导出：** 支持 JSON、NDJSON、CSV，以及兼容 Parquet 的列式格式，便于对接下游大数据处理流水线。
- **命令行工具：** `serve`、`create-world`、`run` / `run-headless`、`generate-dataset`、`benchmark`。

### 项目结构

```
void/
├── backend/                  # Go 模拟核心 + API
│   ├── cmd/void/              # 命令行入口
│   └── internal/
│       ├── entity/            # 通用实体模型 + 原型
│       ├── behavior/           # 规则引擎、状态机、效用型 AI
│       ├── population/         # 统计分布 + 生成器
│       ├── event/               # 优先级事件总线
│       ├── world/                # 世界构建器
│       ├── simulation/            # 并发 Tick 引擎
│       ├── coordinator/            # 分布式分区/健康检查
│       ├── scenario/                # 场景构建器 + 模板库
│       ├── chaos/                    # 故障注入引擎
│       ├── loadtest/                  # HTTP/TCP/WS 负载适配器
│       ├── social/ business/           # 垂直领域模拟器
│       ├── finance/ network/ iot/
│       ├── snapshot/ replay/ export/    # 状态管理 + I/O
│       ├── scheduler/                    # 任务队列 + 运营时间
│       ├── storage/                       # 可对接 Postgres/Redis/NATS/S3 的数据访问层
│       ├── metrics/                        # 兼容 Prometheus 的遥测
│       └── api/                             # REST + WebSocket 服务器
├── frontend/                 # Next.js + TypeScript 控制中心
│   ├── app/                   # 页面（仪表盘、世界、实体……）
│   ├── components/             # 侧边栏、主题/语言切换器、图表
│   └── lib/                     # i18n 语言包 + 类型化 API 客户端
├── deploy/helm/               # Kubernetes Helm Chart
├── docker-compose.yml
└── Makefile
```

### 快速开始

**前置要求：** Go ≥ 1.22，Node.js ≥ 20，（可选）Docker 与 Docker Compose。

**1. 运行后端 API：**

```bash
cd backend
go build -o void ./cmd/void
./void serve
# API 监听在 :8080 端口 —— 可执行 curl localhost:8080/healthz 测试
```

**2. 运行仪表盘：**

```bash
cd frontend
npm install
npm run dev
# 仪表盘运行在 :3000 端口，通过 NEXT_PUBLIC_API_BASE（默认 :8080）与 API 通信
```

**3. 或使用 Docker Compose 一键运行全部服务：**

```bash
docker compose up --build
# API      → http://localhost:8080
# 仪表盘 → http://localhost:3000
```

**4. 或完全通过命令行工具驱动（无需仪表盘）：**

```bash
cd backend
go run ./cmd/void create-world --seed 42 --size medium
go run ./cmd/void run --agents 100000 --ticks 500 --workers 8
go run ./cmd/void generate-dataset --agents 50000 --out dataset.csv --format csv
go run ./cmd/void benchmark --agents 1000000 --ticks 50 --workers 16
```

### 配置

所有配置均通过环境变量进行（参见 `.env.example`）：

| 变量 | 用途 | 默认值 |
|---|---|---|
| `VOID_HTTP_ADDR` | API 绑定地址 | `:8080` |
| `VOID_DATA_DIR` | 本地数据/导出/快照目录 | `./data` |
| `VOID_WORKER_COUNT` | 默认模拟工作池大小 | `8` |
| `VOID_JWT_SECRET` | 身份验证签名密钥（生产环境请设置真实值） | — |
| `VOID_POSTGRES_DSN` / `VOID_REDIS_ADDR` / `VOID_NATS_URL` / `VOID_S3_ENDPOINT` | 存储接口背后的可选持久化后端 | 留空 = 使用内存/文件 |
| `NEXT_PUBLIC_API_BASE` | 仪表盘所调用的 API 地址 | `http://localhost:8080` |

### 运营时间（完全由用户配置）

在仪表盘的**设置 → 运营时间**中（或通过 `PUT /api/v1/scheduler/operating-hours` 接口），您可以——在没有任何预设值的情况下——精确定义系统在任意 IANA 时区下、每周的哪些天、哪些具体时间应被视为“开放”以供计划任务运行。该面板会实时显示开放/关闭状态，并根据您所输入的内容计算距下一次状态切换的倒计时。

### API

服务器运行后，完整的 REST + WebSocket 接口文档可通过 `GET /openapi.json` 查看。主要接口包括：`POST /api/v1/worlds`、`POST /api/v1/simulations`、`POST /api/v1/simulations/{id}/scenarios/{scenarioId}/inject`、`POST /api/v1/chaos/faults`、`GET /metrics`（Prometheus 格式）、`GET /ws/simulations/{id}`（实时 Tick 流）。

### 许可

本项目按“原样”提供，供评估与进一步开发使用。在进行任何生产环境或公开部署之前，请审查并调整安全机制、身份验证方式及存储后端。
