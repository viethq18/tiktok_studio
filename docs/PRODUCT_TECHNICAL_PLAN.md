TikTok Carousel AI Platform
Product & Technical Implementation Blueprint
Version: 1.0
Status: Production MVP
Primary Platform: TikTok Carousel
Future Platform: Instagram Carousel
Frontend: Next.js + React + TypeScript + Tailwind CSS + shadcn/ui
Canvas Engine: Fabric.js
Backend: Golang
Database: PostgreSQL
Object Storage: MinIO
Queue: Redis
Authentication: Google OAuth + Development Email Login
AI: OpenAI-compatible API
Image Provider: Unsplash
1. Product Vision
1.1 Problem
Một cá nhân muốn xây dựng TikTok channel thường phải thực hiện rất nhiều công việc:
Tìm topic.
Research thông tin.
Xác định content angle.
Viết hook.
Viết nội dung từng slide.
Chọn CTA.
Tìm hình ảnh.
Nghĩ layout.
Chọn font.
Thiết kế từng slide.
Export.
Lặp lại quy trình cho carousel tiếp theo.
Điều này khiến việc duy trì một content channel trở nên tốn thời gian.
1.2 Product solution
Sản phẩm biến quy trình trên thành:
Niche
  ↓
AI hiểu channel
  ↓
AI tạo content configuration
  ↓
User input một vài thông tin
  ↓
AI research
  ↓
AI viết content
  ↓
AI quyết định carousel formula
  ↓
AI quyết định layout + typography
  ↓
AI tìm ảnh
  ↓
Carousel được generate
  ↓
User chỉnh rất ít nếu cần
  ↓
Export ZIP

Mục tiêu không phải tạo ra một Canva clone.
Mục tiêu là:

AI tạo ra 80–95% carousel hoàn chỉnh, user chỉ cần review và chỉnh một vài chi tiết nếu muốn.
2. Product Principles
2.1 AI-first
User không nên phải suy nghĩ quá nhiều về:
slide structure
content formula
font
typography
spacing
image keyword
layout
AI chịu trách nhiệm chính.
2.2 User remains in control
AI generate nhưng user luôn có thể:
edit text
move text
resize text
thay background
chọn ảnh khác
undo/redo
regenerate
2.3 Dynamic by project
Không hard-code một form chung cho tất cả niche.
Ví dụ:

Project A
Baby Care
→ Baby age
→ Parenting problem
→ Advice type
→ CTA

Trong khi:
Project B
Personal Finance
→ Investment topic
→ Risk level
→ Investment horizon
→ CTA

sẽ có schema khác.
2.4 Structured AI
Không để AI trả output tự do.
Mọi AI output quan trọng phải:

Generate
→ Parse
→ JSON Schema validation
→ Business validation
→ Persist

2.5 Design JSON as source of truth
Không lưu HTML làm design source.
Không để canvas state trở thành một black box.

Architecture:

Design JSON
      ↓
Fabric.js
      ↓
Visual Editor
      ↓
Design JSON

3. Target Users
MVP tập trung vào:
Individual TikTok creators
Không tập trung vào:
Enterprise
Agency
Large teams
Social media departments
Tuy nhiên hệ thống phải hỗ trợ:
1 User
 ├── Project A
 │    ├── Carousel 1
 │    ├── Carousel 2
 │    └── Carousel 3
 │
 └── Project B
      ├── Carousel 1
      └── Carousel 2

MVP chưa cần project sharing giữa users.
4. Core User Journey
4.1 First-time user
Landing
  ↓
Google Login
  ↓
Create Project
  ↓
Enter niche
  ↓
AI Research
  ↓
AI creates Project Intelligence
  ↓
User reviews
  ↓
Project created

Ví dụ:
"Tôi muốn xây dựng TikTok chia sẻ kiến thức chăm sóc trẻ sơ sinh 0–3 tuổi."
AI tạo:
Project Name
Baby Care 0–3

Target Audience
Parents with babies 0–3

Tone
Friendly
Trustworthy
Simple

Content Pillars
- Sleep
- Nutrition
- Development
- Common mistakes
- Parenting tips

CTA options
- Save
- Follow
- Share
- Comment

5. Project Concept
Project không chỉ là folder.
Project là:

AI Content Brain của một TikTok channel.
Project lưu context lâu dài để các carousel sau không cần user nhập lại.
5.1 Project structure
Project
├── Identity
│   ├── name
│   ├── niche
│   ├── language
│   └── description
│
├── Audience
│   ├── target audience
│   ├── demographics
│   └── pain points
│
├── Content Strategy
│   ├── content pillars
│   ├── tone
│   ├── writing style
│   └── preferred angles
│
├── CTA
│   ├── save
│   ├── follow
│   ├── share
│   └── comment
│
├── Brand
│   ├── logo
│   ├── colors
│   └── fonts
│
└── Carousels

6. Onboarding
6.1 Screen
MVP onboarding chỉ cần một màn hình đơn giản.
┌──────────────────────────────────────────────┐
│                                              │
│        Build your TikTok content channel     │
│                                              │
│  What is your channel about?                 │
│                                              │
│  ┌────────────────────────────────────────┐  │
│  │ e.g. Baby care for 0–3 year olds...   │  │
│  └────────────────────────────────────────┘  │
│                                              │
│  Examples                                   │
│                                              │
│  • Baby care 0–3 years                       │
│  • Personal finance for beginners            │
│  • Healthy meals for busy people             │
│                                              │
│             [ Create Project ]               │
│                                              │
└──────────────────────────────────────────────┘

6.2 AI onboarding process
User Input
    ↓
Normalize niche
    ↓
Research
    ↓
Analyze audience
    ↓
Generate content strategy
    ↓
Generate CTA options
    ↓
Generate project configuration
    ↓
Validate
    ↓
Save Project

7. Project Intelligence Schema
Project intelligence nên có version.
Ví dụ:

{
  "schema_version": "1.0",
  "identity": {
    "niche": "Baby care 0-3",
    "language": "vi"
  },
  "audience": {
    "description": "Parents with babies 0-3 years old"
  },
  "tone": [
    "friendly",
    "trustworthy",
    "simple"
  ],
  "content_pillars": [
    {
      "id": "sleep",
      "name": "Baby sleep"
    },
    {
      "id": "nutrition",
      "name": "Nutrition"
    }
  ],
  "cta_options": [
    {
      "id": "save",
      "label": "Save this post"
    },
    {
      "id": "follow",
      "label": "Follow for more"
    }
  ]
}

Không nên để frontend phụ thuộc trực tiếp vào raw AI output.
Backend phải normalize output thành internal schema.

8. Create Carousel Flow
Project dashboard:
[ + Create Carousel ]

Click:
Create Carousel
       ↓
AI analyzes project
       ↓
Generate carousel input schema
       ↓
Render dynamic form

9. Dynamic Schema Architecture
9.1 Core principle
AI quyết định:
Field nào cần xuất hiện?
Frontend quyết định:
Field đó được render bằng component nào?
Không cho AI generate arbitrary UI.
Architecture:

AI
 ↓
JSON Schema
 ↓
Schema Validator
 ↓
Schema Renderer
 ↓
UI Components

9.2 Supported component registry
MVP:
text
textarea
number
select
multi_select
radio
checkbox
slider
image
url

V1:
color
font
date
rich_text
repeatable_group
conditional_group

9.3 Example schema
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "topic": {
      "type": "string",
      "title": "Topic"
    },
    "slide_count": {
      "type": "integer",
      "minimum": 3,
      "maximum": 7,
      "default": 5
    },
    "cta": {
      "type": "string",
      "title": "CTA",
      "enum": [
        "save",
        "follow",
        "share",
        "comment"
      ]
    }
  },
  "required": [
    "topic",
    "slide_count",
    "cta"
  ]
}

