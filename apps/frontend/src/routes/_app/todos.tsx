import { useGetAllCategories } from "@/api/hooks/use-category-query";
import { useDebounce } from "@/api/hooks/use-debounce";
import { useGetAllTodos } from "@/api/hooks/use-todo-query";
import { TodoCard } from "@/components/todos/todo-card";
import { TodoCreateForm } from "@/components/todos/todo-create-form";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import {
  buildTodoApiQuery,
  defaultTodoRouteSearch,
  getTodoSortValue,
  hasActiveTodoFilters,
  parseTodoRouteSearch,
  type TodoRouteSearch,
} from "@/routes/-search";
import { createFileRoute } from "@tanstack/react-router";
import {
  Plus,
  Search,
  SlidersHorizontal,
  X,
  Sparkles,
  CheckCircle2,
  Clock,
  Archive,
  FileText,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { AnimatePresence, motion } from "motion/react";
import { useEffect, useState } from "react";

export const Route = createFileRoute("/_app/todos")({
  validateSearch: parseTodoRouteSearch,
  component: TodosPage,
});

type TodoStatus = TodoRouteSearch["status"];
type TodoPriority = TodoRouteSearch["priority"];

const statusConfig = {
  all: { label: "All", icon: CheckCircle2, color: "text-foreground" },
  draft: { label: "Draft", icon: FileText, color: "text-muted-foreground" },
  active: { label: "Active", icon: Clock, color: "text-chart-2" },
  completed: { label: "Completed", icon: CheckCircle2, color: "text-accent" },
  archived: {
    label: "Archived",
    icon: Archive,
    color: "text-muted-foreground",
  },
} as const;

function TodosPage() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();

  const [searchInput, setSearchInput] = useState(search.q ?? "");
  const [selectedTodos, setSelectedTodos] = useState<string[]>([]);
  const [showFilters, setShowFilters] = useState(false);

  const debouncedSearch = useDebounce(searchInput, 300);

  useEffect(() => {
    setSearchInput(search.q ?? "");
  }, [search.q]);

  useEffect(() => {
    if (debouncedSearch === (search.q ?? "")) {
      return;
    }

    void navigate({
      search: (prev: TodoRouteSearch) => ({
        ...prev,
        q: debouncedSearch || undefined,
        page: 1,
      }),
      replace: true,
    });
  }, [debouncedSearch, navigate, search.q]);

  useEffect(() => {
    setSelectedTodos([]);
  }, [
    search.categoryId,
    search.order,
    search.page,
    search.priority,
    search.q,
    search.sort,
    search.status,
  ]);

  const { data: todos, isLoading } = useGetAllTodos({
    query: buildTodoApiQuery(search),
  });

  const { data: categories } = useGetAllCategories({
    query: { page: 1, limit: 100 },
  });

  const updateSearch = (
    updater: (prev: TodoRouteSearch) => TodoRouteSearch,
    replace = false,
  ) => {
    void navigate({ search: updater, replace });
  };

  const handleSelectTodo = (todoId: string, checked: boolean) => {
    if (checked) {
      setSelectedTodos((prev) =>
        prev.includes(todoId) ? prev : [...prev, todoId],
      );
      return;
    }

    setSelectedTodos((prev) => prev.filter((id) => id !== todoId));
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked && todos?.data) {
      setSelectedTodos(todos.data.map((todo) => todo.id));
      return;
    }

    setSelectedTodos([]);
  };

  const clearFilters = () => {
    setSearchInput("");
    updateSearch(() => defaultTodoRouteSearch, true);
  };

  const clearSearch = () => {
    setSearchInput("");
    updateSearch(
      (prev) => ({
        ...prev,
        q: undefined,
        page: 1,
      }),
      true,
    );
  };

  const hasActiveFilters =
    Boolean(searchInput.trim()) || hasActiveTodoFilters(search);
  const visibleTodoCount = todos?.data.length ?? 0;
  const allVisibleTodosSelected =
    visibleTodoCount > 0 && selectedTodos.length === visibleTodoCount;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
        >
          <h1 className="text-3xl md:text-4xl font-bold tracking-tight">
            Tasks
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage and organize your tasks
          </p>
        </motion.div>
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.4, delay: 0.1 }}
        >
          <TodoCreateForm>
            <Button className="shadow-soft press-scale group">
              <Plus className="w-4 h-4 mr-2" />
              New Task
            </Button>
          </TodoCreateForm>
        </motion.div>
      </div>

      {/* Search and Filters */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, delay: 0.1 }}
      >
        <Card className="shadow-soft overflow-hidden">
          <CardContent className="p-4 space-y-4">
            {/* Search Bar */}
            <div className="flex flex-col sm:flex-row gap-3">
              <div className="relative flex-1">
                <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder="Search tasks..."
                  value={searchInput}
                  onChange={(e) => setSearchInput(e.target.value)}
                  className="pl-10 h-11 bg-muted/30 border-transparent focus:border-accent focus:bg-background transition-colors"
                />
                {searchInput && (
                  <button
                    type="button"
                    onClick={clearSearch}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  >
                    <X className="w-4 h-4" />
                  </button>
                )}
              </div>
              <Button
                variant={showFilters ? "secondary" : "outline"}
                onClick={() => setShowFilters((prev) => !prev)}
                className={cn(
                  "gap-2 shrink-0",
                  hasActiveFilters && "ring-2 ring-accent/20",
                )}
              >
                <SlidersHorizontal className="w-4 h-4" />
                Filters
                {hasActiveFilters && (
                  <span className="w-2 h-2 rounded-full bg-accent" />
                )}
              </Button>
            </div>

            {/* Filter Panel */}
            <AnimatePresence>
              {showFilters && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: "auto", opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.2 }}
                  className="overflow-hidden"
                >
                  <div className="pt-4 border-t border-border/50">
                    <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-3">
                      <Select
                        value={search.categoryId ?? "all"}
                        onValueChange={(value) =>
                          updateSearch(
                            (prev) => ({
                              ...prev,
                              categoryId: value === "all" ? undefined : value,
                              page: 1,
                            }),
                            true,
                          )
                        }
                      >
                        <SelectTrigger className="bg-muted/30 border-transparent focus:border-accent">
                          <SelectValue placeholder="Category" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="all">All Categories</SelectItem>
                          {categories?.data?.map((category) => (
                            <SelectItem key={category.id} value={category.id}>
                              <div className="flex items-center gap-2">
                                <div
                                  className="w-3 h-3 rounded-md"
                                  style={{ backgroundColor: category.color }}
                                />
                                {category.name}
                              </div>
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>

                      <Select
                        value={search.priority}
                        onValueChange={(value) =>
                          updateSearch(
                            (prev) => ({
                              ...prev,
                              priority: value as TodoPriority,
                              page: 1,
                            }),
                            true,
                          )
                        }
                      >
                        <SelectTrigger className="bg-muted/30 border-transparent focus:border-accent">
                          <SelectValue placeholder="Priority" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="all">All Priorities</SelectItem>
                          <SelectItem value="high">
                            <div className="flex items-center gap-2">
                              <div className="w-2 h-2 rounded-full bg-destructive" />
                              High
                            </div>
                          </SelectItem>
                          <SelectItem value="medium">
                            <div className="flex items-center gap-2">
                              <div className="w-2 h-2 rounded-full bg-accent" />
                              Medium
                            </div>
                          </SelectItem>
                          <SelectItem value="low">
                            <div className="flex items-center gap-2">
                              <div className="w-2 h-2 rounded-full bg-chart-2" />
                              Low
                            </div>
                          </SelectItem>
                        </SelectContent>
                      </Select>

                      <Select
                        value={getTodoSortValue(search)}
                        onValueChange={(value) => {
                          const [sort, order] = value.split("-") as [
                            TodoRouteSearch["sort"],
                            TodoRouteSearch["order"],
                          ];

                          updateSearch(
                            (prev) => ({
                              ...prev,
                              sort,
                              order,
                              page: 1,
                            }),
                            true,
                          );
                        }}
                      >
                        <SelectTrigger className="bg-muted/30 border-transparent focus:border-accent">
                          <SelectValue placeholder="Sort by" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="updated_at-desc">
                            Latest
                          </SelectItem>
                          <SelectItem value="updated_at-asc">Oldest</SelectItem>
                          <SelectItem value="title-asc">Title A-Z</SelectItem>
                          <SelectItem value="title-desc">Title Z-A</SelectItem>
                          <SelectItem value="priority-desc">
                            High Priority
                          </SelectItem>
                          <SelectItem value="due_date-asc">Due Date</SelectItem>
                        </SelectContent>
                      </Select>

                      <Button
                        variant="ghost"
                        onClick={clearFilters}
                        disabled={!hasActiveFilters}
                        className="text-muted-foreground"
                      >
                        <X className="w-4 h-4 mr-2" />
                        Clear filters
                      </Button>
                    </div>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </CardContent>
        </Card>
      </motion.div>

      {/* Status Tabs */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, delay: 0.15 }}
      >
        <Tabs
          value={search.status}
          onValueChange={(value) =>
            updateSearch(
              (prev) => ({
                ...prev,
                status: value as TodoStatus,
                page: 1,
              }),
              true,
            )
          }
        >
          <TabsList className="h-auto p-1.5 bg-muted/50 flex-wrap">
            {(
              Object.keys(statusConfig) as Array<keyof typeof statusConfig>
            ).map((status) => {
              const config = statusConfig[status];
              const Icon = config.icon;
              return (
                <TabsTrigger
                  key={status}
                  value={status}
                  className={cn(
                    "gap-2 px-4 py-2 data-[state=active]:shadow-soft",
                    search.status === status && config.color,
                  )}
                >
                  <Icon className="w-4 h-4" />
                  {config.label}
                </TabsTrigger>
              );
            })}
          </TabsList>

          <TabsContent value={search.status} className="mt-6 space-y-4">
            {/* Bulk Selection */}
            <AnimatePresence>
              {selectedTodos.length > 0 && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                >
                  <Card className="border-accent/30 bg-accent/5">
                    <CardContent className="py-3 flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <Checkbox
                          checked={allVisibleTodosSelected}
                          onCheckedChange={(checked) =>
                            handleSelectAll(Boolean(checked))
                          }
                        />
                        <span className="text-sm font-medium">
                          {selectedTodos.length} selected
                        </span>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setSelectedTodos([])}
                      >
                        Clear selection
                      </Button>
                    </CardContent>
                  </Card>
                </motion.div>
              )}
            </AnimatePresence>

            {/* Todo List */}
            <div className="space-y-3">
              {isLoading ? (
                <div className="space-y-3">
                  {Array.from({ length: 5 }).map((_, i) => (
                    <Card key={i} className="shadow-soft">
                      <CardContent className="p-4">
                        <div className="space-y-3">
                          <Skeleton className="h-5 w-3/4" />
                          <Skeleton className="h-4 w-1/2" />
                          <div className="flex gap-2">
                            <Skeleton className="h-6 w-20 rounded-full" />
                            <Skeleton className="h-6 w-20 rounded-full" />
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              ) : todos?.data?.length ? (
                <>
                  {/* Select All */}
                  {todos.data.length > 1 && (
                    <Card className="shadow-soft bg-muted/20">
                      <CardContent className="py-3">
                        <div className="flex items-center gap-3">
                          <Checkbox
                            checked={allVisibleTodosSelected}
                            onCheckedChange={(checked) =>
                              handleSelectAll(Boolean(checked))
                            }
                          />
                          <span className="text-sm text-muted-foreground">
                            Select all {todos.data.length} tasks
                          </span>
                        </div>
                      </CardContent>
                    </Card>
                  )}

                  {/* Task Cards */}
                  <div className="space-y-3">
                    {todos.data.map((todo) => (
                      <div key={todo.id} className="flex items-start gap-3">
                        <div className="pt-4">
                          <Checkbox
                            checked={selectedTodos.includes(todo.id)}
                            onCheckedChange={(checked) =>
                              handleSelectTodo(todo.id, Boolean(checked))
                            }
                          />
                        </div>
                        <div className="flex-1">
                          <TodoCard todo={todo} />
                        </div>
                      </div>
                    ))}
                  </div>

                  {/* Pagination */}
                  {todos.totalPages > 1 && (
                    <div className="flex items-center justify-between pt-4">
                      <p className="text-sm text-muted-foreground">
                        Showing {(search.page - 1) * 20 + 1} to{" "}
                        {Math.min(search.page * 20, todos.total)} of{" "}
                        {todos.total} tasks
                      </p>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={search.page <= 1}
                          onClick={() =>
                            updateSearch((prev) => ({
                              ...prev,
                              page: prev.page - 1,
                            }))
                          }
                          className="gap-1"
                        >
                          <ChevronLeft className="w-4 h-4" />
                          Previous
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={search.page >= todos.totalPages}
                          onClick={() =>
                            updateSearch((prev) => ({
                              ...prev,
                              page: prev.page + 1,
                            }))
                          }
                          className="gap-1"
                        >
                          Next
                          <ChevronRight className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>
                  )}
                </>
              ) : (
                <motion.div
                  className="py-16 text-center"
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ duration: 0.4 }}
                >
                  <div className="w-16 h-16 mx-auto rounded-2xl bg-muted/50 flex items-center justify-center mb-4">
                    <Sparkles className="w-8 h-8 text-muted-foreground" />
                  </div>
                  <p className="text-lg font-medium text-muted-foreground">
                    No tasks found
                  </p>
                  <p className="text-sm text-muted-foreground/70 mt-1">
                    {hasActiveFilters
                      ? "Try adjusting your filters"
                      : "Create your first task to get started"}
                  </p>
                  {hasActiveFilters && (
                    <Button
                      variant="outline"
                      className="mt-4"
                      onClick={clearFilters}
                    >
                      Clear filters
                    </Button>
                  )}
                </motion.div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </motion.div>
    </div>
  );
}
