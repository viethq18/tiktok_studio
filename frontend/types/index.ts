// Mirrors of the Go contracts. The frontend depends on these normalized shapes,
// never on raw AI output (blueprint §7).

export type User = {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
  provider: string;
};

export type BrandScope = 'off' | 'first' | 'last' | 'first_last' | 'all';
export type BrandPosition =
  | 'top_left' | 'top_center' | 'top_right'
  | 'bottom_left' | 'bottom_center' | 'bottom_right';

export type Placement = { scope: BrandScope; position: BrandPosition };

export type BrandDisplay = {
  handle: Placement;
  logo: Placement;
  website: Placement;
};

export type Brand = {
  logo_asset_id?: string;
  logo_url?: string;
  primary_color?: string;
  secondary_color?: string;
  accent_color?: string;
  font_family?: string;
  website?: string;
  handle?: string;
  display: BrandDisplay;
};

export const DEFAULT_BRAND_DISPLAY: BrandDisplay = {
  handle: { scope: 'all', position: 'bottom_center' },
  logo: { scope: 'first_last', position: 'top_left' },
  website: { scope: 'last', position: 'bottom_center' },
};

export type Caption = {
  caption: string;
  hashtags: string[];
  edited: boolean;
  generated_at?: string;
  updated_at: string;
};

export type Pillar = { id: string; name: string; description?: string };
export type CTAOption = { id: string; label: string };

export type ProjectContext = {
  schema_version: string;
  audience: { description: string; demographics: string; pain_points: string[] };
  tone: string[];
  writing_style: string;
  preferred_angles: string[];
  content_pillars: Pillar[];
  cta_options: CTAOption[];
  suggested_topics: string[];
};

export type Project = {
  id: string;
  name: string;
  niche: string;
  description: string;
  language: string;
  status: string;
  brand: Brand;
  context?: ProjectContext;
  context_version: number;
  carousel_count: number;
  created_at: string;
  updated_at: string;
};

// ---- dynamic form (§9) ----

export type FieldComponent =
  | 'text' | 'textarea' | 'number' | 'select' | 'multi_select'
  | 'radio' | 'checkbox' | 'slider' | 'image' | 'url';

export type FormProperty = {
  name: string;
  type: 'string' | 'integer' | 'number' | 'boolean' | 'array';
  title: string;
  description?: string;
  enum?: string[];
  enum_labels?: Record<string, string>;
  minimum?: number;
  maximum?: number;
  max_length?: number;
  default?: unknown;
  component: FieldComponent;
  placeholder?: string;
  help?: string;
  required: boolean;
  items?: string;
};

export type Form = { version: number; properties: FormProperty[] };

// ---- carousel ----

export type CarouselStatus = 'draft' | 'generating' | 'ready' | 'archived' | 'failed';

export type Carousel = {
  id: string;
  project_id: string;
  title: string;
  status: CarouselStatus;
  platform: string;
  canvas_ratio: string;
  canvas_width: number;
  canvas_height: number;
  formula_id: string;
  thumbnail_url?: string;
  schema_version: number;
  input?: Record<string, unknown>;
  design_version: number;
  created_at: string;
  updated_at: string;
};

export type JobStep = { status: string; label: string; state: 'pending' | 'active' | 'done' };

export type Job = {
  id: string;
  carousel_id: string;
  status: string;
  progress: number;
  current_step: string;
  error_message?: string;
  steps: JobStep[];
};

// ---- design JSON (§22) ----

export type TextStyle = {
  fontFamily: string;
  fontSize: number;
  fontWeight: number;
  color: string;
  textAlign: 'left' | 'center' | 'right';
  lineHeight: number;
};

export type Overlay = { color: string; opacity: number };

export type DesignElement = {
  id: string;
  type: 'text' | 'image' | 'shape';
  z: number;
  x: number;
  y: number;
  width: number;
  height: number;
  opacity: number;
  locked?: boolean;
  content?: string;
  role?: string;
  style?: TextStyle;
  asset_id?: string;
  url?: string;
  fit?: string;
  overlay?: Overlay;
  shape?: string;
  fill?: string;
  radius?: number;
};

export type DesignSlide = {
  id: string;
  index: number;
  role: string;
  template?: string;
  background: string;
  image_intent?: string;
  image_query?: string;
  elements: DesignElement[];
};

export type Palette = {
  primary: string;
  secondary: string;
  accent: string;
  background: string;
  text: string;
};

export type SafeArea = { x: number; y: number; width: number; height: number };

export type Design = {
  version: string;
  canvas: { width: number; height: number; ratio: string; safe_area: SafeArea };
  palette: Palette;
  slides: DesignSlide[];
};

export type DesignView = { version: number; design: Design };

// ---- images / assets ----

export type ImageCandidate = {
  id: string;
  provider: string;
  thumb_url: string;
  preview_url: string;
  width: number;
  height: number;
  color?: string;
  description?: string;
  photographer_name?: string;
  photographer_url?: string;
  source_url?: string;
};

export type Asset = {
  id: string;
  source: string;
  mime_type: string;
  width: number;
  height: number;
  url?: string;
};

export type ExportJob = {
  id: string;
  carousel_id: string;
  status: string;
  progress: number;
  error?: string;
  download_url?: string;
};

export type ContentLanguage = {
  code: string;
  native_name: string;
  english_name: string;
};

export type Registry = {
  languages: ContentLanguage[];
  fonts: { id: string; family: string; css_stack: string; weights: number[]; vietnamese: boolean }[];
  formulas: { id: string; name: string; description: string; structure: string[] }[];
  presets: { ratio: string; width: number; height: number; label: string }[];
};