10. Dynamic Form UX
Left panel:
┌──────────────────────────────┐
│ Create Carousel              │
├──────────────────────────────┤
│                              │
│ Topic                        │
│ ┌──────────────────────────┐ │
│ │ 5 signs baby is tired    │ │
│ └──────────────────────────┘ │
│                              │
│ Number of slides             │
│                              │
│ 3 ─────────●─────── 7       │
│             5                │
│                              │
│ CTA                          │
│ ○ Save this post             │
│ ○ Follow for more            │
│ ● Share with another parent  │
│                              │
│ [ Generate Carousel ]        │
│                              │
└──────────────────────────────┘

11. Schema Safety
AI-generated schema phải được validate bằng:
JSON parsing.
JSON Schema validation.
Allowed component validation.
Field depth limit.
Number range validation.
Enum size limit.
String length limit.
Schema complexity limit.
Ví dụ không cho AI tạo:
{
  "component": "execute_javascript"
}

Chỉ allowed component registry mới được render.
12. Carousel Generation Pipeline
Đây là core backend workflow.
Create Carousel
      ↓
Create generation job
      ↓
Redis
      ↓
Worker
      ↓
Load Project Context
      ↓
Research
      ↓
Generate Content
      ↓
Validate Content
      ↓
Select Formula
      ↓
Generate Design Plan
      ↓
Validate Design JSON
      ↓
Search Unsplash
      ↓
Download Assets
      ↓
Save to MinIO
      ↓
Attach Assets
      ↓
Carousel READY

13. Async Job Architecture
Không giữ HTTP connection trong suốt quá trình generation.
API:

POST /api/v1/carousels

Backend:
create carousel
create generation_job
push Redis
return 202

Response:
{
  "carousel_id": "car_123",
  "job_id": "job_456",
  "status": "queued"
}

Frontend subscribe/poll:
GET /api/v1/jobs/job_456

V1 có thể polling.
V2 có thể dùng WebSocket/SSE.

14. Generation Job State Machine
QUEUED
  ↓
RESEARCHING
  ↓
GENERATING_CONTENT
  ↓
VALIDATING_CONTENT
  ↓
GENERATING_DESIGN
  ↓
SEARCHING_IMAGES
  ↓
DOWNLOADING_ASSETS
  ↓
FINALIZING
  ↓
COMPLETED

Failure:
ANY STATE
   ↓
FAILED
   ↓
RETRY

15. Generation Progress UI
Right side:
Generating your carousel...

✓ Researching topic
✓ Creating content
✓ Choosing carousel formula
● Designing slides
○ Finding images
○ Preparing final design

Mục tiêu UX:
User luôn biết AI đang làm gì.
16. Internet Research
Research được chia làm hai tầng.
16.1 Project research
Dùng khi tạo Project.
Mục tiêu:

understand niche
audience
content pillars
common topics
CTA patterns
16.2 Carousel research
Chạy mỗi lần generate carousel.
Mục tiêu:

fact checking
current information
topic context
trend discovery
content inspiration
Không copy content từ sources.
AI phải synthesize.

17. Research Sources
Research result nên lưu:
{
  "source_id": "src_123",
  "url": "...",
  "title": "...",
  "domain": "...",
  "retrieved_at": "..."
}

Carousel có:
Research Sources
├── Source 1
├── Source 2
└── Source 3

MVP có thể chưa expose đầy đủ UI, nhưng data nên được lưu.
18. Content Generation
Content generator nhận:
Project Context
+
Carousel Input
+
Research
+
CTA
+
Platform constraints

Output:
{
  "slides": [
    {
      "index": 1,
      "role": "hook",
      "headline": "...",
      "body": null
    },
    {
      "index": 2,
      "role": "point",
      "headline": "...",
      "body": "..."
    },
    {
      "index": 5,
      "role": "cta",
      "headline": "...",
      "body": "..."
    }
  ]
}

19. Carousel Formula
AI tự chọn formula.
MVP formula library:

Problem → Solution
Listicle
Myth → Fact
How-to
Mistake → Correction
Before → After
Question → Answer
Story → Lesson

Formula không hard-code vào UI.
Nên có internal entity:

carousel_formulas

AI chọn:
{
  "formula_id": "listicle",
  "reason": "Topic naturally fits a numbered list."
}

20. CTA
User chọn CTA preference.
AI vẫn chịu trách nhiệm đưa CTA vào vị trí phù hợp.

Ví dụ:

CTA:
Save this post

AI có thể generate:
Slide 5:
"Save this carousel so you can check these signs later."

CTA không nhất thiết phải copy literal user selection.
21. Design Generation
Sau khi content hoàn tất:
Content JSON
      ↓
Design Planner
      ↓
Design JSON

AI quyết định:
layout
text placement
hierarchy
font
font size
color
spacing
alignment
background style
image requirement
22. Design JSON
Ví dụ:
{
  "version": "1.0",
  "canvas": {
    "width": 1080,
    "height": 1350
  },
  "slides": [
    {
      "id": "slide_1",
      "elements": [
        {
          "id": "text_1",
          "type": "text",
          "content": "5 dấu hiệu bé đang thiếu ngủ",
          "x": 100,
          "y": 180,
          "width": 880,
          "height": 220,
          "style": {
            "fontFamily": "Inter",
            "fontSize": 72,
            "fontWeight": 700,
            "color": "#FFFFFF",
            "textAlign": "center"
          }
        },
        {
          "id": "background_1",
          "type": "image",
          "asset_id": null,
          "x": 0,
          "y": 0,
          "width": 1080,
          "height": 1350,
          "opacity": 1
        }
      ]
    }
  ]
}

23. Design JSON Rules
AI chỉ được sử dụng:
Allowed element types:
- text
- image
- shape

MVP editor chỉ expose:
text
image

Shape có thể được dùng bởi AI nhưng chưa nhất thiết cho user edit trực tiếp.
24. Canvas Ratios
MVP support:
Ratio	Resolution	Use case
4:5	1080×1350	Recommended
1:1	1080×1080	Square
9:16	1080×1920	Full-screen

Default:
4:5

AI phải biết canvas dimensions trước khi generate design.
25. Recommended Ratio UX
Khi tạo carousel:
Canvas format

○ 4:5
  1080 × 1350
  Recommended

○ 1:1
  1080 × 1080

○ 9:16
  1080 × 1920

26. Canvas Architecture
Sử dụng Fabric.js.
Không sử dụng DOM/CSS làm primary editor.

Architecture:

Design JSON
     ↓
Fabric Adapter
     ↓
Fabric Canvas
     ↓
User Interaction
     ↓
Fabric State
     ↓
Serialize
     ↓
Design JSON

27. Why Fabric.js
Fabric.js phù hợp vì:
object-based canvas
text objects
image objects
drag
resize
selection
serialization
layer ordering
export
transformation
Quan trọng nhất:
Design JSON có thể map khá tự nhiên sang Fabric objects.
28. Editor Layout
┌────────────────────────────────────────────────────────────┐
│ Logo       Project / Carousel          Undo Redo    Export │
├────────────────┬───────────────────────────────┬───────────┤
│                │                               │           │
│ LEFT PANEL     │       CANVAS EDITOR           │ SLIDES    │
│                │                               │           │
│ Content        │       ┌───────────────┐       │ ┌──────┐  │
│                │       │               │       │ │  01  │  │
│ Text           │       │    Canvas     │       │ └──────┘  │
│                │       │               │       │           │
│ Image          │       │               │       │ ┌──────┐  │
│                │       └───────────────┘       │ │  02  │  │
│                │                               │ └──────┘  │
└────────────────┴───────────────────────────────┴───────────┘

