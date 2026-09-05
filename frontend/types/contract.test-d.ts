/**
 * Compile-time drift guard.
 *
 * `types/index.ts` is hand-written for ergonomics; `types/api.generated.ts` is
 * derived from the Go structs by `make apigen`. These assertions fail the type
 * check when the two diverge, so a backend rename, removal or type change
 * cannot reach runtime.
 *
 * The guard is deliberately one-directional: the web app may ignore fields it
 * does not need, but every field it *does* declare must exist on the backend
 * with a compatible type. Nothing here is imported at runtime — `tsc --noEmit`
 * is the enforcement.
 */
import type * as Api from './api.generated';
import type { Brand, Caption, Carousel, Design, DesignView, ExportJob, Project, User } from './index';

/**
 * `AssertNever` must be applied at the call site with concrete types. Wrapping
 * it in another generic would make TypeScript check the constraint against the
 * unresolved parameter, which fails for every input.
 */
type AssertNever<T extends never> = T;

/** The keys a local type declares that the backend does not have. */
type Extra<Local, Remote> = Exclude<keyof Local, keyof Remote>;

/** Fails to compile unless the fields Local declares match Remote's. */
type Compatible<Local extends Pick<Remote, Extract<keyof Local, keyof Remote>>, Remote> = true;

// `context` and `input` are free-form JSON server-side (json.RawMessage), so the
// web app narrows them locally; compare the rest.
type ProjectRemote = Omit<Api.Project, 'context'> & { context?: unknown };
type CarouselRemote = Omit<Api.Carousel, 'input'> & { input?: unknown };
type BrandRemote = Api.ProjectBrand & { display: Api.ProjectBrandDisplay };

export type _UserKeys = AssertNever<Extra<User, Api.AuthUser>>;
export type _User = Compatible<User, Api.AuthUser>;

export type _ProjectKeys = AssertNever<Extra<Project, ProjectRemote>>;
export type _Project = Compatible<Project, ProjectRemote>;

export type _BrandKeys = AssertNever<Extra<Brand, BrandRemote>>;
export type _Brand = Compatible<Brand, BrandRemote>;

export type _CarouselKeys = AssertNever<Extra<Carousel, CarouselRemote>>;
export type _Carousel = Compatible<Carousel, CarouselRemote>;

export type _DesignKeys = AssertNever<Extra<Design, Api.Design>>;
export type _Design = Compatible<Design, Api.Design>;

export type _DesignViewKeys = AssertNever<Extra<DesignView, Api.CarouselDesignView>>;
export type _DesignView = Compatible<DesignView, Api.CarouselDesignView>;

export type _CaptionKeys = AssertNever<Extra<Caption, Api.CarouselCaption>>;
export type _Caption = Compatible<Caption, Api.CarouselCaption>;

export type _ExportJobKeys = AssertNever<Extra<ExportJob, Api.ExportJob>>;
export type _ExportJob = Compatible<ExportJob, Api.ExportJob>;
