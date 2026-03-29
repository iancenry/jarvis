import {
  buildCategoryApiQuery,
  buildTodoApiQuery,
  defaultCategoryRouteSearch,
  defaultTodoRouteSearch,
  getTodoSortValue,
  hasActiveCategoryFilters,
  hasActiveTodoFilters,
  parseCategoryRouteSearch,
  parseTodoRouteSearch,
} from "./-search";
import { describe, expect, it } from "vitest";

describe("parseTodoRouteSearch", () => {
  it("applies stable defaults", () => {
    expect(parseTodoRouteSearch({})).toEqual(defaultTodoRouteSearch);
  });

  it("normalizes invalid or empty search params", () => {
    expect(
      parseTodoRouteSearch({
        page: "0",
        q: "   ",
        categoryId: "not-a-uuid",
        status: "unknown",
        priority: "unknown",
        sort: "unknown",
        order: "unknown",
      }),
    ).toEqual(defaultTodoRouteSearch);
  });

  it("keeps valid filter values", () => {
    expect(
      parseTodoRouteSearch({
        page: "3",
        q: " launch ",
        status: "active",
        priority: "high",
        sort: "due_date",
        order: "asc",
        categoryId: "11111111-1111-1111-1111-111111111111",
      }),
    ).toEqual({
      page: 3,
      q: "launch",
      status: "active",
      priority: "high",
      sort: "due_date",
      order: "asc",
      categoryId: "11111111-1111-1111-1111-111111111111",
    });
  });
});

describe("todo search helpers", () => {
  it("reports active filters and maps API query shape", () => {
    const search = parseTodoRouteSearch({
      q: "draft plan",
      status: "completed",
      priority: "low",
      sort: "title",
      order: "asc",
    });

    expect(hasActiveTodoFilters(search)).toBe(true);
    expect(getTodoSortValue(search)).toBe("title-asc");
    expect(buildTodoApiQuery(search)).toEqual({
      page: 1,
      limit: 20,
      search: "draft plan",
      status: "completed",
      categoryId: undefined,
      priority: "low",
      sort: "title",
      order: "asc",
    });
  });

  it("removes all-filters values from the API query", () => {
    expect(buildTodoApiQuery(defaultTodoRouteSearch)).toEqual({
      page: 1,
      limit: 20,
      search: undefined,
      status: undefined,
      categoryId: undefined,
      priority: undefined,
      sort: "updated_at",
      order: "desc",
    });
  });
});

describe("category search helpers", () => {
  it("applies stable defaults and trims search", () => {
    expect(parseCategoryRouteSearch({})).toEqual(defaultCategoryRouteSearch);
    expect(parseCategoryRouteSearch({ q: "  work  ", page: "2" })).toEqual({
      page: 2,
      q: "work",
      sort: "name",
      order: "asc",
    });
  });

  it("detects active filters and maps category API params", () => {
    const search = parseCategoryRouteSearch({
      q: "finance",
      sort: "updated_at",
      order: "desc",
      page: "4",
    });

    expect(hasActiveCategoryFilters(search)).toBe(true);
    expect(buildCategoryApiQuery(search)).toEqual({
      page: 4,
      limit: 20,
      search: "finance",
      sort: "updated_at",
      order: "desc",
    });
  });
});