29. Editor MVP Features
Text
Select
Move
Resize
Edit content
Image
Select
Replace
Move
Resize/crop
Set background
Canvas
Zoom
Pan
Slide switching
History
Undo
Redo
Save
Autosave
30. Text Editing UX
Double click text:
┌───────────────────────┐
│ 5 dấu hiệu bé đang    │
│ thiếu ngủ             │
└───────────────────────┘

User có thể sửa trực tiếp.
MVP chưa cần:

complex rich text
per-character styling
advanced effects
animation
31. Undo / Redo
Implement command/state history.
Không chỉ rely vào browser undo.

History:

State 0
 ↓
Move text
 ↓
State 1
 ↓
Edit text
 ↓
State 2
 ↓
Resize
 ↓
State 3

Undo:
State 3
 ↓
State 2

32. Autosave
Editor phải autosave.
Ví dụ:

User edit
 ↓
Debounce 1–2 seconds
 ↓
Save Design JSON

UI:
Saving...
Saved ✓

Không save mỗi mousemove trực tiếp.
Có thể batch/throttle updates.

33. Unsplash Image Pipeline
Mỗi slide:
AI determines image intent
        ↓
Generate search keywords
        ↓
Unsplash search
        ↓
3 candidate images
        ↓
User selects

Ví dụ:
Slide 1

[ image 1 ] [ image 2 ] [ image 3 ]

              [ Research more ]

34. Image Search API
Backend không nên để browser gọi trực tiếp Unsplash.
Architecture:

Frontend
   ↓
Go API
   ↓
Unsplash

Lợi ích:
hide API configuration
rate limiting
logging
caching
future provider abstraction
35. Image Provider Abstraction
Không hard-code Unsplash vào domain logic.
ImageProvider
   │
   └── UnsplashProvider

Future:
ImageProvider
├── Unsplash
├── Pexels
└── OtherProvider

36. Selected Image Storage
Khi user chọn ảnh:
Unsplash
   ↓
Download
   ↓
MinIO
   ↓
Asset
   ↓
Carousel

Không phụ thuộc lâu dài vào external URL.
37. MinIO Structure
Object key đề xuất:
users/{user_id}/
  projects/{project_id}/
    carousels/{carousel_id}/
      assets/
        {asset_id}.jpg
      exports/
        {export_id}.zip

38. Local Upload
MVP hỗ trợ:
Image
├── Unsplash
└── Upload

Upload:
Browser
 ↓
Go API
 ↓
MinIO
 ↓
Asset record

Validate:
MIME type
file size
dimensions
extension
Không tin extension do client gửi.
39. Brand Kit
MVP foundation.
Project:

Brand
├── Logo
├── Primary Color
├── Secondary Color
├── Accent Color
└── Font Preferences

AI phải respect Brand Kit khi generate Design JSON.
40. Font System
Không cho AI chọn arbitrary font.
Backend duy trì:

font_registry

Ví dụ:
Inter
Roboto
Montserrat
Poppins
Playfair Display

AI chỉ được select font từ registry.
41. Project Dashboard
┌────────────────────────────────────────────────────┐
│ Baby Care 0–3                         + Carousel    │
├────────────────────────────────────────────────────┤
│ Search carousels...                                │
│                                                    │
│ All   Draft   Ready   Archived                     │
│                                                    │
│ ┌───────────┐ ┌───────────┐ ┌───────────┐         │
│ │ thumbnail │ │ thumbnail │ │ thumbnail │         │
│ │           │ │           │ │           │         │
│ └───────────┘ └───────────┘ └───────────┘         │
│                                                    │
│ 5 Signs Baby...                                    │
│ 7 Foods...                                         │
│ Baby Sleep...                                      │
└────────────────────────────────────────────────────┘

42. Carousel CRUD
Required:
Create
Read
Update
Delete
Duplicate
Archive

Delete nên soft-delete.
43. Carousel Status
DRAFT
GENERATING
READY
ARCHIVED
FAILED

Không dùng PUBLISHED trong MVP vì platform chưa publish trực tiếp lên TikTok.
44. Carousel Versioning
MVP foundation nên lưu:
carousel_versions

Mỗi major generation/edit có snapshot.
Tuy nhiên UI restore version có thể đưa sang V1.

45. Regeneration
MVP nên support:
Regenerate Carousel

V1:
Regenerate Slide

V1 advanced:
Rewrite text
Change tone
Make shorter
Make more engaging
Change image

46. Locked Elements
V1 nên support:
🔒 Lock background
🔒 Lock logo
🔓 Unlock text

AI regeneration không được overwrite locked elements.
47. Export Pipeline
User click:
Export

Backend:
Load Design JSON
 ↓
Render slides
 ↓
Generate PNG
 ↓
Create ZIP
 ↓
Store in MinIO
 ↓
Return download URL

48. Rendering Architecture
Nên tách:
Editor Renderer

và:
Export Renderer

Editor:
Fabric.js

Export:
Server-side renderer

Mục tiêu:
Export không phụ thuộc vào browser đang mở.
49. Export Output
MVP:
PNG
ZIP

Ví dụ:
carousel.zip

01.png
02.png
03.png
04.png
05.png

V1:
JPG
PDF

50. Future Instagram Architecture
Không hard-code:
TikTokCarousel

vào mọi domain object.
Nên có:

Platform
├── TikTok
└── Instagram

và:
CanvasPreset
├── TikTok 4:5
├── TikTok 1:1
├── TikTok 9:16
└── Instagram 4:5

Design JSON platform-agnostic.
51. Backend Architecture
Đề xuất Go modular monolith.
Không cần microservices ở MVP.

Go Application
│
├── Auth
├── Users
├── Projects
├── Schemas
├── Carousels
├── AI
├── Research
├── Images
├── Assets
├── Design
├── Export
└── Jobs

52. Suggested Go Structure
backend/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── auth/
│   ├── user/
│   ├── project/
│   ├── carousel/
│   ├── schema/
│   ├── ai/
│   ├── research/
│   ├── image/
│   ├── asset/
│   ├── design/
│   ├── export/
│   └── job/
│
├── pkg/
│   ├── logger/
│   ├── storage/
│   └── jsonschema/
│
├── migrations/
│
└── configs/

53. Recommended Go Architecture
Dùng layered architecture:
HTTP Handler
     ↓
Application Service
     ↓
Domain
     ↓
Repository
     ↓
PostgreSQL

Ví dụ:
POST /carousels
       ↓
CarouselHandler
       ↓
CreateCarouselService
       ↓
CarouselRepository

Không để business logic trong HTTP handler.
54. AI Architecture
AIService
│
├── ProjectAnalyzer
├── SchemaGenerator
├── ContentResearcher
├── ContentGenerator
├── FormulaSelector
├── DesignGenerator
└── ImageIntentGenerator

Provider:
AIProvider
   ↓
OpenAICompatibleProvider

Configuration:
AI_BASE_URL
AI_API_KEY
AI_MODEL

55. AI Retry Strategy
AI calls phải có:
timeout
retry
backoff
schema validation
max token limits

Nếu JSON invalid:
AI response
 ↓
Validation failed
 ↓
Repair prompt
 ↓
Retry

Không retry vô hạn.
Ví dụ:

max_attempts = 3

56. AI Prompt Architecture
Không hard-code prompt trong handler.
Ví dụ:

internal/ai/prompts/

project_analysis
project_schema
carousel_schema
research
content_generation
formula_selection
design_generation

Prompt nên versioned:
prompt_name
version
template

Ví dụ:
design_generation:v3

57. AI Output Contracts
Mỗi AI task có contract riêng.
Ví dụ:

ProjectAnalysisOutput
CarouselSchemaOutput
ResearchOutput
ContentOutput
DesignOutput

Không dùng một generic:
map[string]interface{}

