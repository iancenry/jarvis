import { useGetAllCategories } from "@/api/hooks/use-category-query";
import { useGetTodoStats, useGetAllTodos } from "@/api/hooks/use-todo-query";
import { TodoCard } from "@/components/todos/todo-card";
import { TodoCreateForm } from "@/components/todos/todo-create-form";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  defaultCategoryRouteSearch,
  defaultTodoRouteSearch,
} from "@/routes/-search";
import { createFileRoute, Link } from "@tanstack/react-router";
import {
  CheckCircle2,
  Clock,
  AlertTriangle,
  Archive,
  Plus,
  ArrowRight,
  Sparkles,
  TrendingUp,
} from "lucide-react";
import { motion } from "motion/react";

export const Route = createFileRoute("/_app/dashboard")({
  component: DashboardPage,
});

function DashboardPage() {
  const { data: stats, isLoading: statsLoading } = useGetTodoStats();
  const { data: recentTodos, isLoading: todosLoading } = useGetAllTodos({
    query: { page: 1, limit: 5, sort: "updated_at", order: "desc" },
  });
  const { data: categories, isLoading: categoriesLoading } =
    useGetAllCategories({
      query: { page: 1, limit: 10 },
    });

  const statCards = [
    {
      title: "Total Tasks",
      value: stats?.total || 0,
      icon: CheckCircle2,
      gradient: "from-chart-2/20 to-chart-2/5",
      iconColor: "text-chart-2",
      description: "All tracked tasks",
    },
    {
      title: "Active",
      value: stats?.active || 0,
      icon: Clock,
      gradient: "from-accent/20 to-accent/5",
      iconColor: "text-accent",
      description: "In progress",
    },
    {
      title: "Overdue",
      value: stats?.overdue || 0,
      icon: AlertTriangle,
      gradient: "from-destructive/20 to-destructive/5",
      iconColor: "text-destructive",
      description: "Need attention",
    },
    {
      title: "Completed",
      value: stats?.completed || 0,
      icon: Archive,
      gradient: "from-muted to-muted/30",
      iconColor: "text-muted-foreground",
      description: "Finished tasks",
    },
  ];

  // Calculate completion percentage
  const completionRate =
    stats?.total && stats?.total > 0
      ? Math.round((stats.completed / stats.total) * 100)
      : 0;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <motion.h1
            className="text-3xl md:text-4xl font-bold tracking-tight"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4 }}
          >
            Good{" "}
            {new Date().getHours() < 12
              ? "morning"
              : new Date().getHours() < 17
                ? "afternoon"
                : "evening"}
          </motion.h1>
          <motion.p
            className="text-muted-foreground mt-1"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: 0.05 }}
          >
            Here's an overview of your productivity
          </motion.p>
        </div>
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

      {/* Progress Banner */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.1 }}
      >
        <Card className="overflow-hidden border-accent/20 shadow-soft">
          <div className="relative">
            <div className="absolute inset-0 bg-gradient-to-r from-accent/10 via-accent/5 to-transparent" />
            <CardContent className="relative py-6">
              <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div className="flex items-center gap-4">
                  <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-accent to-accent/70 flex items-center justify-center shadow-soft">
                    <TrendingUp className="w-7 h-7 text-accent-foreground" />
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground font-medium">
                      Completion Rate
                    </p>
                    {statsLoading ? (
                      <Skeleton className="h-9 w-20" />
                    ) : (
                      <p className="text-3xl font-bold">{completionRate}%</p>
                    )}
                  </div>
                </div>
                <div className="flex-1 max-w-md">
                  {statsLoading ? (
                    <>
                      <Skeleton className="h-3 w-full rounded-full" />
                      <Skeleton className="mt-2 h-3 w-32" />
                    </>
                  ) : (
                    <>
                      <div className="h-3 rounded-full bg-muted overflow-hidden">
                        <motion.div
                          className="h-full rounded-full bg-gradient-to-r from-accent to-accent/80"
                          initial={{ width: 0 }}
                          animate={{ width: `${completionRate}%` }}
                          transition={{
                            duration: 0.8,
                            delay: 0.3,
                            ease: "easeOut",
                          }}
                        />
                      </div>
                      <p className="text-xs text-muted-foreground mt-2">
                        {stats?.completed || 0} of {stats?.total || 0} tasks
                        completed
                      </p>
                    </>
                  )}
                </div>
              </div>
            </CardContent>
          </div>
        </Card>
      </motion.div>

      {/* Stats Cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat, index) => {
          const Icon = stat.icon;
          return (
            <motion.div
              key={stat.title}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, delay: 0.15 + index * 0.05 }}
            >
              <Card className="hover-lift transition-all duration-300 hover:border-accent/30 shadow-soft">
                <CardContent className="pt-6">
                  <div className="flex items-start justify-between">
                    <div className="space-y-2">
                      <p className="text-sm text-muted-foreground font-medium">
                        {stat.title}
                      </p>
                      {statsLoading ? (
                        <Skeleton className="h-9 w-14" />
                      ) : (
                        <p className="text-3xl font-bold tracking-tight">
                          {stat.value}
                        </p>
                      )}
                      <p className="text-xs text-muted-foreground">
                        {stat.description}
                      </p>
                    </div>
                    <div
                      className={`w-11 h-11 rounded-xl bg-gradient-to-br ${stat.gradient} flex items-center justify-center`}
                    >
                      <Icon className={`w-5 h-5 ${stat.iconColor}`} />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          );
        })}
      </div>

      {/* Content Grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Recent Tasks */}
        <motion.div
          className="lg:col-span-2"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.35 }}
        >
          <Card className="shadow-soft">
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle className="text-lg font-semibold">
                Recent Tasks
              </CardTitle>
              <Link to="/todos" search={defaultTodoRouteSearch}>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-muted-foreground hover:text-foreground group"
                >
                  View all
                  <ArrowRight className="w-4 h-4 ml-1 transition-transform group-hover:translate-x-0.5" />
                </Button>
              </Link>
            </CardHeader>
            <CardContent className="space-y-3">
              {todosLoading ? (
                Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="space-y-2 p-4 rounded-xl bg-muted/30">
                    <Skeleton className="h-5 w-3/4" />
                    <Skeleton className="h-4 w-1/2" />
                  </div>
                ))
              ) : recentTodos?.data?.length ? (
                <div className="space-y-3">
                  {recentTodos.data.map((todo) => (
                    <TodoCard key={todo.id} todo={todo} compact />
                  ))}
                </div>
              ) : (
                <div className="py-12 text-center">
                  <div className="w-14 h-14 mx-auto rounded-2xl bg-muted/50 flex items-center justify-center mb-4">
                    <Sparkles className="w-7 h-7 text-muted-foreground" />
                  </div>
                  <p className="text-muted-foreground font-medium">
                    No tasks yet
                  </p>
                  <p className="text-sm text-muted-foreground/70">
                    Create your first task to get started
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </motion.div>

        {/* Categories */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.4 }}
        >
          <Card className="shadow-soft">
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle className="text-lg font-semibold">
                Categories
              </CardTitle>
              <Link to="/categories" search={defaultCategoryRouteSearch}>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-muted-foreground hover:text-foreground group"
                >
                  Manage
                  <ArrowRight className="w-4 h-4 ml-1 transition-transform group-hover:translate-x-0.5" />
                </Button>
              </Link>
            </CardHeader>
            <CardContent>
              {categoriesLoading ? (
                <div className="space-y-3">
                  {Array.from({ length: 4 }).map((_, i) => (
                    <Skeleton key={i} className="h-12 w-full rounded-xl" />
                  ))}
                </div>
              ) : categories?.data?.length ? (
                <div className="space-y-2">
                  {categories.data.slice(0, 6).map((category) => (
                    <div
                      key={category.id}
                      className="flex items-center justify-between p-3 rounded-xl bg-muted/30 hover:bg-muted/50 transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        <div
                          className="w-4 h-4 rounded-md shadow-sm"
                          style={{ backgroundColor: category.color }}
                        />
                        <span className="font-medium text-sm">
                          {category.name}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="py-8 text-center">
                  <p className="text-muted-foreground text-sm">
                    No categories yet
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </motion.div>
      </div>
    </div>
  );
}
