import { ZTodoPriority, ZTodoStatus } from "@jarvis/zod";
import { z } from "zod";

const searchTermSchema = z.preprocess((value) => {
  if (typeof value !== "string") {
    return undefined;
  }

  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
}, z.string().max(255).optional());

const pageSchema = z.coerce.number().int().min(1).catch(1);
const sortOrderSchema = z.enum(["asc", "desc"]);

const todoStatusSearchSchema = z.enum(["all", ...ZTodoStatus.options]);
const todoPrioritySearchSchema = z.enum(["all", ...ZTodoPriority.options]);
const todoSortSchema = z.enum([
  "created_at",
  "updated_at",
  "title",
  "priority",
  "due_date",
  "status",
]);
const categorySortSchema = z.enum(["created_at", "updated_at", "name"]);

const todoRouteSearchSchema = z.object({
  page: pageSchema.default(1),
  q: searchTermSchema.catch(undefined),
  status: todoStatusSearchSchema.catch("all").default("all"),
  categoryId: z
    .preprocess(
      (value) =>
        typeof value === "string" && value.length > 0 ? value : undefined,
      z.string().uuid().optional(),
    )
    .catch(undefined),
  priority: todoPrioritySearchSchema.catch("all").default("all"),
  sort: todoSortSchema.catch("updated_at").default("updated_at"),
  order: sortOrderSchema.catch("desc").default("desc"),
});

const categoryRouteSearchSchema = z.object({
  page: pageSchema.default(1),
  q: searchTermSchema.catch(undefined),
  sort: categorySortSchema.catch("name").default("name"),
  order: sortOrderSchema.catch("asc").default("asc"),
});

export type TodoRouteSearch = z.infer<typeof todoRouteSearchSchema>;
export type CategoryRouteSearch = z.infer<typeof categoryRouteSearchSchema>;

export const defaultTodoRouteSearch: TodoRouteSearch = {
  page: 1,
  q: undefined,
  status: "all",
  categoryId: undefined,
  priority: "all",
  sort: "updated_at",
  order: "desc",
};

export const defaultCategoryRouteSearch: CategoryRouteSearch = {
  page: 1,
  q: undefined,
  sort: "name",
  order: "asc",
};

export function parseTodoRouteSearch(
  search: Record<string, unknown>,
): TodoRouteSearch {
  return todoRouteSearchSchema.parse(search);
}

export function parseCategoryRouteSearch(
  search: Record<string, unknown>,
): CategoryRouteSearch {
  return categoryRouteSearchSchema.parse(search);
}

export function hasActiveTodoFilters(search: TodoRouteSearch): boolean {
  return (
    Boolean(search.q) ||
    search.status !== defaultTodoRouteSearch.status ||
    search.categoryId !== undefined ||
    search.priority !== defaultTodoRouteSearch.priority ||
    search.sort !== defaultTodoRouteSearch.sort ||
    search.order !== defaultTodoRouteSearch.order
  );
}

export function hasActiveCategoryFilters(search: CategoryRouteSearch): boolean {
  return (
    Boolean(search.q) ||
    search.sort !== defaultCategoryRouteSearch.sort ||
    search.order !== defaultCategoryRouteSearch.order
  );
}

export function getTodoSortValue(search: TodoRouteSearch): string {
  return `${search.sort}-${search.order}`;
}

export function buildTodoApiQuery(search: TodoRouteSearch) {
  return {
    page: search.page,
    limit: 20,
    search: search.q,
    status: search.status === "all" ? undefined : search.status,
    categoryId: search.categoryId,
    priority: search.priority === "all" ? undefined : search.priority,
    sort: search.sort,
    order: search.order,
  };
}

export function buildCategoryApiQuery(search: CategoryRouteSearch) {
  return {
    page: search.page,
    limit: 20,
    search: search.q,
    sort: search.sort,
    order: search.order,
  };
}