cho toàn bộ AI system.
Dynamic schema là ngoại lệ có validation riêng.

58. PostgreSQL Core Tables
Đề xuất:
users
projects
project_ai_contexts
project_schemas
carousels
carousel_inputs
carousel_content
carousel_designs
carousel_versions
assets
carousel_assets
research_sources
generation_jobs
ai_generations
image_searches
export_jobs
exports
font_registry
formula_registry

59. Users
users
------
id
email
name
avatar_url
provider
provider_user_id
created_at
updated_at

Unique:
provider + provider_user_id
email

60. Projects
projects
--------
id
user_id
name
niche
description
language
status
created_at
updated_at
deleted_at

Index:
user_id
created_at
deleted_at

61. Project AI Context
project_ai_contexts
-------------------
id
project_id
version
context_json
created_at
updated_at

context_json contains:
audience
tone
content pillars
CTA
preferences
strategy
62. Project Schema
project_schemas
---------------
id
project_id
version
schema_json
created_at

Schema có version để AI có thể thay đổi schema mà không phá các carousel cũ.
63. Carousels
carousels
---------
id
project_id
title
status
platform
canvas_ratio
canvas_width
canvas_height
created_at
updated_at
deleted_at

64. Carousel Inputs
carousel_inputs
---------------
id
carousel_id
schema_version
input_json
created_at
updated_at

Không cần tạo database column cho từng dynamic field.
Ví dụ:

{
  "topic": "...",
  "slide_count": 5,
  "cta": "save"
}

65. Carousel Content
carousel_content
----------------
id
carousel_id
version
content_json
created_at

66. Carousel Design
carousel_designs
----------------
id
carousel_id
version
design_json
created_at
updated_at

Latest version được đánh dấu hoặc lấy theo max version.
67. Assets
assets
------
id
user_id
project_id
carousel_id
source
source_id
storage_key
mime_type
width
height
file_size
metadata_json
created_at

Source:
unsplash
upload
generated

68. Research Sources
research_sources
----------------
id
project_id
carousel_id
url
title
domain
source_type
metadata_json
created_at

69. Generation Jobs
generation_jobs
---------------
id
user_id
project_id
carousel_id
type
status
progress
current_step
attempt
error_code
error_message
created_at
started_at
completed_at

70. AI Generations
Audit AI calls:
ai_generations
--------------
id
job_id
provider
model
task_type
prompt_version
input_json
output_json
token_usage
latency_ms
status
created_at

Không nhất thiết lưu full prompt nếu chứa sensitive information.
71. Image Searches
image_searches
--------------
id
carousel_id
slide_id
query
provider
results_json
created_at

Có thể cache image search.
72. Export Jobs
export_jobs
-----------
id
carousel_id
format
status
progress
created_at
completed_at

73. API Structure
Base:
/api/v1

Auth:
GET  /auth/google
GET  /auth/callback
POST /auth/logout
GET  /auth/me

Projects:
GET    /projects
POST   /projects
GET    /projects/:id
PATCH  /projects/:id
DELETE /projects/:id

Carousels:
GET    /projects/:id/carousels
POST   /projects/:id/carousels
GET    /carousels/:id
PATCH  /carousels/:id
DELETE /carousels/:id
POST   /carousels/:id/duplicate
POST   /carousels/:id/archive

74. Dynamic Schema API
POST /projects/:id/generate-schema
GET  /projects/:id/schema

Carousel schema:
POST /projects/:id/carousel-schema

Có thể cache schema theo project context version.
75. Generation API
POST /projects/:id/carousels/generate

Body:
{
  "inputs": {
    "topic": "5 signs baby is tired",
    "slide_count": 5,
    "cta": "save"
  },
  "ratio": "4:5"
}

Response:
{
  "carousel_id": "car_123",
  "job_id": "job_123",
  "status": "queued"
}

76. Job API
GET /jobs/:id

Response:
{
  "id": "job_123",
  "status": "generating_design",
  "progress": 72,
  "current_step": "Finding images"
}

77. Image API
GET /carousels/:id/slides/:slideId/images
POST /carousels/:id/slides/:slideId/images/search
POST /carousels/:id/slides/:slideId/images/select

78. Asset API
POST /assets/upload
GET  /assets/:id
DELETE /assets/:id

Upload dùng multipart/form-data.
79. Design API
GET   /carousels/:id/design
PATCH /carousels/:id/design
POST  /carousels/:id/design/versions

Autosave:
PATCH /carousels/:id/design

Body:
{
  "version": 14,
  "design": {}
}

Backend kiểm tra optimistic concurrency.
80. Optimistic Concurrency
Nếu browser A đang version 14 nhưng server đã version 15:
PATCH version=14

Server:
409 Conflict

Frontend phải reload/merge.
Điều này quan trọng khi user mở nhiều tab.

81. Redis
Redis dùng cho:
generation jobs
export jobs
retry queue
rate limiting
temporary progress
caching

MVP có thể dùng Redis Streams hoặc một job library ổn định.
Không cần tự xây queue protocol phức tạp nếu không cần.

82. Worker
Nên tách process:
go-api
go-worker

API:
HTTP

Worker:
Redis
AI
Unsplash
MinIO
Export

Có thể deploy cùng image nhưng chạy command khác.
83. Transaction Boundaries
Ví dụ tạo carousel:
BEGIN
 create carousel
 create generation_job
COMMIT

enqueue job

Không giữ DB transaction trong suốt AI generation.
84. Security
MVP cần:
OAuth validation
secure cookies
CSRF protection nếu cookie auth
rate limiting
input validation
file validation
ownership checks
SQL parameterization
signed asset URLs
no arbitrary URL fetching
Đặc biệt:
User A không bao giờ được truy cập project/asset của User B chỉ bằng cách thay ID.
85. Object Storage Security
MinIO bucket nên private.
Browser nhận:

signed URL

thay vì public bucket.
86. API Authorization
Mỗi request protected phải check:
authenticated user
      ↓
resource ownership

Ví dụ:
GET /projects/project_123

Backend:
project.user_id == current_user.id

87. Rate Limiting
Các endpoint cần rate limit cao:
AI generation
Image search
Export
Upload

Không để user spam AI API.
88. Cost Control
AI generation là cost center.
Mỗi generation nên lưu:

model
input tokens
output tokens
estimated cost

Có thể thêm:
user_usage

sau này.
89. Usage Model
MVP có thể chưa billing.
Nhưng architecture nên chuẩn bị:

usage_events

Ví dụ:
carousel_generation
image_search
export
research

90. Error Handling
User-facing error:
We couldn't generate your carousel.

[ Try again ]

Không expose:
OpenAI timeout
Redis error
Postgres constraint

Backend log chi tiết.
91. Retry Categories
Retry được:
AI timeout
AI invalid JSON
Unsplash temporary error
MinIO temporary error
network error

Không retry:
invalid user input
permission denied
unsupported file
schema business validation failure

92. Observability
MVP nên có:
structured logs
request ID
job ID
AI generation ID
error tracking
latency
job duration
Mỗi request:
request_id

Mỗi generation:
generation_job_id

Nhờ vậy debug được:
User
→ Request
→ Job
→ AI call
→ Image search
→ Export

93. Frontend Architecture
frontend/
├── app/
├── components/
│   ├── ui/
│   ├── project/
│   ├── carousel/
│   ├── schema-form/
│   ├── editor/
│   └── generation/
│
├── features/
│   ├── auth/
│   ├── projects/
│   ├── carousels/
│   ├── editor/
│   └── assets/
│
├── lib/
│   ├── api/
│   ├── fabric/
│   ├── schema/
│   └── utils/
│
└── types/

94. State Management
Tách:
Server state
Dùng React Query/TanStack Query.
Ví dụ:

projects
carousels
jobs
assets

Local editor state
Dùng Zustand hoặc local state.
Editor:

selected object
active slide
zoom
history
dirty state

Không đưa toàn bộ Fabric state vào global server state.
95. Schema Renderer
Core:
<DynamicForm schema={schema} />

Internally:
SchemaRenderer
├── TextField
├── TextareaField
├── SelectField
├── MultiSelectField
├── RadioField
├── CheckboxField
├── SliderField
└── ImageField

96. Schema Validation Frontend
Frontend validate để UX tốt.
Backend vẫn validate lại.

Không bao giờ tin frontend validation.

97. Editor Component Structure
<CarouselEditor>
  <EditorToolbar />
  <EditorSidebar />
  <CanvasWorkspace />
  <SlideNavigator />
  <PropertiesPanel />
</CarouselEditor>

MVP:
EditorSidebar
CanvasWorkspace
SlideNavigator

Properties panel chỉ expose basic text editing.
98. Fabric Adapter
Không để business logic trực tiếp phụ thuộc Fabric everywhere.
Tạo:

FabricCanvasAdapter

Methods:
loadDesign()
serializeDesign()
addText()
updateText()
addImage()
updateImage()
removeObject()
undo()
redo()

Như vậy sau này có thể thay renderer nếu cần.
99. Canvas Coordinate System
Luôn lưu coordinates theo logical canvas size:
1080 × 1350

Không lưu theo viewport pixels.
Browser zoom chỉ là rendering concern.

Ví dụ:

x = 100
y = 200

luôn có nghĩa:
100 / 200 trên 1080×1350 canvas

100. Responsive Editor
Canvas preview có thể scale:
Actual canvas:
1080×1350

Browser:
540×675

Coordinates vẫn:
1080×1350

Fabric viewport transform xử lý scale.
101. Project Navigation
Sidebar:
Projects
  ↓
Current Project
  ├── Carousels
  ├── Brand
  └── Settings

MVP chưa cần quá nhiều navigation.
102. Important UX: Don't Make AI Feel Slow
AI workflow nên progressive.
Không:

Generating...

trong 60 giây.
Nên:

Researching your topic...
✓

Writing your carousel...
✓

Designing your slides...
●

Finding images...
○

103. Important UX: Don't Expose Complexity
User không cần biết:
JSON Schema
Design JSON
Fabric.js
Redis
AI provider

Đây là internal implementation.
User experience nên nói:

Tell us about your content

không phải:
Configure schema

104. MVP Screens
MVP cần khoảng các screen sau:
1. Login
2. Onboarding
3. Project Dashboard
4. Project Settings
5. Create Carousel
6. Generation Progress
7. Carousel Editor
8. Export

105. Login Screen
Continue with Google

Development:
Dev Login
Email

Chỉ render Dev Login khi:
ALLOW_DEV_LOGIN=true

106. Project Dashboard
Actions:
Create Project
Open Project
Rename
Delete

Project card:
Name
Niche
Number of carousels
Last edited

107. Project Settings
Tabs:
General
AI Strategy
Brand

AI Strategy:
Audience
Tone
Content pillars
CTA preferences

User có thể chỉnh AI-generated settings.
108. Create Carousel Screen
Layout:
┌─────────────────┬────────────────────────────┐
│ Input           │ Preview / Generation      │
│                 │                            │
│ Dynamic Form    │ Carousel Preview           │
│                 │                            │
│                 │                            │
└─────────────────┴────────────────────────────┘

109. Generation Result
Sau khi AI hoàn tất:
┌─────────────────┬────────────────────────────┐
│ Inputs          │ Canvas                     │
│                 │                            │
│ Topic           │     Slide 1                │
│ CTA             │                            │
│ Slide count     │                            │
│                 │                            │
│ [Regenerate]    │                            │
└─────────────────┴────────────────────────────┘

User có thể chuyển trực tiếp sang editor.
110. MVP Acceptance Criteria
Project creation
User đăng nhập Google.
User nhập niche.
AI tạo project intelligence.
Project được lưu.
User có thể chỉnh settings.
Dynamic schema
AI tạo valid JSON Schema.
Backend validate schema.
Frontend render dynamic form.
Không cần hard-code field theo niche.
Generation
User submit form.
Job được tạo.
Worker xử lý async.
UI hiển thị progress.
Carousel được tạo.
Design
Mỗi slide có Design JSON.
Design JSON render thành Fabric canvas.
Text editable.
Text movable.
Text resizable.
Undo/redo hoạt động.
Images
Mỗi slide có 3 Unsplash candidates.
User chọn image.
Image được lưu vào MinIO.
Background được cập nhật.
Export
User click Export.
Backend render.
ZIP được tạo.
User download được.
111. P0 — MVP Features
AUTH
✓ Google login
✓ Dev login

PROJECT
✓ CRUD
✓ AI project analysis
✓ Project context
✓ Project settings
✓ Brand kit foundation

CAROUSEL
✓ CRUD
✓ Duplicate
✓ Archive
✓ Dynamic schema
✓ Dynamic form

AI
✓ Research
✓ Content generation
✓ Formula selection
✓ Design generation
✓ JSON validation
✓ Retry

IMAGE
✓ Unsplash
✓ 3 candidates / slide
✓ Research more
✓ Local upload
✓ MinIO

EDITOR
✓ Fabric.js
✓ Text edit
✓ Text move
✓ Text resize
✓ Image background
✓ Undo
✓ Redo
✓ Autosave

EXPORT
✓ PNG
✓ ZIP

INFRA
✓ PostgreSQL
✓ Redis
✓ MinIO
✓ Go worker

112. P1 — Immediately After MVP
Regenerate individual slide
Rewrite text
AI alternatives
Lock elements
Advanced typography
Better image search
JPG export
PDF export
Version history UI
Brand templates
More formulas
Content calendar
Usage limits

113. P2 — Product Expansion
Instagram
TikTok account analytics
Competitor analysis
Content calendar
Auto publishing
Team collaboration
Workspace
Roles
Billing
Admin dashboard
Analytics dashboard
Template marketplace
AI image generation
Video carousel

114. Instagram Expansion Strategy
Không tạo một editor mới.
Reuse:

Design JSON
Canvas Renderer
Asset System
Export Pipeline

Chỉ thêm:
PlatformPreset
InstagramPreset

Ví dụ:
Platform
├── TikTok
│   ├── 4:5
│   ├── 1:1
│   └── 9:16
│
└── Instagram
    ├── 4:5
    ├── 1:1
    └── ...

115. Analytics Architecture
Chưa cần UI nhưng nên chuẩn bị event tracking.
Events:

user_signed_up
project_created
project_updated
carousel_created
carousel_generation_started
carousel_generation_completed
carousel_generation_failed
image_search_started
image_selected
carousel_edited
carousel_exported

Event structure:
{
  "event": "carousel_generation_completed",
  "user_id": "...",
  "project_id": "...",
  "carousel_id": "...",
  "metadata": {}
}

116. Admin Architecture
Chưa cần UI MVP.
Nhưng domain nên chuẩn bị:

Admin
├── Users
├── Projects
├── Carousels
├── AI Jobs
├── AI Generations
├── Prompts
├── Formulas
├── Fonts
└── Usage

117. Prompt Management
Một trong những thứ nên thiết kế sớm.
Không hard-code prompt trong source code.

Có thể bắt đầu bằng DB/file versioned:

prompt:
project_analysis
version:
1

Sau này admin có thể thay prompt mà không deploy backend.
118. Formula Management
Internal registry:
id
name
description
structure
rules
recommended_for

Ví dụ:
{
  "id": "listicle",
  "name": "Listicle",
  "structure": [
    "hook",
    "item",
    "item",
    "item",
    "cta"
  ]
}

119. Content Quality Layer
Một AI generation không nên coi là thành công chỉ vì JSON valid.
Cần hai loại validation:

Structural validation
JSON valid
Schema valid
Required fields

Semantic validation
Hook exists
CTA exists
Slide count correct
No duplicate slides
No excessively long text
No unsupported claims

120. Design Quality Validation
Backend kiểm tra:
text outside canvas
text width <= canvas width
font exists
font size reasonable
color valid
image exists
all slide IDs unique

Có thể thêm AI visual critique sau này.
121. Text Overflow
Đây là một vấn đề cực kỳ quan trọng.
AI có thể generate:

1000 characters

nhưng design chỉ chứa được:
300 characters

MVP cần:
max text length

và:
overflow detection

Nếu overflow:
Design validation failed
 ↓
AI rewrite shorter

122. AI Design Constraints
Prompt phải cung cấp:
Canvas size
Safe margins
Max characters
Allowed fonts
Allowed colors
Allowed elements

Ví dụ:
Safe margin:
80px

Maximum headline:
90 chars

Maximum body:
220 chars

123. Safe Zone
Mỗi canvas nên có:
safe_area

Ví dụ:
x: 80
y: 80
width: 920
height: 1190

AI không đặt text ngoài safe area.
124. Typography Hierarchy
AI output nên support roles:
hook
headline
subheadline
body
caption
cta

Mỗi role có default style.
Ví dụ:

hook → 72px
headline → 64px
body → 36px
caption → 24px

AI có thể override trong allowed range.
125. Color System
AI không nên random hex toàn bộ.
Project có:

brand colors

Design generator sử dụng:
primary
secondary
accent
background
text

Nếu chưa có brand:
AI generates temporary palette

Palette được lưu vào carousel design.
126. Design Templates
MVP không cần user-facing template marketplace.
Nhưng nên có internal templates:

Template
├── Minimal
├── Editorial
├── Bold
├── Photo-heavy
└── Educational

AI chọn template style.
Template không phải final design.

Nó là constraint/design language.

127. Suggested Generation Context
Design generator nhận:
Project
+
Brand
+
Canvas
+
Formula
+
Content
+
Template style
+
Image intent

Output:
Design JSON

128. End-to-End Example
User:
"Tạo carousel về 5 dấu hiệu trẻ sơ sinh đang thiếu ngủ."
Input:
{
  "topic": "5 dấu hiệu trẻ sơ sinh đang thiếu ngủ",
  "slide_count": 5,
  "cta": "save",
  "ratio": "4:5"
}

AI research:
Search current reliable sources

Content:
Slide 1:
5 dấu hiệu bé đang thiếu ngủ

Slide 2:
Dụi mắt liên tục

Slide 3:
Khó chịu / cáu gắt

Slide 4:
Khó đi vào giấc ngủ

Slide 5:
Lưu lại để kiểm tra sau

Formula:
Hook → Education → Education → Education → CTA

Design:
Slide 1:
Large centered headline
Photo background
Dark overlay

Slide 2:
Photo left
Text right

...

Image search:
3 images per slide

Editor:
User moves headline

Autosave:
Design version 12

Export:
carousel.zip

129. Development Phases
Phase 0 — Foundation
Backend
Go project
PostgreSQL
migrations
Redis
MinIO
configuration
logging
Docker Compose
health checks
Frontend
Next.js
shadcn/ui
auth shell
API client
TanStack Query
routing
130. Phase 1 — Authentication
Implement:
Google OAuth
Session
Current user
Logout
Dev login

Acceptance:
User can login
User remains authenticated after refresh
User can logout

131. Phase 2 — Project
Implement:
Project CRUD
Project dashboard
Project settings
Project AI context

132. Phase 3 — AI Project Intelligence
Implement:
Niche analysis
Research
Audience
Tone
Content pillars
CTA options

Add:
AI validation
retry
logging

133. Phase 4 — Dynamic Schema
Implement:
Schema generation
Schema validation
Schema renderer
Component registry
Dynamic form

This is a major milestone.
Acceptance:

Không sửa frontend code nhưng AI có thể tạo schema mới và frontend render được schema đó nếu sử dụng allowed component types.
134. Phase 5 — Carousel CRUD
Implement:
Create
Read
Update
Delete
Duplicate
Archive

135. Phase 6 — AI Generation Worker
Implement:
Redis
Jobs
Worker
Progress
Retry
Research
Content
Formula

136. Phase 7 — Design Generation
Implement:
Design JSON
Validation
Typography
Layout
Canvas constraints

137. Phase 8 — Fabric Editor
Implement theo thứ tự:
1. Canvas initialization
2. Load Design JSON
3. Render text
4. Render image
5. Select
6. Move
7. Resize
8. Edit text
9. Slide switching
10. Undo
11. Redo
12. Serialize
13. Autosave

Không nên build advanced editor trước khi basic pipeline hoạt động.
138. Phase 9 — Unsplash
Implement:
Search
3 candidates
Research more
Select
Download
MinIO
Attach asset
Render background

139. Phase 10 — Export
Implement:
Render
PNG
ZIP
MinIO
Download

140. Phase 11 — Hardening
Test:
AI failures
Redis failures
MinIO failures
Unsplash failures
Invalid schema
Invalid design
Large uploads
Concurrent edits
Expired sessions
Unauthorized access

141. Suggested Sprint Breakdown
Sprint 1
Infrastructure
Auth
Database
Project CRUD

Sprint 2
Project AI
Research
Project intelligence

Sprint 3
Dynamic JSON Schema
Schema Renderer
Create Carousel

Sprint 4
Redis
Worker
AI Content Generation

Sprint 5
Design JSON
Design Generator

Sprint 6
Fabric.js
Basic editor
Undo/Redo
Autosave

Sprint 7
Unsplash
MinIO
Asset management

Sprint 8
Export
ZIP
Error handling
Observability

Sprint 9
QA
Security
Performance
AI quality tuning
Production deployment

142. Critical Path
Nếu team muốn tối ưu thời gian, thứ tự quan trọng nhất là:
Auth
 ↓
Project
 ↓
AI Project Context
 ↓
Dynamic Schema
 ↓
Carousel
 ↓
Content AI
 ↓
Design JSON
 ↓
Fabric
 ↓
Images
 ↓
Export

Không nên build advanced editor trước AI → Design JSON pipeline.
143. What NOT to Build in MVP
Không build:
TikTok direct publishing
Instagram
Team collaboration
Workspace
Competitor analysis
TikTok analytics
Content calendar
AI video
Complex animation
Template marketplace
Advanced typography
Real-time collaboration
Billing
Admin UI

Architecture có thể chuẩn bị, nhưng implementation chưa cần.
144. Biggest Technical Risks
Risk 1 — AI generates poor Design JSON
Mitigation:
strict schema
allowed components
validation
design constraints
retry

Risk 2 — AI text overflows
Mitigation:
character limits
layout validation
AI shortening

Risk 3 — Generation takes too long
Mitigation:
async worker
progress UI
parallel image searches
caching

Risk 4 — Canvas becomes hard to maintain
Mitigation:
Design JSON
Fabric Adapter
isolated editor state

Risk 5 — AI costs grow quickly
Mitigation:
model abstraction
token tracking
caching
usage limits

Risk 6 — Editor/Export rendering mismatch (WYSIWYG bị phá vỡ)
Mitigation:
chọn rendering engine rõ ràng cho export (mục 155)
golden-slide regression test giữa editor screenshot và export output

Risk 7 — Unsplash compliance/rate limit chặn launch
Mitigation:
apply Production access sớm (Phase 0/1)
download-trigger + attribution đúng chuẩn (mục 156)

Risk 8 — Vietnamese text measurement lệch giữa client/server
Mitigation:
kiểm tra Vietnamese glyph coverage của từng font trong registry
đồng bộ logic đo text-width client/server (mục 157)

145. Important Architectural Decision
Một nguyên tắc cần giữ xuyên suốt:
AI ≠ Application Logic

AI đề xuất:
content
schema
formula
design
image intent

Application quyết định:
security
permissions
schema validity
allowed components
storage
state
export
business rules

Không để AI trở thành source of truth cho security hoặc application behavior.
146. Recommended MVP Definition
MVP được coi là thành công khi user có thể:
1. Login
2. Create project
3. Nhập niche
4. AI hiểu niche
5. AI tạo project configuration
6. Create carousel
7. AI tạo dynamic form
8. User nhập vài thông tin
9. AI research
10. AI viết content
11. AI chọn formula
12. AI tạo design
13. AI tìm 3 ảnh / slide
14. User chọn ảnh
15. Carousel render
16. User sửa text
17. User move/resize text
18. Undo/redo
19. Autosave
20. Export ZIP

Nếu toàn bộ flow này chạy ổn định, product đã đạt Production MVP.
147. Final System Architecture
                         ┌────────────────────┐
                         │      Next.js       │
                         │ React + shadcn/ui  │
                         └─────────┬──────────┘
                                   │
                                   │ HTTPS
                                   ▼
                         ┌────────────────────┐
                         │      Go API        │
                         │                    │
                         │ Auth               │
                         │ Projects           │
                         │ Carousels          │
                         │ Schema             │
                         │ Design             │
                         │ Assets             │
                         │ Export             │
                         └──────┬─────┬───────┘
                                │     │
                    ┌───────────┘     └─────────────┐
                    ▼                               ▼
             ┌─────────────┐                 ┌─────────────┐
             │ PostgreSQL  │                 │    Redis    │
             │             │                 │             │
             │ Users       │                 │ Jobs        │
             │ Projects    │                 │ Queue       │
             │ Carousels   │                 │ Cache       │
             │ Designs     │                 └──────┬──────┘
             │ Assets      │                        │
             └─────────────┘                        ▼
                                            ┌────────────────┐
                                            │ Go Worker      │
                                            │                │
                                            │ Research       │
                                            │ AI             │
                                            │ Images         │
                                            │ Export         │
                                            └───┬──────┬─────┘
                                                │      │
                          ┌─────────────────────┘      └────────────────┐
                          ▼                                            ▼
                  ┌───────────────┐                            ┌───────────────┐
                  │ AI Provider   │                            │   Unsplash    │
                  │ OpenAI        │                            │               │
                  │ Compatible    │                            │ Image Search  │
                  └───────────────┘                            └───────────────┘

                                      │
                                      ▼
                                ┌─────────────┐
                                │    MinIO    │
                                │             │
                                │ Images      │
                                │ Assets      │
                                │ Exports     │
                                └─────────────┘

148. Final Domain Architecture
User
 │
 └── Project
      │
      ├── Project Intelligence
      │
      ├── Project Schema
      │
      ├── Brand Kit
      │
      └── Carousel
           │
           ├── Input
           │
           ├── Research
           │    └── Sources
           │
           ├── Content
           │
           ├── Formula
           │
           ├── Design
           │    └── Slides
           │         └── Elements
           │
           ├── Assets
           │
           ├── Versions
           │
           └── Exports

149. The Most Important Product Abstraction
To make the product scalable, tôi đề xuất xem toàn bộ hệ thống như 5 tầng:
┌─────────────────────────────────────┐
│ 1. PROJECT INTELLIGENCE             │
│    "What is this TikTok channel?"   │
├─────────────────────────────────────┤
│ 2. CONTENT INPUT                    │
│    "What should this carousel say?" │
├─────────────────────────────────────┤
│ 3. CONTENT                          │
│    "What does each slide say?"      │
├─────────────────────────────────────┤
│ 4. DESIGN                           │
│    "How does each slide look?"      │
├─────────────────────────────────────┤
│ 5. RENDER                           │
│    "Turn design into images."       │
└─────────────────────────────────────┘

Đây là abstraction quan trọng nhất của sản phẩm.
Nó cho phép sau này:

TikTok
Instagram
Facebook
LinkedIn

có thể dùng chung:
Project Intelligence
Content
Design
Asset

và chỉ thay đổi:
Platform constraints
Canvas presets
Export

150. Recommended Implementation Priority
Nếu tôi là Product Owner chịu trách nhiệm delivery, tôi sẽ ưu tiên theo thứ tự:
Priority 1 — AI → Schema
Đảm bảo AI có thể hiểu niche và tạo form dynamic tốt.
Priority 2 — AI → Content
Đảm bảo content thực sự tốt.
Priority 3 — Content → Design JSON
Đảm bảo AI có thể biến content thành visual structure.
Priority 4 — Design JSON → Fabric
Đảm bảo visual editor hoạt động.
Priority 5 — Unsplash → Asset
Đảm bảo hình ảnh được đưa vào design.
Priority 6 — Export
Đảm bảo user lấy được sản phẩm cuối.
Lý do:

Nếu AI content + design không tốt thì việc xây một editor cực kỳ đẹp cũng không cứu được product.
MVP nên tập trung vào việc tạo ra trải nghiệm:
“Tôi nhập một ý tưởng → vài giây/phút sau có một carousel gần như hoàn chỉnh.”
Đó mới là core value proposition của sản phẩm.
151. Product North Star
North Star Metric đề xuất:
Number of AI-generated carousels successfully exported per active creator.
Supporting metrics:
Project creation rate
Carousel generation rate
Generation success rate
Generation → export conversion
Average edit time per carousel
Average number of manual edits
Image selection rate
Regeneration rate
Export success rate

Một metric đặc biệt quan trọng:
Average manual edit time per carousel
Mục tiêu sản phẩm là làm metric này giảm xuống.
Nếu user mất 15 phút chỉnh mỗi carousel, AI chưa làm đủ tốt.

Nếu user chỉ mất 30–90 giây để review/chỉnh nhẹ rồi export, product đang đi đúng hướng.

152. Final Product Goal
Sản phẩm cuối cùng không nên được định vị trong đầu team là:
“Canva cho TikTok Carousel.”
Mà nên được hiểu là:
“AI Content-to-Carousel Engine cho TikTok creators.”
Canva/editor chỉ là một phần của hệ thống.
Core moat nên nằm ở:

Niche Understanding
        +
Content Research
        +
Content Strategy
        +
Carousel Formula
        +
Content Generation
        +
Visual Design Generation
        +
Fast Editing

Canvas editor chỉ là lớp cuối để user có quyền kiểm soát.
153. Definition of Done cho Production MVP
MVP chỉ được coi là hoàn thành khi một creator mới có thể:
Sign in
   ↓
Create a niche-based project
   ↓
Let AI understand the channel
   ↓
Create a carousel
   ↓
Fill a dynamically generated form
   ↓
Wait while AI researches and generates
   ↓
Receive a multi-slide carousel
   ↓
Choose images
   ↓
Edit text if necessary
   ↓
Undo / redo changes
   ↓
Leave the editor and return later
   ↓
Open the same carousel
   ↓
See the saved design
   ↓
Export ZIP

và team có thể quan sát được:
What AI generated
What failed
How long it took
How much AI usage occurred
Which step caused errors

Nếu đạt được flow này với độ ổn định production, hệ thống đã có nền tảng rất tốt để mở rộng sang Instagram và các format content khác.
Nếu triển khai thực tế, tôi sẽ không bắt đầu code toàn bộ cùng lúc. Bước tiếp theo hợp lý nhất là chuyển blueprint này thành Engineering Specification: ERD PostgreSQL cụ thể, toàn bộ JSON Schema contract, API request/response, Redis job payload, Design JSON schema cho Fabric.js, Go interface, Next.js folder/component structure và backlog Jira theo từng ticket.

154. Addendum — Bổ sung sau review
Các mục 155–164 dưới đây bổ sung cho bản v1.0, tập trung vào những rủi ro thực thi (execution risk) chưa được cover trong bản gốc: rendering engine, compliance bên thứ ba, tiếng Việt, và cost/operations control. Đây là những quyết định nên chốt sớm (trước hoặc trong lúc viết Engineering Specification), vì càng để muộn chi phí sửa càng cao.

155. Export Rendering Engine Decision
Editor Renderer (Fabric.js/browser) và Export Renderer (server-side) là hai engine khác nhau (mục 48).
Go không có engine tương đương Fabric.js, nên phải chọn rõ một hướng trước khi code Phase 7/8.

Option A — Headless Browser Rendering
Design JSON
  ↓
Reuse Fabric.js code trong một Node service
  ↓
Headless Chromium (Playwright)
  ↓
Screenshot mỗi slide
  ↓
PNG

Ưu điểm: pixel-parity gần như tuyệt đối với editor.
Nhược điểm: thêm một Node runtime bên cạnh Go, tăng độ phức tạp deploy, chậm hơn native rendering.

Option B — Native Go Renderer
Design JSON
  ↓
Go canvas library (gg / go-skia)
  ↓
PNG

Ưu điểm: cùng ngôn ngữ với backend, nhẹ, dễ scale theo worker.
Nhược điểm: phải tự implement lại text layout/line-wrap/font metrics tương đương Fabric.js → rủi ro lệch giữa cái user thấy trong editor và ảnh export ra (phá vỡ kỳ vọng WYSIWYG).

Khuyến nghị:
Chốt option ngay ở Phase 7, và bổ sung một "golden slide" regression test: so sánh screenshot editor vs export output cho cùng một Design JSON, chạy như một test tự động chứ không chỉ review bằng mắt.

156. Unsplash API Compliance
Unsplash API có các ràng buộc bắt buộc, chưa được nhắc trong pipeline ở mục 33–37.

Download trigger:
Khi user chọn ảnh, backend phải gọi GET /photos/:id/download trước khi lưu vào MinIO. Đây là điều khoản bắt buộc trong Unsplash API Guidelines, không phải tuỳ chọn.

Attribution:
Nên lưu kèm mỗi asset nguồn Unsplash:
{
  "photographer_name": "...",
  "photographer_url": "...",
  "unsplash_url": "..."
}
để sẵn sàng hiển thị attribution nếu cần, kể cả khi MVP chưa show UI này.

Rate limit:
Demo app: 50 requests/hour.
Production app: cần Unsplash duyệt, ~5000 requests/hour.
Action: nên apply Production access ngay từ Phase 0/1 — approval có thể mất thời gian và có thể chặn launch nếu để tới cuối.

157. Vietnamese Typography & Text Measurement
Ví dụ sản phẩm dùng tiếng Việt (mục 128), nhưng font_registry (mục 40) chưa xác nhận font nào hỗ trợ đầy đủ dấu tiếng Việt (ư, ơ, dấu thanh chồng dấu) khi render server-side.

Cần:
Kiểm tra Vietnamese glyph coverage của từng font trước khi thêm vào allowed registry.
Đảm bảo text-width measurement nhất quán giữa:
  Client — canvas measureText(), dùng cho overflow warning trong editor.
  Server — renderer đo text theo đúng cùng logic, dùng cho export và cho design validation ở mục 121.
Nếu hai bên đo lệch nhau, overflow detection (mục 121) sẽ sai lệch — đây là nguồn bug tinh vi và khó debug nhất trong toàn bộ design pipeline.

158. Job Retry & Checkpointing
Mục 14 (Generation Job State Machine) chỉ định nghĩa FAILED → RETRY, chưa nói retry resume từ đâu.

Vấn đề:
Nếu retry chạy lại toàn bộ pipeline từ QUEUED, mỗi lần fail ở bước cuối (vd SEARCHING_IMAGES) sẽ tốn lại chi phí AI research + content + design đã chạy thành công trước đó.

Đề xuất:
generation_jobs lưu thêm:
last_completed_step
step_outputs (research_id, content_id, design_id, ...)

Retry logic:
FAILED tại step X
   ↓
Resume từ last_completed_step
   ↓
Không chạy lại các step đã COMPLETED

159. Idempotency cho Generation Request
POST /projects/:id/carousels/generate hiện chưa có cơ chế chống double-submit (double click, mất mạng rồi retry từ frontend).

Đề xuất:
Client gửi kèm Idempotency-Key (UUID) mỗi lần submit form.
Backend: nếu key đã tồn tại và đang xử lý/đã xử lý → trả về job hiện có thay vì tạo job mới.

160. Content Moderation & Compliance Layer
Nhiều niche khả năng đụng nội dung nhạy cảm (sức khoẻ trẻ em, tài chính cá nhân, y tế...). Mục 119 có tiêu chí "No unsupported claims" nhưng chưa có cơ chế enforce cụ thể.

Đề xuất bổ sung vào Content Generation pipeline (mục 18):
Content JSON
   ↓
Moderation check (banned topics, absolute medical/financial claims)
   ↓
Nếu vi phạm: reject + regenerate với constraint chặt hơn
   ↓
Persist

Ở mức Project, nên có danh sách niche/topic nhạy cảm cần disclaimer tự động (vd: chèn câu miễn trừ trách nhiệm ở slide cuối cho nội dung y tế).

161. AI Prompt Evaluation Harness
Mục 117 (Prompt Management) đã có versioning, nhưng chưa có cách đo prompt version mới có tốt hơn version cũ hay không.

Đề xuất:
prompt_eval_sets
------------------
id
task_type       (vd: content_generation)
niche
input_json
expected_criteria_json

Khi sửa prompt:
Prompt v3
   ↓
Run against eval set (5–10 niche mẫu)
   ↓
So sánh output v2 vs v3 (structural + semantic validation, mục 119)
   ↓
Approve nếu không regress

Không bắt buộc ngay ngày đầu MVP, nhưng nên có trước khi bắt đầu tune prompt thường xuyên — nếu không, mỗi lần sửa prompt là một canh bạc không đo lường được.

162. Session & Auth Mechanism
Mục 84–86 đã nêu yêu cầu bảo mật (secure cookies, CSRF) nhưng chưa chốt cơ chế session cụ thể.

Đề xuất:
Server-side session (session ID trong httpOnly cookie, session data trong Redis/Postgres) thay vì JWT thuần — dễ revoke khi logout, không cần refresh-token flow phức tạp cho một sản phẩm cá nhân ở giai đoạn MVP.

sessions
--------
id
user_id
created_at
expires_at
last_seen_at

163. Rate Limit Baseline Numbers (MVP)
Mục 87 nêu nguyên tắc nhưng chưa có con số cụ thể. Đề xuất baseline cho MVP (điều chỉnh sau khi có dữ liệu thật):

carousel generation: 20/ngày/user
image search ("research more"): 30/ngày/user
export: 20/ngày/user
upload: 50MB/file, 100 file/ngày/user

Vượt giới hạn → trả lỗi rõ ràng ("Bạn đã đạt giới hạn hôm nay"), không silent fail.

164. Internationalization Scope
Cần chốt rõ trước khi code frontend:
UI language (giao diện sản phẩm) — cố định một ngôn ngữ cho MVP, hay đa ngôn ngữ ngay từ đầu?
Content language (nội dung carousel) — đã có sẵn trong Project Identity (mục 7, field language), độc lập với UI language.

Đề xuất cho MVP: UI cố định một ngôn ngữ, content language theo Project — tránh chi phí i18n UI không cần thiết ở giai đoạn này.